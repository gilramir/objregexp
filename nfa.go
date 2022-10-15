// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	"fmt"
	"strings"
)

// Made by starting with these sources, and modifying:
// https://medium.com/@phanindramoganti/regex-under-the-hood-implementing-a-simple-regex-compiler-in-go-ef2af5c6079
// https://github.com/phanix5/Simple-Regex-Complier-in-Go/blob/master/regex.go

// With more info at:
// https://www.oilshell.org/archive/Thompson-1968.pdf
// https://swtch.com/~rsc/regexp/regexp1.html

// The state needed to convert a stream of tokenT's into a Regexp with
// an nfa in it.
type nfaFactory[T comparable] struct {
	compiler *Compiler[T]

	// stack pointer, and stack, while buildind the NFA stack (regexp)
	stp   int
	stack []fragT[T]

	// how many registers are addressed by this regex
	numRegisters int
}

func newNfaFactory[T comparable](compiler *Compiler[T]) *nfaFactory[T] {
	return &nfaFactory[T]{
		compiler: compiler,
		stack:    make([]fragT[T], 0),
	}
}

/*
 Represents an NFA state plus zero or one or two arrows exiting.
 if c == nClass, it's something to match (Class), and only out is set
 if c == ntMatch, no arrows out; matching state.
 If c == ntSplit, unlabeled arrows to out and out1 (if != NULL).

*/
type nodeType int

const (
	ntClass nodeType = iota
	ntIdentity
	ntMeta
	ntMatch
	ntSplit
)

type metaType int

const (
	// An unitialized MetaT will be 0 but with no enum name
	mtAny metaType = iota + 1
)

// Important - once a regex is compiled, nothing in a nfaStateT can change.
// Otherwise, a single regex cannot be used in multiple concurrent goroutines
type nfaStateT[T comparable] struct {
	c nodeType

	// oClass is set if c is ntClass
	oClass *Class[T]

	// iObj and iName are set if c is ntIdentity
	iObj  T
	iName string

	// negation is valid for either oClass or iObj
	negation bool
	// meta is set if c is ntMeta
	meta metaType
	//	register int

	out, out1 *nfaStateT[T]

	// At this node, which registers start collecting
	// (an open paren) and which ones stop collection
	// (a closed paren)
	startsRegisters []int
	endsRegisters   []int
}

func stateListRepr[T comparable](stateList []*nfaStateT[T]) string {
	labels := make([]string, len(stateList))
	for i, ns := range stateList {
		labels[i] = ns.Repr0()
	}
	return fmt.Sprintf("[%s]", strings.Join(labels, ", "))
}

func (s *nfaStateT[T]) Repr0() string {
	var label string
	switch s.c {
	case ntMatch:
		label = "MATCH"
	case ntSplit:
		label = "SPLIT"
	case ntMeta:
		switch s.meta {
		case mtAny:
			label = "ANY"
		default:
			label = "MT?"
		}
	default:
		if s.oClass != nil {
			if s.negation {
				label = "!" + s.oClass.Name
			} else {
				label = s.oClass.Name
			}
		} else {
			label = "?"
		}
	}
	return fmt.Sprintf("<State %s sr:%v er:%v>", label,
		s.startsRegisters, s.endsRegisters)
}

func (s *nfaStateT[T]) Repr() string {
	saw := make(map[*nfaStateT[T]]bool)
	return s.ReprN(0, saw)
}

func (s *nfaStateT[T]) ReprN(n int, saw map[*nfaStateT[T]]bool) string {
	// Don't record MATCH as seen; we always want to display it
	if s.c != ntMatch {
		saw[s] = true
	}
	indent := strings.Repeat("  ", n)
	txt := fmt.Sprintf("%s%s", indent, s.Repr0())
	if s.out != nil && !saw[s.out] {
		txt += "\n" + s.out.ReprN(n+1, saw)
	}
	if s.out1 != nil && !saw[s.out1] {
		txt += "\n" + s.out1.ReprN(n+1, saw)
	}
	return txt
}

type fragT[T comparable] struct {
	start *nfaStateT[T]
	out   []**nfaStateT[T]
}

func (s *fragT[T]) Repr() string {
	out_repr := make([]string, len(s.out))
	for i, o := range s.out {
		if *o == nil {
			out_repr[i] = "<nil>"
		} else {
			out_repr[i] = (*o).Repr()
		}
	}
	return fmt.Sprintf("start: %s out: %v", s.start.Repr(), out_repr)
}

/* Patch the list of states at out to point to s. */
func (s *nfaFactory[T]) patch(out []**nfaStateT[T], ns *nfaStateT[T]) {
	for _, p := range out {
		*p = ns
	}
}

func (s *nfaFactory[T]) ensure_stack_space() {
	if len(s.stack) <= s.stp {
		extra := s.stp - len(s.stack) + 1
		s.stack = append(s.stack, make([]fragT[T], extra)...)
	}
}

func (s *nfaFactory[T]) token2nfa(token tokenT) error {

	dlog.Printf("token2nfa: %+v", token)
	switch token.ttype {

	case tClass: // could be a Class or an identity
		ctype, has := s.compiler.namespace[token.name]
		if !has {
			return fmt.Errorf("No such class or identity name '%s' at pos %d", token.name, token.pos)
		}
		var ns nfaStateT[T]
		switch ctype {
		case ccClass:
			class, has := s.compiler.classMap[token.name]
			if !has {
				panic(fmt.Sprintf("Should have found class '%s' at pos %d", token.name, token.pos))
			}
			ns = nfaStateT[T]{c: ntClass, oClass: class, negation: token.negation,
				out: nil, out1: nil}
		case ccIdentity:
			obj, has := s.compiler.identityObj[token.name]
			if !has {
				panic(fmt.Sprintf("Should have found identity '%s' at pos %d", token.name, token.pos))
			}
			ns = nfaStateT[T]{c: ntIdentity, iObj: obj, iName: token.name, negation: token.negation,
				out: nil, out1: nil}
		}
		s.stack[s.stp] = fragT[T]{&ns, []**nfaStateT[T]{&ns.out}}
		s.stp++
		s.ensure_stack_space()

	case tConcat:
		s.stp--
		e2 := s.stack[s.stp]
		s.stp--
		e1 := s.stack[s.stp]
		// concatenate
		s.patch(e1.out, e2.start)
		s.stack[s.stp] = fragT[T]{e1.start, e2.out}
		s.stp++
		// No need to call ensure_stack_space here; we popped 2
		// and added 1

	case tAlternate: // |
		s.stp--
		e2 := s.stack[s.stp]
		s.stp--
		e1 := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e1.start, out1: e2.start}
		s.stack[s.stp] = fragT[T]{&ns, append(e1.out, e2.out...)}
		s.stp++
		// No need to call ensure_stack_space here; we popped 2
		// and added 1

	case tGlobQuestion: // 0 or 1
		s.stp--
		e := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e.start}
		s.stack[s.stp] = fragT[T]{&ns, append(e.out, &ns.out1)}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tGlobStar: // 0 or more
		s.stp--
		e := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e.start}
		s.patch(e.out, &ns)
		s.stack[s.stp] = fragT[T]{&ns, []**nfaStateT[T]{&ns.out1}}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tGlobPlus: // 1 or more
		s.stp--
		e := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e.start}
		s.patch(e.out, &ns)
		s.stack[s.stp] = fragT[T]{e.start, []**nfaStateT[T]{&ns.out1}}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tAny:
		ns := nfaStateT[T]{c: ntMeta, meta: mtAny, out: nil, out1: nil}
		s.stack[s.stp] = fragT[T]{&ns, []**nfaStateT[T]{&ns.out}}
		s.stp++
		s.ensure_stack_space()

	case tStartRegister:
		ns := s.stack[s.stp-1].start
		ns.startsRegisters = append(ns.startsRegisters, token.int1)
		if token.int1 > s.numRegisters {
			s.numRegisters = token.int1
		}

	case tEndRegister:
		ns := s.stack[s.stp-1].start
		if ns.c == ntSplit {
			ns.out.endsRegisters = append(ns.out.endsRegisters, token.int1)
			ns.out1.endsRegisters = append(ns.out1.endsRegisters, token.int1)
		} else {
			ns.endsRegisters = append(ns.endsRegisters, token.int1)
		}

	default:
		e := fmt.Sprintf("token2nfa: %s not yet handled\n", string(token.ttype))
		panic(e)
	}
	return nil
}

func (s *nfaFactory[T]) compile(text string) (*Regexp[T], error) {

	tokens, err := parseRegex(text)
	if err != nil {
		return nil, fmt.Errorf("Parsing objregexp: %w", err)
	}

	for _, token := range tokens {
		dlog.Printf("token: %+v", token)
	}

	// stp is where a new item will be placed in the stack.
	// nfastack must always have allocated space for an item at index 'stp'
	s.stp = 0
	s.stack = make([]fragT[T], 1)

	for _, token := range tokens {
		err = s.token2nfa(token)
		if err != nil {
			return nil, err
		}
	}

	// After pushing and popping the stack, it should be empty
	s.stp--
	e := s.stack[s.stp]
	if s.stp != 0 {
		panic(fmt.Sprintf("compile failed: stp=%d e=%s", s.stp,
			e.Repr()))
	}
	re := &Regexp[T]{
		numRegisters: s.numRegisters,
	}
	re.matchstate.c = ntMatch

	s.patch(e.out, &re.matchstate)
	re.nfa = e.start

	// Dump it.
	dlog.Printf("nfa:\n%s", re.nfa.Repr())

	return re, nil
}

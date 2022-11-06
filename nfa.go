// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	"fmt"
	"io"
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

	// Stack pointer, and stack, while buildind the NFA stack (regexp)
	// Literals push these NFA fragments onto the stack.
	// Operators pop NFA fragments off ot the stack.
	stp   int
	stack []fragT[T]

	// how many registers are addressed by this regex
	numRegisters int

	// Maps regNames to regNums
	regNameMap map[string]int
}

func newNfaFactory[T comparable](compiler *Compiler[T]) *nfaFactory[T] {
	return &nfaFactory[T]{
		compiler:   compiler,
		stack:      make([]fragT[T], 0),
		regNameMap: make(map[string]int),
	}
}

func (s *nfaFactory[T]) stackRepr() string {
	repr := ""
	for i := 0; i < s.stp; i++ {
		repr += fmt.Sprintf("#%d: %s\n", i, s.stack[i].Repr())
	}
	return repr
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
	ntDynClass
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
// The NFA is built from a linked collection of nfsStateT objects
type nfaStateT[T comparable] struct {
	// What type (class) of nfa node is this?
	c nodeType

	// oClass is set if c is ntClass
	oClass *Class[T]

	// iObj is set if c is ntIdentity
	iObj T

	// if c is ntDynClass
	dynClass *dynClassT[T]

	// cName is set if c is ntClass or ntIdentity or ntDynClass
	cName string

	// negation is valid for either oClass or iObj
	negation bool

	// meta is set if c is ntMeta
	meta metaType

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

	case ntDynClass:
		label = s.cName

	case ntClass:
		if s.negation {
			label = "!" + s.oClass.Name
		} else {
			label = s.oClass.Name
		}
	case ntIdentity:
		if s.negation {
			label = "!" + s.cName
		} else {
			label = s.cName
		}
	}
	return fmt.Sprintf("<State %s sr:%v er:%v>", label,
		s.startsRegisters, s.endsRegisters)
}
func (s *nfaStateT[T]) Repr0Dot() string {
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

	case ntDynClass:
		label = s.cName
	case ntClass:
		if s.negation {
			label = "!" + s.oClass.Name
		} else {
			label = s.oClass.Name
		}
	case ntIdentity:
		if s.negation {
			label = "!" + s.cName
		} else {
			label = s.cName
		}
	}
	return fmt.Sprintf("State %s\\nsr:%v er:%v", label,
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

// Write the NFA to a dot file, for visualization with graphviz
func (s *nfaStateT[T]) writeDot(saw map[*nfaStateT[T]]bool, fh io.Writer) error {
	_, err := fmt.Fprintf(fh, "\tN%p [label=\"%s\"]\n", s, s.Repr0Dot())
	if err != nil {
		return err
	}
	saw[s] = true
	if s.out != nil {
		_, err = fmt.Fprintf(fh, "\tN%p -> N%p\n", s, s.out)
	}
	if err != nil {
		return err
	}
	if s.out1 != nil {
		_, err := fmt.Fprintf(fh, "\tN%p -> N%p\n", s, s.out1)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(fh)
	if err != nil {
		return err
	}
	if s.out != nil && !saw[s.out] {
		s.out.writeDot(saw, fh)
	}
	if s.out1 != nil && !saw[s.out1] {
		s.out1.writeDot(saw, fh)
	}
	return err
}

type fragT[T comparable] struct {
	// The start node of the fragment
	start *nfaStateT[T]
	// The out's that need connections
	out []**nfaStateT[T]

	// the endsRegisters to place on the outs after they are connected
	endsRegisters []int
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
	return fmt.Sprintf("start: %s out: %v er: %v", s.start.Repr(), out_repr, s.endsRegisters)
}

/* Patch the list of states at out to point to s. */
func (s *nfaFactory[T]) patch(f fragT[T], out []**nfaStateT[T], ns *nfaStateT[T]) {
	for _, p := range out {
		if len(f.endsRegisters) > 0 {
			ns.endsRegisters = append(ns.endsRegisters, f.endsRegisters...)
		}
		*p = ns
	}
}

func (s *nfaFactory[T]) ensure_stack_space() {
	if len(s.stack) <= s.stp {
		extra := s.stp - len(s.stack) + 1
		s.stack = append(s.stack, make([]fragT[T], extra)...)
	}
}

func (s *nfaFactory[T]) token2nfa(tnum int, token tokenT) error {

	dlog.Printf("---------------------")
	dlog.Printf("token2nfa: #%d %s", tnum, token.Repr())
	dlog.Printf("stack:\n%s", s.stackRepr())
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
			ns = nfaStateT[T]{c: ntClass, oClass: class, cName: token.name, negation: token.negation,
				out: nil, out1: nil}
		case ccIdentity:
			obj, has := s.compiler.identityObj[token.name]
			if !has {
				panic(fmt.Sprintf("Should have found identity '%s' at pos %d", token.name, token.pos))
			}
			ns = nfaStateT[T]{c: ntIdentity, iObj: obj, cName: token.name, negation: token.negation,
				out: nil, out1: nil}
		default:
			panic(fmt.Sprintf("Unexpected ctype %v for token %s", ctype, token.name))
		}
		s.stack[s.stp] = fragT[T]{&ns, []**nfaStateT[T]{&ns.out}, []int{}}
		s.stp++
		s.ensure_stack_space()

	case tDynClass:
		dynClass, err := newDynClassT[T](token.name, s.compiler)
		if err != nil {
			return fmt.Errorf("Parsing class string at pos %d: %s",
				token.pos, err)
		}

		ns := nfaStateT[T]{c: ntDynClass, dynClass: dynClass, cName: token.name,
			out: nil, out1: nil}
		s.stack[s.stp] = fragT[T]{&ns, []**nfaStateT[T]{&ns.out}, []int{}}
		s.stp++
		s.ensure_stack_space()

	case tConcat:
		s.stp--
		e2 := s.stack[s.stp]
		s.stp--
		e1 := s.stack[s.stp]
		// concatenate
		s.patch(e1, e1.out, e2.start)
		s.stack[s.stp] = fragT[T]{e1.start, e2.out, e2.endsRegisters}
		s.stp++
		// No need to call ensure_stack_space here; we popped 2
		// and added 1

	case tAlternate: // |
		s.stp--
		e2 := s.stack[s.stp]
		s.stp--
		e1 := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e1.start, out1: e2.start}
		s.stack[s.stp] = fragT[T]{&ns, append(e1.out, e2.out...), []int{}}
		s.stp++
		// No need to call ensure_stack_space here; we popped 2
		// and added 1

	case tGlobQuestion: // 0 or 1
		s.stp--
		e := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e.start}
		s.stack[s.stp] = fragT[T]{&ns, append(e.out, &ns.out1), []int{}}

		if len(e.endsRegisters) > 0 {
			s.stack[s.stp].endsRegisters = e.endsRegisters
			e.endsRegisters = []int{}
		}

		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tGlobStar: // 0 or more
		s.stp--
		e := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e.start}
		s.patch(e, e.out, &ns)
		s.stack[s.stp] = fragT[T]{&ns, []**nfaStateT[T]{&ns.out1}, []int{}}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tGlobPlus: // 1 or more
		s.stp--
		e := s.stack[s.stp]
		ns := nfaStateT[T]{c: ntSplit, out: e.start}
		s.patch(e, e.out, &ns)
		s.stack[s.stp] = fragT[T]{e.start, []**nfaStateT[T]{&ns.out1}, []int{}}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tAny:
		ns := nfaStateT[T]{c: ntMeta, meta: mtAny, out: nil, out1: nil}
		s.stack[s.stp] = fragT[T]{&ns, []**nfaStateT[T]{&ns.out}, []int{}}
		s.stp++
		s.ensure_stack_space()

	case tEndRegister:
		// An EndRegister cannot exist on an ntSplit node. It is pushed
		// down onto the final leavs of the ntSplit node/tree (ending up
		// on the er slices)
		dlog.Printf("tEndRegister reg#%d name %s", token.regNum, token.regName)

		ns := s.stack[s.stp-1].start
		ns.startsRegisters = append(ns.startsRegisters, token.regNum)
		s.stack[s.stp-1].endsRegisters = append(s.stack[s.stp-1].endsRegisters, token.regNum)

		if token.regNum > s.numRegisters {
			s.numRegisters = token.regNum
		}
		if token.regName != "" {
			if registeredNum, has := s.regNameMap[token.regName]; has {
				panic(fmt.Sprintf("reg#%d has name %s but it was seen with #%d", token.regNum, token.regName, registeredNum))
			}
			s.regNameMap[token.regName] = token.regNum
		}

	default:
		e := fmt.Sprintf("%s not handled\n", string(token.ttype))
		panic(e)
	}
	dlog.Printf("now the stack is:\n%s", s.stackRepr())
	return nil
}

func (s *nfaFactory[T]) compile(text string) (*Regexp[T], error) {

	tokens, err := parseRegex(text)
	if err != nil {
		return nil, fmt.Errorf("Parsing objregexp: %w", err)
	}

	printTokens(tokens)

	// stp is where a new item will be placed in the stack.
	// nfastack must always have allocated space for an item at index 'stp'
	s.stp = 0
	s.stack = make([]fragT[T], 1)

	for i, token := range tokens {
		err = s.token2nfa(i, token)
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
		regNameMap:   s.regNameMap,
	}
	re.matchstate.c = ntMatch

	s.patch(e, e.out, &re.matchstate)
	re.nfa = e.start

	// Dump it.
	dlog.Printf("nfa:\n%s", re.nfa.Repr())

	return re, nil
}

package objregexp

import (
	"fmt"
	"strings"
)

// https://medium.com/@phanindramoganti/regex-under-the-hood-implementing-a-simple-regex-compiler-in-go-ef2af5c6079
// https://github.com/phanix5/Simple-Regex-Complier-in-Go/blob/master/regex.go
// https://www.oilshell.org/archive/Thompson-1968.pdf
// https://swtch.com/~rsc/regexp/regexp1.html

type reFactory[T comparable] struct {
	compiler *Compiler[T]

	// stack pointer, and stack, while buildind the NFA stack (regexp)
	stp   int
	stack []Frag[T]

	// how many registers are addressed by this regex
	numRegisters int
}

func newReFactory[T comparable](compiler *Compiler[T]) *reFactory[T] {
	return &reFactory[T]{
		compiler: compiler,
		stack:    make([]Frag[T], 0),
	}
}

/*
 Represents an NFA state plus zero or one or two arrows exiting.
 if c == nClass, it's something to match (Class), and only out is set
 if c == NMatch, no arrows out; matching state.
 If c == NSplit, unlabeled arrows to out and out1 (if != NULL).

*/
type NodeT int

const (
	NClass NodeT = iota
	NMeta
	NMatch
	NSplit
)

type MetaT int

const (
	// An unitialized MetaT will be 0 but with no enum name
	MTAny MetaT = iota + 1
)

// Important - once a regex is compiled, nothing in State can change.
// Otherwise, a single regex cannot be used in multiple concurrent goroutines
// TODO - State -> stateT ; it doesn't need to be exported
type State[T comparable] struct {
	c NodeT

	// oClass or MetaT is set if c is 0
	oClass   *Class[T]
	negation bool
	meta     MetaT
	register int

	out, out1 *State[T]

	startsRegisters []int
	endsRegisters   []int
}

func stateListRepr[T comparable](stateList []*State[T]) string {
	labels := make([]string, len(stateList))
	for i, ns := range stateList {
		labels[i] = ns.Repr0()
	}
	return fmt.Sprintf("[%s]", strings.Join(labels, ", "))
}

func (s *State[T]) Repr0() string {
	var label string
	switch s.c {
	case NMatch:
		label = "MATCH"
	case NSplit:
		label = "SPLIT"
	case NMeta:
		switch s.meta {
		case MTAny:
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

func (s *State[T]) Repr() string {
	saw := make(map[*State[T]]bool)
	return s.ReprN(0, saw)
}

func (s *State[T]) ReprN(n int, saw map[*State[T]]bool) string {
	// Don't record MATCH as seen; we always want to display it
	if s.c != NMatch {
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

type Frag[T comparable] struct {
	start *State[T]
	out   []**State[T]
}

func (s *Frag[T]) Repr() string {
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
func (s *reFactory[T]) patch(out []**State[T], ns *State[T]) {
	for _, p := range out {
		*p = ns
	}
}

func (s *reFactory[T]) ensure_stack_space() {
	if len(s.stack) <= s.stp {
		extra := s.stp - len(s.stack) + 1
		s.stack = append(s.stack, make([]Frag[T], extra)...)
	}
}

func (s *reFactory[T]) token2nfa(token tokenT) error {

	fmt.Printf("token2nfa: %+v\n", token)
	switch token.ttype {

	case tClass:
		oclass, has := s.compiler.oclassMap[token.name]
		if !has {
			return fmt.Errorf("No such class name '%s' at pos %d", token.name, token.pos)
		}
		ns := State[T]{c: NClass, oClass: oclass, negation: token.negation,
			out: nil, out1: nil}
		s.stack[s.stp] = Frag[T]{&ns, []**State[T]{&ns.out}}
		s.stp++
		s.ensure_stack_space()

	case tConcat:
		s.stp--
		e2 := s.stack[s.stp]
		s.stp--
		e1 := s.stack[s.stp]
		// concatenate
		s.patch(e1.out, e2.start)
		s.stack[s.stp] = Frag[T]{e1.start, e2.out}
		s.stp++
		// No need to call ensure_stack_space here; we popped 2
		// and added 1

	case tAlternate: // |
		s.stp--
		e2 := s.stack[s.stp]
		s.stp--
		e1 := s.stack[s.stp]
		ns := State[T]{c: NSplit, out: e1.start, out1: e2.start}
		s.stack[s.stp] = Frag[T]{&ns, append(e1.out, e2.out...)}
		s.stp++
		// No need to call ensure_stack_space here; we popped 2
		// and added 1

	case tGlobQuestion: // 0 or 1
		s.stp--
		e := s.stack[s.stp]
		ns := State[T]{c: NSplit, out: e.start}
		s.stack[s.stp] = Frag[T]{&ns, append(e.out, &ns.out1)}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tGlobStar: // 0 or more
		s.stp--
		e := s.stack[s.stp]
		ns := State[T]{c: NSplit, out: e.start}
		s.patch(e.out, &ns)
		s.stack[s.stp] = Frag[T]{&ns, []**State[T]{&ns.out1}}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tGlobPlus: // 1 or more
		s.stp--
		e := s.stack[s.stp]
		ns := State[T]{c: NSplit, out: e.start}
		s.patch(e.out, &ns)
		s.stack[s.stp] = Frag[T]{e.start, []**State[T]{&ns.out1}}
		s.stp++
		// No need to call ensure_stack_space here; we popped 1
		// and added 1

	case tAny:
		ns := State[T]{c: NMeta, meta: MTAny, out: nil, out1: nil}
		s.stack[s.stp] = Frag[T]{&ns, []**State[T]{&ns.out}}
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
		if ns.c == NSplit {
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

func (s *reFactory[T]) compile(text string) (*Regexp[T], error) {

	tokens, err := parseRegex(text)
	if err != nil {
		return nil, fmt.Errorf("Parsing objregexp: %w", err)
	}

	for _, token := range tokens {
		fmt.Printf("token: %+v\n", token)
	}

	// stp is where a new item will be placed in the stack.
	// nfastack must always have allocated space for an item at index 'stp'
	s.stp = 0
	s.stack = make([]Frag[T], 1)

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
	re.matchstate.c = NMatch

	s.patch(e.out, &re.matchstate)
	re.nfa = e.start

	// Dump it.
	fmt.Printf("nfa:\n%s\n", re.nfa.Repr())

	return re, nil
}

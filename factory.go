package objregexp

import (
	"fmt"
	"strings"
)

// https://medium.com/@phanindramoganti/regex-under-the-hood-implementing-a-simple-regex-compiler-in-go-ef2af5c6079
// https://github.com/phanix5/Simple-Regex-Complier-in-Go/blob/master/regex.go
// https://www.oilshell.org/archive/Thompson-1968.pdf

type reFactory[T comparable] struct {
	compiler *Compiler[T]

	// stack pointer, and stack, while buildind the NFA stack (regexp)
	stp   int
	stack []Frag[T]
}

func newReFactory[T comparable](compiler *Compiler[T]) *reFactory[T] {
	return &reFactory[T]{
		compiler: compiler,
		stack:    make([]Frag[T], 0),
	}
}

/*
 Represents an NFA state plus zero or one or two arrows exiting.
 if c == 0, it's something to match (Class), and only out is set
 if c == NMatch, no arrows out; matching state.
 If c == NSplit, unlabeled arrows to out and out1 (if != NULL).

*/
const (
	NMatch = iota + 1
	NSplit
)

type State[T comparable] struct {
	c int

	// oClass is set if c is 0
	oClass *Class[T]

	out, out1 *State[T]
	lastlist  int
}

func (s *State[T]) Repr0() string {
	var label string
	if s.c == NMatch {
		label = "MATCH"
	} else if s.c == NSplit {
		label = "SPLIT"
	} else {
		if s.oClass != nil {
			label = s.oClass.Name
		} else {
			label = "?"
		}
	}
	return fmt.Sprintf("<State %s lastlist=%d>", label, s.lastlist)
}

func (s *State[T]) Repr() string {
	return s.ReprN(0)
}

func (s *State[T]) ReprN(n int) string {
	indent := strings.Repeat("  ", n)
	txt := fmt.Sprintf("%s%s", indent, s.Repr0())
	if s.out != nil {
		txt += "\n" + s.out.ReprN(n+1)
	}
	if s.out1 != nil {
		txt += "\n" + s.out1.ReprN(n+1)
	}
	return txt
}

// Won't need this
func (s *State[T]) RecursiveClearState() {
	s.lastlist = 0
	if s.out != nil {
		s.out.RecursiveClearState()
	}
	if s.out1 != nil {
		s.out1.RecursiveClearState()
	}
}

type Frag[T comparable] struct {
	start *State[T]
	out   []**State[T]
}

func (s *Frag[T]) Repr() string {
	out_repr := make([]string, len(s.out))
	for i, o := range s.out {
		out_repr[i] = (*o).Repr()
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
	default:
		e := fmt.Sprintf("token2nfa: %s not yet handled\n", string(token.ttype))
		panic(e)

	case tClass:
		oclass, has := s.compiler.oclassMap[token.name]
		if !has {
			return fmt.Errorf("No such class name '%s' at pos %d", token.name, token.pos)
		}
		ns := State[T]{oClass: oclass, out: nil, out1: nil}
		s.stack[s.stp] = Frag[T]{&ns, []**State[T]{&ns.out}}
		s.stp++
		s.ensure_stack_space()
	}
	return nil
}

/*
func (s *reFactory[T]) node2nfa(node *ParserTree[O]) {

	switch node.item.NodeType() {
	default:
		e := fmt.Sprintf("node2nfa: %s not yet handled\n", NodeTypeName(node.item.NodeType()))
		panic(e)

	case OPIOClass:
		ns := State[I, O]{io: node.item.(*IOClass[I, O]), out: nil, out1: nil}
		s.stack[s.stp] = Frag[I, O]{&ns, []**State[I, O]{&ns.out}}
		s.stp++
		s.ensure_stack_space()

	case OPAnd: // concatenate
		if len(node.children) < 2 {
			panic(fmt.Sprintf("%s has only %d children", node.Repr(), len(node.children)))
		}

		for i, ch := range node.children {
			s.node2nfa(ch)
			if i >= 1 {
				s.stp--
				e2 := s.stack[s.stp]
				s.stp--
				e1 := s.stack[s.stp]
				// concatenate
				s.patch(e1.out, e2.start)
				s.stack[s.stp] = Frag[I, O]{e1.start, e2.out}
				s.stp++
				// No need to call ensure_stack_space here; we popped 2
				// and added 1
			}
		}

	case OPOr: // alternate
		// Create entries in the stack for each child, and along the
		// way, pair them with SPLIT nodes. Each SPLIT node can only
		// have 2 outputs
		if len(node.children) < 2 {
			panic(fmt.Sprintf("%s has only %d children", node.Repr(), len(node.children)))
		}
		for i, ch := range node.children {
			s.node2nfa(ch)
			if i >= 1 {
				s.stp--
				e2 := s.stack[s.stp]
				s.stp--
				e1 := s.stack[s.stp]
				ns := State[I, O]{c: NSplit, out: e1.start, out1: e2.start}
				s.stack[s.stp] = Frag[I, O]{&ns, append(e1.out, e2.out...)}
				s.stp++
				// No need to call ensure_stack_space here; we popped 2
				// and added 1
			}
		}
	}
}
*/

/*
func (s *reFactory[T]) Parse(input []T) ([]O, error) {
	fmt.Printf("Parsing %v\n", input)
	o := make([]O, 0)

	// Reset the state after any previous parse
	s.nfa.RecursiveClearState()
	s.listid = 0

	m := s.match(s.nfa, input)
	fmt.Printf("Matches: %v\n", m)
	return o, nil
}
*/

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
	re := &Regexp[T]{}

	s.patch(e.out, &re.matchstate)
	re.nfa = e.start

	// Dump it.
	fmt.Printf("nfa:\n%s\n", re.nfa.Repr())

	return re, nil
}

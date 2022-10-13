package objregexp

import (
	"fmt"
	"strings"
)

type Regexp[T comparable] struct {
	// the root node of the stack; where the parse begins
	nfa *State[T]

	// singleton matching state used to denote all end states
	matchstate State[T]
	listid     int
}

type Range struct {
	Start int
	End   int
}

type Match[T comparable] struct {
	Success bool
	Range   Range
	Group   map[string]Range
}

func (s *Regexp[T]) FullMatch(input []T) Match[T] {
	s.listid = 0
	s.nfa.RecursiveClearState()

	vars := make(map[string]Range)
	matched := s.match(s.nfa, input)
	if matched {
		return Match[T]{
			Success: true,
			Group:   vars,
			Range: Range{
				Start: 0,
				End:   len(input),
			},
		}
	} else {
		return Match[T]{
			Success: false,
			Group:   make(map[string]Range),
		}
	}
}
func (s *Regexp[T]) Match(input []T) Match[T] {
	s.listid = 0
	s.nfa.RecursiveClearState()

	vars := make(map[string]Range)
	matched, n := s.match2(s.nfa, input, 0)
	if matched {
		return Match[T]{
			Success: true,
			Range: Range{
				Start: 0,
				End:   n,
			},
			Group: vars,
		}
	} else {
		return Match[T]{
			Success: false,
			Group:   make(map[string]Range),
		}
	}
}

func (s *Regexp[T]) stateListRepr(stateList []*State[T]) string {
	labels := make([]string, len(stateList))
	for i, ns := range stateList {
		labels[i] = ns.Repr0()
	}
	return fmt.Sprintf("[%s]", strings.Join(labels, ", "))
}

func (s *Regexp[T]) match2(start *State[T], input []T, from int) (bool, int) {

	var clist, nlist []*State[T]
	s.listid++
	clist = s.addstate(clist, start)

	for i := from; i < len(input); i++ {
		ch := input[i]
		fmt.Printf("Input #%d: %v: clist=%s nlist=%s\n",
			i, ch, s.stateListRepr(clist),
			s.stateListRepr(nlist))
		nlist = s.step(clist, ch, nlist)
		clist, nlist = nlist, clist
		fmt.Printf("\tnew clist: %s\n", s.stateListRepr(clist))

		if s.ismatch(clist) {
			return true, i - from + 1
		}
	}
	matched := s.ismatch(clist)
	if matched {
		return true, len(input)
	} else {
		return false, 0
	}
}

func (s *Regexp[T]) match(start *State[T], input []T) bool {

	var clist, nlist []*State[T]
	s.listid++
	clist = s.addstate(clist, start)

	for i, ch := range input {
		fmt.Printf("Input #%d: %v: clist=%s nlist=%s\n",
			i, ch, s.stateListRepr(clist),
			s.stateListRepr(nlist))
		nlist = s.step(clist, ch, nlist)
		clist, nlist = nlist, clist
	}
	return s.ismatch(clist)
}

// Add s to l, following unlabeled arrows.
func (s *Regexp[T]) addstate(l []*State[T], ns *State[T]) []*State[T] {
	if ns == nil || ns.lastlist == s.listid {
		return l
	}
	ns.lastlist = s.listid
	if ns.c == NSplit {
		l = s.addstate(l, ns.out)
		l = s.addstate(l, ns.out1)
	}
	l = append(l, ns)
	return l
}

/*
 * Step the NFA from the states in clist
 * past the character ch,
 * to create next NFA state set nlist.
 */
func (s *Regexp[T]) step(clist []*State[T], ch T, nlist []*State[T]) []*State[T] {
	s.listid++
	nlist = nlist[:0]
	//fmt.Printf("step: nlist=%s\n", s.stateListRepr(nlist))
	for _, ns := range clist {
		if ns.oClass == nil {
			continue
		}
		m := ns.oClass.Matches(ch)
		// Are we testing for non-memberhood?
		if ns.negation {
			m = !m
		}
		//fmt.Printf("step: clist %d %s => %v\n", i, ns.Repr0(), m)
		// TODO - how to record the output?
		if m {
			nlist = s.addstate(nlist, ns.out)
		}
	}
	return nlist
}

// Check whether state list contains a match.
func (s *Regexp[T]) ismatch(l []*State[T]) bool {
	for _, ns := range l {
		if ns == &s.matchstate {
			return true
		}
	}
	return false
}

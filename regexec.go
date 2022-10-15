// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	"fmt"
	"strings"
)

// Keeps track of state needed/modified during the exection of a regex
type executorT[T comparable] struct {
	regex *Regexp[T]

	// for stepping
	listid int

	// Every State should have only 1 copy as an exState, so we need this
	// cache
	stCache map[*nfaStateT[T]]*exStateT[T]

	matchstate *exStateT[T]
}

// Initialize an executorT from a Regexp
func (s *executorT[T]) Initialize(regex *Regexp[T]) {
	s.regex = regex
	s.listid = 0
	s.stCache = make(map[*nfaStateT[T]]*exStateT[T])
	s.matchstate = s.exState(&regex.matchstate)
}

// This mirrors a state object, but it's modifiable so that the same
// Regexp object (State) can be used in multiple concrent goroutines
type exStateT[T comparable] struct {
	st *nfaStateT[T]

	// for the nfa
	lastlist int

	// We keep a singly-linked-list backwards from tail to root
	prev *exStateT[T]

	out, out1 *exStateT[T]

	registers *registersT
}

type registersT struct {
	active []bool
	ranges []Range
}

func (s *registersT) Copy() *registersT {
	r := &registersT{
		active: make([]bool, len(s.active)),
		ranges: make([]Range, len(s.ranges)),
	}
	copy(r.active, s.active)
	copy(r.ranges, s.ranges)
	return r
}

func (s *executorT[T]) exStateRecursive(state *nfaStateT[T]) *exStateT[T] {

	return s.exState(state)
}

func (s *executorT[T]) exState(state *nfaStateT[T]) *exStateT[T] {
	if state == nil {
		return nil
	}
	if xs, has := s.stCache[state]; has {
		return xs
	}
	xs := &exStateT[T]{
		st:        state,
		registers: s.newRegisters(),
	}
	s.stCache[state] = xs

	// TODO - make this iterative instead of recursive
	xs.out = s.exState(state.out)
	xs.out1 = s.exState(state.out1)

	return xs
}
func (s *executorT[T]) newRegisters() *registersT {
	r := &registersT{
		active: make([]bool, s.regex.numRegisters),
		ranges: make([]Range, s.regex.numRegisters),
	}
	for i := 0; i < s.regex.numRegisters; i++ {
		r.ranges[i].Start = -1
		r.ranges[i].End = -1
	}
	return r
}

// A few data dumpers

func (s *exStateT[T]) Repr0() string {
	return fmt.Sprintf("<exStateT %s reg:+%v>", s.st.Repr0(), s.registers.ranges)
}

func exStateListRepr[T comparable](exStateList []*exStateT[T]) string {
	labels := make([]string, len(exStateList))
	for i, x := range exStateList {
		labels[i] = x.Repr0()
	}
	return fmt.Sprintf("[%s]", strings.Join(labels, ", "))
}

type hitT[T comparable] struct {
	x      *exStateT[T]
	length int
}

// Walk the list of states, which can expand into branches.
// If full is true, wait until the end of the string to check for a final match
// If full is false, return true as soon as a match is found
func (s *executorT[T]) match(start *nfaStateT[T], input []T, from int, full bool) (bool, int, *exStateT[T]) {

	xstart := s.exStateRecursive(start)

	// Any starting registers?
	for _, rn := range s.regex.startRegisters {
		xstart.registers.ranges[rn-1].Start = 0
	}

	var clist, nlist []*exStateT[T]
	s.listid++
	clist = s.addstate(clist, xstart)

	// Keep track of matches because we want to be a little greedy
	// and not return too early
	var hit hitT[T]

	for i := from; i < len(input); i++ {
		ch := input[i]
		dlog.Printf("Input #%d: %v: clist=%s nlist=%s",
			i, ch, exStateListRepr(clist),
			exStateListRepr(nlist))
		nlist = s.step(i, clist, ch, nlist)
		clist, nlist = nlist, clist
		dlog.Printf("\tnew clist: %s", exStateListRepr(clist))

		if !full {
			if matched, xns := s.ismatch(clist); matched {
				hit = hitT[T]{x: xns, length: i - from + 1}
				dlog.Printf("MATCHED and stored hit %+v", hit)
				// keep going
			} else {
				dlog.Printf("NO MATCH and stored hit %+v\n", hit)
				if hit.x != nil {
					// now we can return
					return true, hit.length, hit.x
				}
			}
		}
	}
	if full {
		// If looking for a full match, did we match at the end?
		if matched, xns := s.ismatch(clist); matched {
			return true, len(input) - from, xns
		} else {
			return false, 0, nil
		}
	} else {
		dlog.Printf("FINISHED and stored hit %+v", hit)
		if hit.x != nil {
			// now we can return
			return true, hit.length, hit.x
		}
		// If we weren't looking for a full match, and didn't already find
		// it, then we didn't find it.
		return false, 0, nil
	}
}

// Add s to l, following unlabeled arrows.
func (s *executorT[T]) addstate(l []*exStateT[T], ns *exStateT[T]) []*exStateT[T] {
	if ns == nil || ns.lastlist == s.listid {
		return l
	}
	ns.lastlist = s.listid
	if ns.st.c == ntSplit {
		fmt.Sprintf("Split %s -> %s and %s\n", ns.Repr0(), ns.out.Repr0(), ns.out1.Repr0())
		ns.out.registers = ns.registers.Copy()
		ns.out1.registers = ns.registers.Copy()
		l = s.addstate(l, ns.out)
		l = s.addstate(l, ns.out1)
	}
	// TODO - I'm not sure why we append ns here if ns == NSPlit. Is it
	// necessary?
	l = append(l, ns)
	return l
}

/*
 * Step the NFA from the states in clist
 * past the character ch,
 * to create next NFA state set nlist.
 */
func (s *executorT[T]) step(pos int, clist []*exStateT[T], ch T, nlist []*exStateT[T]) []*exStateT[T] {
	s.listid++
	nlist = nlist[:0]
	dlog.Printf("step @ %d: clist has %d : %s", pos, len(clist), exStateListRepr(clist))
	for ci, xns := range clist {
		dlog.Printf("looking at clist #%d: %s", ci, xns.Repr0())
		ns := xns.st
		var matches bool
		switch ns.c {
		default:
			dlog.Printf("<skipping>")
			continue
		case ntClass:
			matches = ns.oClass.Matches(ch)
			dlog.Printf("Matches class %s: %v", ns.oClass.Name, matches)
			// Are we testing for non-memberhood?
			if ns.negation {
				matches = !matches
				dlog.Printf("Negation -> %v", matches)
			}
		case ntIdentity:
			matches = ns.iObj == ch
			dlog.Printf("Identiy %s: %v", ns.iName, matches)
			// Are we testing for non-memberhood?
			if ns.negation {
				matches = !matches
				dlog.Printf("Negation -> %v", matches)
			}
		case ntMeta:
			switch ns.meta {
			case mtAny:
				matches = true

			default:
				panic(fmt.Sprintf("Unexpected meta '%v'", ns.meta))
			}
		}
		if matches {

			for _, rn := range ns.startsRegisters {
				// The matching character would be 1 after this pos
				xns.registers.ranges[rn-1].Start = pos + 1
			}
			for _, rn := range ns.endsRegisters {
				// The end paren is this pos, but we record pos+1
				// to be more like Go slices
				xns.registers.ranges[rn-1].End = pos + 1
			}

			dlog.Printf("match; calling addstate()")
			// Copy the previous registers
			xns.out.registers = xns.registers.Copy()
			dlog.Printf("new registers: %+v", xns.out.registers.ranges)
			nlist = s.addstate(nlist, xns.out)
		}
	}
	return nlist
}

// Check whether state list contains a match.
func (s *executorT[T]) ismatch(l []*exStateT[T]) (bool, *exStateT[T]) {
	for _, ns := range l {
		if ns == s.matchstate {
			dlog.Printf("matched; registers: %+v", ns.registers.ranges)
			return true, ns
		}
	}
	return false, nil
}

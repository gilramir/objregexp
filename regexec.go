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

	out, out1 *exStateT[T]
}

type registersT struct {
	ranges []Range
}

func (s *executorT[T]) newRegisters() *registersT {
	r := &registersT{
		ranges: make([]Range, s.regex.numRegisters),
	}
	for i := 0; i < s.regex.numRegisters; i++ {
		r.ranges[i].Start = -1
		r.ranges[i].End = -1
	}
	return r
}

func (s *registersT) Copy() *registersT {
	r := &registersT{
		ranges: make([]Range, len(s.ranges)),
	}
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
		st: state,
	}
	s.stCache[state] = xs

	// TODO - make this iterative instead of recursive
	xs.out = s.exState(state.out)
	xs.out1 = s.exState(state.out1)

	return xs
}

// A few data dumpers

func (s *exStateT[T]) Repr0() string {
	return fmt.Sprintf("<exStateT %s>", s.st.Repr0())
}

func (s *exStateT[T]) Repr() string {
	saw := make(map[*exStateT[T]]bool)
	return s.ReprN(0, saw)
}

func (s *exStateT[T]) ReprN(n int, saw map[*exStateT[T]]bool) string {
	// Don't record MATCH as seen; we always want to display it
	if s.st.c != ntMatch {
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

func exStateListRepr[T comparable](exStateList []*exStateT[T]) string {
	labels := make([]string, len(exStateList))
	for i, x := range exStateList {
		labels[i] = x.Repr0()
	}
	return fmt.Sprintf("[%s]", strings.Join(labels, ", "))
}

type hitT[T comparable] struct {
	x      *nfaRegStateT[T]
	length int
}

// Walk the list of states, which can expand into branches.
// If full is true, wait until the end of the string to check for a final match
// If full is false, return true as soon as a match is found
func (s *executorT[T]) match(start *nfaStateT[T], input []T, from int, full bool) (bool, int, *nfaRegStateT[T]) {
	ok, count, xns := s._match(start, input, from, full)
	if ok {
		// It's possible for us to have -1's on one side (start/end)
		// of the range of a register. That's ok, but we need to clean
		// them up.
		for i, reg := range xns.registers.ranges {
			dlog.Printf("checking reg %d: %+v\n", i+1, reg)
			// TODO - is this situation still possible?
			if reg.End == -1 {
				xns.registers.ranges[i].Start = -1
			} else if reg.Start == -1 {
				xns.registers.ranges[i].End = -1
			} else if reg.Start == reg.End {
				// If they start and stop on the same object,
				// the range doesn't exist
				xns.registers.ranges[i].Start = -1
				xns.registers.ranges[i].End = -1
			}

		}
	}
	return ok, count, xns
}

type nfaRegStateT[T comparable] struct {
	root      *exStateT[T]
	registers *registersT
}

func (s *executorT[T]) _match(start *nfaStateT[T], input []T, from int, full bool) (bool, int, *nfaRegStateT[T]) {

	xstart := s.exStateRecursive(start)

	var clist, nlist []*nfaRegStateT[T]
	s.listid++
	// first pos is = -1, in case SPLIT is the 1st nfa node
	dlog.Printf("calling addstate on root nfa")
	clist = s.addstate(-1, clist, &nfaRegStateT[T]{xstart, s.newRegisters()})

	// Keep track of matches because we want to be a little greedy
	// and not return too early
	var hit hitT[T]

	for i := from; i < len(input); i++ {
		ch := input[i]
		dlog.Printf("=========================================")
		dlog.Printf("Input #%d: %v", i, ch)
		dlog.Printf("clist has %d items:", len(clist))
		for cxi, cxsr := range clist {
			cxs := cxsr.root
			dlog.Printf("clist item #%d regs:%v\n%s", cxi, cxsr.registers.ranges, cxs.Repr())
		}
		nlist = s.step(i, clist, ch, nlist)
		clist, nlist = nlist, clist
		//		dlog.Printf("\tnew clist: %s", exStateListRepr(clist))

		if !full {
			if matched, xns := s.ismatch(i+1, clist); matched {
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
		if matched, xns := s.ismatch(len(input)-from, clist); matched {
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

// Add ns to l, following unlabeled arrows.
func (s *executorT[T]) addstate(pos int, l []*nfaRegStateT[T], nsx *nfaRegStateT[T]) []*nfaRegStateT[T] {
	ns := nsx.root
	regs := nsx.registers
	if ns == nil || ns.lastlist == s.listid {
		return l
	}
	ns.lastlist = s.listid
	dlog.Printf("in addstate at pos %d, ns=%s l list has %d items:", pos, ns.Repr0(), len(l))
	for li, lnx := range l {
		lx := lnx.root
		dlog.Printf("state #%d: reg:%v\n%s", li, lnx.registers.ranges, lx.Repr())
	}
	if ns.st.c == ntSplit {
		for _, rn := range ns.st.startsRegisters {
			dlog.Printf("addstate setting start reg #%d = pos %d", rn, pos)
			// The matching character starts this register,
			// unless it was already seen (due to "*" glob)
			//			if regs.ranges[rn-1].Start == -1 {
			regs.ranges[rn-1].Start = pos + 1
			//			}
		}
		for _, rn := range ns.st.endsRegisters {
			dlog.Printf("addstate setting end reg #%d = pos %d", rn, pos)
			// The end paren is this pos, but we record pos+1
			// to be more like Go slices
			// check that start was seen first; it won't be
			// in "*" glob
			if regs.ranges[rn-1].Start != -1 {
				regs.ranges[rn-1].End = pos + 1
			}
		}
		l = s.addstate(pos, l, &nfaRegStateT[T]{ns.out, nsx.registers.Copy()})
		l = s.addstate(pos, l, &nfaRegStateT[T]{ns.out1, nsx.registers.Copy()})
	}
	// TODO - I'm not sure why we append ns here if ns == NSPlit. Is it
	// necessary?
	l = append(l, nsx)
	return l
}

/*
 * Step the NFA from the states in clist
 * past the character ch,
 * to create next NFA state set nlist.
 */
func (s *executorT[T]) step(pos int, clist []*nfaRegStateT[T], ch T, nlist []*nfaRegStateT[T]) []*nfaRegStateT[T] {
	s.listid++
	nlist = nlist[:0]
	dlog.Printf("step @ %d: clist has %d items", pos, len(clist))
	for li, lnx := range clist {
		lx := lnx.root
		dlog.Printf("clist #%d: reg:%v\n%s", li, lnx.registers.ranges, lx.Repr())
	}
	for ci, xnsr := range clist {
		xns := xnsr.root
		regs := xnsr.registers
		dlog.Printf("looking at clist #%d: %s", ci, xns.Repr0())
		ns := xns.st
		var matches bool
		switch ns.c {
		default:
			switch ns.c {
			case ntMatch:
				dlog.Printf("<skipping ntMatch>")
			case ntSplit:
				dlog.Printf("<skipping ntPslit>")
			default:
				dlog.Printf("<skipping ? %d>", ns.c)
			}
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
			dlog.Printf("Identity %s: %v", ns.cName, matches)
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
		case ntDynClass:
			matches = ns.dynClass.Matches(ch)
			dlog.Printf("Matches dynClass %s: %v", ns.cName, matches)
		}

		if matches {
			dlog.Printf("matched : %s", xns.Repr0())
			for _, rn := range ns.startsRegisters {
				// The matching character starts this register,
				// unless it was already set (due to "*" glob)
				if regs.ranges[rn-1].Start == -1 {
					regs.ranges[rn-1].Start = pos
				}
			}
			for _, rn := range ns.endsRegisters {
				// The end paren is this pos, but we record pos+1
				// to be more like Go slices
				regs.ranges[rn-1].End = pos
			}
			dlog.Printf("This nfa's registers: %v\n", regs.ranges)
			nlist = s.addstate(pos, nlist, &nfaRegStateT[T]{xns.out, regs.Copy()})
		}
	}
	return nlist
}

// Check whether state list contains a match.
func (s *executorT[T]) ismatch(pos int, l []*nfaRegStateT[T]) (bool, *nfaRegStateT[T]) {
	for _, nsr := range l {
		ns := nsr.root
		regs := nsr.registers
		if ns == s.matchstate {
			for _, rn := range ns.st.endsRegisters {
				dlog.Printf("ismatch setting end reg #%d = pos %d", rn, pos)
				regs.ranges[rn-1].End = pos
			}
			dlog.Printf("matched; registers: %+v", nsr.registers.ranges)
			return true, nsr
		}
	}
	return false, nil
}

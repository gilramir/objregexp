// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	"fmt"
	"os"
)

// The compiled regex.
type Regexp[T comparable] struct {
	// the root node of the stack; where the parse begins
	nfa *nfaStateT[T]

	// singleton matching state used to denote all end states
	matchstate nfaStateT[T]

	// How many registers can be saved to by this regex
	numRegisters int

	// Maps regNames to regNums
	regNameMap map[string]int

	// Does the regexp need to start with some specific object?
	// This helps the Search() method.
	initialObj *nfaStateT[T]
}

// Does this regex only match at the beginning of the input?
// That is, must ^ be satisified always for this regexp?
func (s *Regexp[T]) onlyMatchesAtBeginning() bool {
	//	log.Printf("nfa: %s", s.nfa.Repr())
	return s.onlyMatchesAtBeginningRecursive(s.nfa)
}

func (s *Regexp[T]) onlyMatchesAtBeginningRecursive(nfa *nfaStateT[T]) bool {
	switch nfa.c {
	case ntMeta:
		return nfa.meta == mtAssertBegin
	case ntSplit:
		return s.onlyMatchesAtBeginningRecursive(nfa.out) &&
			s.onlyMatchesAtBeginningRecursive(nfa.out1)
	default:
		return false
	}
}

// Does this regex only match at the beginning of the input?
// If an nfaStateT is returned, it will be an ntClass,
// ntIdentity, or ntDynClass. Otherwise, nil is returned.
func (s *Regexp[T]) mustStartWith() *nfaStateT[T] {
	switch s.nfa.c {
	case ntClass:
		return &nfaStateT[T]{
			c:      s.nfa.c,
			oClass: s.nfa.oClass,
		}
	case ntIdentity:
		return &nfaStateT[T]{
			c:    s.nfa.c,
			iObj: s.nfa.iObj,
		}
	case ntDynClass:
		return &nfaStateT[T]{
			c:        s.nfa.c,
			dynClass: s.nfa.dynClass,
		}
	default:
		return nil
	}
}

// Write the NFA to a dot file, for visualization with graphviz
func (s *Regexp[T]) WriteDot(filename string) error {
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = fmt.Fprintf(fh, "digraph {\n")
	if err != nil {
		return err
	}

	root := fmt.Sprintf("Root #r:%d", s.numRegisters)

	_, err = fmt.Fprintf(fh, "\troot [label=\"%s\"]\n", root)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(fh, "\troot -> N%p\n", s.nfa)
	if err != nil {
		return err
	}

	saw := make(map[*nfaStateT[T]]bool)
	err = s.nfa.writeDot(saw, fh)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fh, "}\n")

	return err
}

// This is used to record the span of objects, relative to the
// slice of input objects that was given. Start and End follow
// Golang slice semantics. The positions are 0-indexed.
// A slice of 1 item at the beginning of the slice has
// Start = 0 and End = 1.
type Range struct {
	Start int
	End   int
}

func (s Range) Length() int {
	return s.End - s.Start
}

// Is there anything in the range?
func (s Range) Empty() bool {
	return s.End-s.Start == 0
}

// Returned by a Regexp matching-related function.
// Even if the Regexp doesn't match, a Match object is returned,
// but Success is false.
type Match struct {
	// Did the Regexp find something?
	Success bool

	// The range for the entire sub-string that matched
	Range Range

	registers  []Range
	regNameMap map[string]int
}

// How many objects did this regexp match? This is for
// the entire regexp, not any capture group within it.
func (s Match) Length() int {
	return s.Range.Length()
}

// Does the numbered capture group have any objects?
// Every left parenthesis in the regex gets a number, starting with 1.
func (s Match) HasGroup(n int) bool {
	reg := s.Group(n)
	return reg.Start != -1 && reg.End != -1
}

// Does the named capture group have any objects?
func (s Match) HasGroupName(name string) bool {
	regNum, has := s.regNameMap[name]
	if !has {
		return false
	}
	return s.HasGroup(regNum)
}

// Get a numbered capture group from the Match object.
// Every left parenthesis in the regex gets a number, starting with 1.
func (s Match) Group(n int) Range {
	if s.Success && n > 0 && n <= len(s.registers) {
		return s.registers[n-1]
	} else {
		return Range{-1, -1}
	}
}

// Get a named capture group from the Match object.
func (s Match) GroupName(name string) Range {
	regNum, has := s.regNameMap[name]
	if !has {
		return Range{-1, -1}
	}
	return s.Group(regNum)
}

// Match the regex against the input, to the end of the input.
// The match begins at input position 0.
func (s *Regexp[T]) FullMatch(input []T) Match {
	return s.matchAt(input, 0, true)
}

// Match the regex against the input, to the end of the input.
// The match begins at the start position you give.
func (s *Regexp[T]) FullMatchAt(input []T, start int) Match {
	return s.matchAt(input, start, true)
}

// Match the regex against the input. This can match a sub-slice of the input.
// The match begins at input position 0.
func (s *Regexp[T]) Match(input []T) Match {
	return s.matchAt(input, 0, false)
}

// Match the regex against the input. This can match a sub-slice of the input.
// The match begins at the start position you give.
func (s *Regexp[T]) MatchAt(input []T, start int) Match {
	return s.matchAt(input, start, false)
}

func (s *Regexp[T]) matchAt(input []T, start int, full bool) Match {

	var executor executorT[T]
	executor.Initialize(s)

	matched, n, xns := executor.match(s.nfa, input, start, full)
	if matched {
		m := Match{
			Success: true,
			Range: Range{
				Start: start,
				End:   start + n,
			},
			registers:  make([]Range, s.numRegisters),
			regNameMap: s.regNameMap,
		}
		copy(m.registers, xns.registers.ranges)
		return m
	} else {
		return Match{
			Success:    false,
			regNameMap: s.regNameMap,
		}
	}
}

// Search every position within the input to match the Regex.
// The match begins at input position 0.
// If the regexp can only possibly match at at the beinning of
// the input, then this reverts to MatchAt(0)
func (s *Regexp[T]) Search(input []T) Match {
	if s.onlyMatchesAtBeginning() {
		return s.MatchAt(input, 0)
	} else {
		return s.SearchAt(input, 0)
	}
}

// Search every position within the input to match the Regex.
// The match begins at the start position you give.
// If the regexp must have a specific starting object class,
// use that to narrow the search space.
func (s *Regexp[T]) SearchAt(input []T, start int) Match {

	if s.initialObj != nil {
		// Do a quick test of each object before calling
		// regexp.Match()
		for i := start; i < len(input); i++ {
			ch := input[i]
			var iMatch bool
			switch s.initialObj.c {
			case ntClass:
				iMatch = s.initialObj.oClass.Matches(ch)
			case ntIdentity:
				iMatch = s.initialObj.iObj == ch
			case ntDynClass:
				iMatch = s.initialObj.dynClass.Matches(ch)
			}
			if iMatch {
				m := s.MatchAt(input, i)
				if m.Success {
					return m
				}
			}
		}
	} else {
		// Call regexp.Match on each object
		for i := start; i < len(input); i++ {
			m := s.MatchAt(input, i)
			if m.Success {
				return m
			}
		}
	}

	return Match{
		Success: false,
		//Group:   make(map[string]Range),
	}
}

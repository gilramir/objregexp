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

	// the register numbers that start at the beginning
	// of the regex, before any other tokens
	startRegisters []int
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

	root := fmt.Sprintf("Root #r:%d sr:%v", s.numRegisters, s.startRegisters)

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
type Match struct {
	// Did the Regexp find something?
	Success bool

	// The range for the entire sub-string that matched
	Range Range

	registers []Range
	//	Group   map[string]Range
}

func (s Match) Length() int {
	return s.Range.Length()
}

func (s Match) HasRegister(n int) bool {
	reg := s.Register(n)
	return reg.Start != -1 && reg.End != -1
}

// Get a numbered register from the Match. Every left parenthesis
// in the regex gets a number, starting with 1.
func (s *Match) Register(n int) Range {
	if s.Success && n > 0 && n <= len(s.registers) {
		return s.registers[n-1]
	} else {
		return Range{-1, -1}
	}
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

	//vars := make(map[string]Range)
	matched, n, xns := executor.match(s.nfa, input, start, full)
	if matched {
		m := Match{
			Success: true,
			Range: Range{
				Start: start,
				End:   start + n,
			},
			registers: make([]Range, s.numRegisters),
			//Group: vars,
		}
		copy(m.registers, xns.registers.ranges)
		return m
	} else {
		return Match{
			Success: false,
			//Group:   make(map[string]Range),
		}
	}
}

// Search every position within the input to match the Regex.
// The match begins at input position 0.
func (s *Regexp[T]) Search(input []T) Match {
	return s.SearchAt(input, 0)
}

// Search every position within the input to match the Regex.
// The match begins at the start position you give.
func (s *Regexp[T]) SearchAt(input []T, start int) Match {
	// We could reduce the search by knowing the minimum sequence
	// of matchable items in the regex. But we don't have a way
	// to calculate that yet

	for i := start; i < len(input); i++ {
		m := s.MatchAt(input, i)
		if m.Success {
			return m
		}
	}

	return Match{
		Success: false,
		//Group:   make(map[string]Range),
	}
}

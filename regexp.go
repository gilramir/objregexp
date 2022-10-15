// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

// The compiled regex.
type Regexp[T any] struct {
	// the root node of the stack; where the parse begins
	nfa *nfaStateT[T]

	// singleton matching state used to denote all end states
	matchstate nfaStateT[T]

	// How many registers can be saved to by this regex
	numRegisters int
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

// Returned by a Regexp matching-related function.
type Match struct {
	// Did the Regexp find something?
	Success bool

	// The range for the entire sub-string that matched
	Range Range

	registers []Range
	//	Group   map[string]Range
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

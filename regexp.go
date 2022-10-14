package objregexp

type Regexp[T comparable] struct {
	// the root node of the stack; where the parse begins
	nfa *State[T]

	// singleton matching state used to denote all end states
	matchstate State[T]

	// How many registers can be saved to by this regex
	numRegisters int
}

type Range struct {
	Start int
	End   int
}

type Match struct {
	Success   bool
	Range     Range
	registers []Range
	//	Group   map[string]Range
}

func (s *Match) Register(n int) Range {
	if s.Success && n > 0 && n <= len(s.registers) {
		return s.registers[n-1]
	} else {
		return Range{-1, -1}
	}
}

func (s *Regexp[T]) FullMatch(input []T) Match {
	return s.matchAt(input, 0, true)
}

func (s *Regexp[T]) FullMatchAt(input []T, start int) Match {
	return s.matchAt(input, start, true)
}

func (s *Regexp[T]) Match(input []T) Match {
	return s.matchAt(input, 0, false)
}
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

func (s *Regexp[T]) Search(input []T) Match {
	return s.SearchAt(input, 0)
}

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

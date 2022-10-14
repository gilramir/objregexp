package objregexp

import (
	"fmt"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Parse the regular expression string
//
// classgroup		[:alpha:,:beta:]
// lparen		( or (?P<name>
// rparen		)
// curlies		{1} or {1,2}

type tokenType string

// The token types
const (
	tError         tokenType = "E"
	tClass                   = "C" // [:alpha:]
	tConcat                  = "." // concatenate, for internal postfix notation
	tAlternate               = "|" // alternate choices
	tGlobStar                = "*" // *
	tGlobPlus                = "+" // +
	tGlobQuestion            = "?" // ?
	tAny                     = "A" // .
	tStartRegister           = "(" // Record info about the open paren
	tEndRegister             = ")" // Record info about the close paren
)

type tokenT struct {
	ttype tokenType
	pos   int
	//value string

	// For tClass, name is the name of the class
	name string

	// negation is only used For tClass
	negation bool

	// For tStartReg and tEndReg, int1 holds the register number
	int1 int
	//int2 int

	err error
}

func printTokens(tokens []tokenT) {
	for i, t := range tokens {
		fmt.Printf("#%d. %+v\n", i, t)
	}
}

func parseRegex(restring string) ([]tokenT, error) {
	var pstate reParserState
	pstate.Initialize()
	pstate.input = restring

	tokens := make([]tokenT, 0)
	go pstate.goparse()

	for token := range pstate.tokenChan {
		switch token.ttype {
		case tError:
			// There better not be any tokens being written
			// to the channel, and thereby not closing the
			// channel, and leaving the goroutine still running,
			// otherwise we'll be waiting here forever.
			pstate.wg.Wait()
			return nil, token.err
		default:
			tokens = append(tokens, token)
		}
	}
	pstate.wg.Wait()
	return tokens, nil
}

type reParserState struct {
	input     string
	pos       int
	tokenChan chan tokenT
	wg        sync.WaitGroup

	groupNumsAllocated int

	// Was a tError emitted?
	emittedError bool

	// The number of binary choices (alternations) that need to still be emitted
	nbin int

	// Number of atoms emitted that still need to be concatened with
	natom int

	// stack of pre-paren states, so that when a paren is closed,
	// the previous state can be restored
	p []backc

	// The pos in p where the next stack entry can be placed.
	j int
}

type backc struct {
	nbin           int
	natom          int
	beforeGroupNum int
}

func (s *reParserState) Initialize() {
	s.tokenChan = make(chan tokenT)
	s.p = make([]backc, 0)
	s.ensure_stack_space()
}

func (s *reParserState) ensure_stack_space() {
	if len(s.p) <= s.j+1 {
		extra := s.j - len(s.p) + 1
		s.p = append(s.p, make([]backc, extra)...)
	}
}

func (s *reParserState) goparse() {
	s.wg.Add(1)
	defer s.wg.Done()
	defer close(s.tokenChan)

	for {
		// Get the next rune
		ok, r, eof := s.getNextRune()
		if !ok {
			return
		}
		if eof {
			break
		}

		switch r {
		case ' ', '\t', '\n':
			continue

		case '(':
			s.parseLParen()

		case '|':
			s.parsePipe()

		case ')':
			s.parseRParen()

		case '*', '+', '?':
			s.parseGlob(r)

		case '[':
			s.parseLBracket()

		case '.':
			s.parseAny()

		default:
			s.emitErrorf("Syntax error at pos %d starting with '%c'", s.pos, r)
			return
		}

		if s.emittedError {
			return
		}
	}

	// If the stack of saved contexts is not empty, we have an error
	if s.j != 0 {
		s.emitUnexpectedEOF()
		return
	}

	for s.natom--; s.natom > 0; s.natom-- {
		s.emitConcatenation()
	}

	for ; s.nbin > 0; s.nbin-- {
		s.emitAlternation()
	}
}

func (s *reParserState) parseLParen() {

	// First emit the tStartRegister
	/*
		if s.natom > 1 {
			s.natom--
			s.emitConcatenation()
		}
	*/
	s.groupNumsAllocated++
	s.tokenChan <- tokenT{
		ttype: tStartRegister,
		pos:   s.pos,
		int1:  s.groupNumsAllocated,
	}
	//	s.natom++

	// Then do the regular LParen logic
	if s.natom > 1 {
		s.natom--
		s.emitConcatenation()
	}

	s.p[s.j].nbin = s.nbin
	s.p[s.j].natom = s.natom
	s.p[s.j].beforeGroupNum = s.groupNumsAllocated
	fmt.Printf("pstack %d => %+v\n", s.j, s.p[s.j])
	s.j++
	s.ensure_stack_space()
	s.nbin = 0
	s.natom = 0

}

func (s *reParserState) parsePipe() {
	if s.natom == 0 {
		s.emitErrorf("'|' at pos %d is not allowed", s.pos)
		return
	}
	for s.natom--; s.natom > 0; s.natom-- {
		s.emitConcatenation()
	}
	s.nbin++
}

func (s *reParserState) parseRParen() {
	// First emit the regular RParen stuff
	if s.j == 0 || s.natom == 0 {
		s.emitErrorf("Close paren ')' at pos %d doesn't follow an opening paren.", s.pos)
		return
	}
	for s.natom--; s.natom > 0; s.natom-- {
		s.emitConcatenation()
	}
	for ; s.nbin > 0; s.nbin-- {
		s.emitAlternation()
	}
	s.j--
	s.nbin = s.p[s.j].nbin
	s.natom = s.p[s.j].natom
	s.natom++

	// Now emit the tEndRegister
	/*
		if s.natom > 1 {
			s.natom--
			s.emitConcatenation()
		}
	*/
	s.tokenChan <- tokenT{
		ttype: tEndRegister,
		pos:   s.pos,
		int1:  s.p[s.j].beforeGroupNum,
	}
	//	s.natom++
}

func (s *reParserState) parseGlob(r rune) {
	if s.natom == 0 {
		s.emitErrorf("Cannot have glob '%c' at pos %d with no preceding item",
			r, s.pos)
		return
	}

	switch r {
	case '*':
		s.tokenChan <- tokenT{
			ttype: tGlobStar,
			pos:   s.pos,
		}
	case '+':
		s.tokenChan <- tokenT{
			ttype: tGlobPlus,
			pos:   s.pos,
		}
	case '?':
		s.tokenChan <- tokenT{
			ttype: tGlobQuestion,
			pos:   s.pos,
		}
	default:
		panic(fmt.Sprintf("Unexpected '%c' at pos %d", r, s.pos))
	}
}

func (s *reParserState) parseAny() {
	if s.natom > 1 {
		s.natom--
		s.emitConcatenation()
	}
	s.tokenChan <- tokenT{
		ttype: tAny,
		pos:   s.pos,
	}
	s.natom++
}

func (s *reParserState) parseLBracket() {
	ok, r, eof := s.getNextRune()
	if eof {
		s.emitUnexpectedEOF()
		return
	}
	if !ok {
		return
	}

	var negation bool
	// We can start with a negation
	if r == '!' {
		negation = true
		ok, r, eof = s.getNextRune()
		if eof {
			s.emitUnexpectedEOF()
			return
		}
		if !ok {
			return
		}
	}

	// This must be a ':'
	if r != ':' {
		s.emitErrorf("Expected ':' to start a class name at pos %d", s.pos)
		return
	}

	// Read rune names until the ending colon
	nameRunes := make([]rune, 0, 20)

	classPos := s.pos
	for {
		ok, r, eof := s.getNextRune()
		if eof {
			s.emitUnexpectedEOF()
			return
		}
		if !ok {
			return
		}
		if r == ' ' {
			s.emitErrorf("The class name starting at pos %d has a space in it",
				classPos)
			return
		}

		if !unicode.IsGraphic(r) {
			s.emitErrorf("The class name starting at pos %d has a non-graphic Unicode code point in it",
				classPos)
			return
		}

		if r == ':' {
			break
		}
		nameRunes = append(nameRunes, r)
	}

	// We need a final ']'
	ok, r, eof = s.getNextRune()
	if eof {
		s.emitUnexpectedEOF()
		return
	}
	if !ok {
		return
	}

	if r != ']' {
		s.emitErrorf("Expected ] to end a class name at pos %d", s.pos)
		return
	}

	if s.natom > 1 {
		s.natom--
		s.emitConcatenation()
	}

	s.tokenChan <- tokenT{
		ttype:    tClass,
		pos:      classPos,
		name:     string(nameRunes),
		negation: negation,
	}
	s.natom++
}

func (s *reParserState) emitConcatenation() {
	// Add a concatention
	s.tokenChan <- tokenT{
		ttype: tConcat,
		pos:   -1,
	}
}
func (s *reParserState) emitAlternation() {
	s.tokenChan <- tokenT{
		ttype: tAlternate,
	}
}

func (s *reParserState) emitErrorf(f string, args ...any) {
	s.tokenChan <- tokenT{
		ttype: tError,
		pos:   s.pos,
		err:   fmt.Errorf(f, args...),
	}
	s.emittedError = true
}

func (s *reParserState) emitRuneError() {
	s.emitErrorf("Bytes starting at position %d aren't valid UTF-8", s.pos)
}
func (s *reParserState) emitUnexpectedEOF() {
	s.emitErrorf("Unexpected end of string")
}

// Return ok, rune, eof, and advances the pointer
// If not ok, this function calls emitRuneError, via _peekNextRune
func (s *reParserState) getNextRune() (bool, rune, bool) {
	ok, r, size, eof := s._peekNextRune()
	if !ok {
		return false, 0, false
	}

	if ok && !eof {
		s.pos += size
	}
	return ok, r, eof
}

// Return ok, rune, eof, but does not advance the pointer
func (s *reParserState) peekNextRune() (bool, rune, bool) {
	ok, r, _, eof := s._peekNextRune()
	return ok, r, eof
}

// Advances the pointer by 1 rune
// If not ok, this function calls emitRuneError, via _peekNextRune
func (s *reParserState) consumeNextRune() (bool, bool) {
	ok, _, size, eof := s._peekNextRune()
	if ok && !eof {
		s.pos += size
	}
	return ok, eof
}

// Internal helper for get/peek/consume- NextRune()
// If not ok, this function calls emitRuneErro
func (s *reParserState) _peekNextRune() (bool, rune, int, bool) {
	// EOS?
	if s.pos == len(s.input) {
		return true, utf8.RuneError, 0, true
	}
	r, size := utf8.DecodeRuneInString(s.input[s.pos:])
	if r == utf8.RuneError {
		s.emitRuneError()
		return false, utf8.RuneError, size, false
	}
	return true, r, size, false
}

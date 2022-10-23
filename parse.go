// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	"fmt"
	"sync"
	"unicode"
)

type tokenTypeT string

// The token types
const (
	tError        tokenTypeT = "E"
	tClass                   = "C" // [:alpha:]
	tDynClass                = "D" // [:alpha:]
	tConcat                  = "." // concatenate, for internal postfix notation
	tAlternate               = "|" // alternate choices
	tGlobStar                = "*" // *
	tGlobPlus                = "+" // +
	tGlobQuestion            = "?" // ?
	tAny                     = "A" // .
	tEndRegister             = ")" // Record info about the close paren
)

type tokenT[T comparable] struct {
	ttype tokenTypeT
	// position in the regex string; used for reporting syntax
	// errors to the user.
	pos int

	// For tClass, name is the name of the class
	name string

	// negation is only used For tClass
	negation bool

	// For tEndReg, holds the register number
	regNum int

	// An error caught during parsing, to cause the
	// parse to fail, and to be reported to the user.
	err error

	dynClass *dynClassT[T]
}

func makeTokensString[T comparable](tokens []tokenT[T]) string {
	text := ""
	for _, t := range tokens {
		text += string(t.ttype)
	}
	return text
}

func (s *tokenT[T]) Repr() string {
	return fmt.Sprintf("<tokenT[T] %s name:%s neg:%t pos:%d reg#:%d>",
		s.ttype, s.name, s.negation, s.pos, s.regNum)
}

func printTokens[T comparable](tokens []tokenT[T]) {
	for i, t := range tokens {
		dlog.Printf("#%d. %s", i, t.Repr())
	}
}

func parseRegex[T comparable](input string, compiler *Compiler[T]) ([]tokenT[T], error) {
	var pstate reParserStateT[T]
	pstate.Initialize(input, compiler)

	tokens := make([]tokenT[T], 0)
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

	// Here can we clean up the start/end register tokens?
	//printTokens(tokens)
	//dlog.Printf("")

	return tokens, nil
}

type reParserStateT[T comparable] struct {
	input     runeBufferT
	tokenChan chan tokenT[T]
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

	compiler *Compiler[T]
}

type backc struct {
	nbin           int
	natom          int
	beforeGroupNum int
}

func (s *reParserStateT[T]) Initialize(input string, compiler *Compiler[T]) {
	s.input.Initialize(input)
	s.input.runeErrorCb = s.emitRuneError
	s.tokenChan = make(chan tokenT[T])
	s.p = make([]backc, 0)
	s.ensure_stack_space()
	s.compiler = compiler
}

func (s *reParserStateT[T]) ensure_stack_space() {
	if len(s.p) <= s.j+1 {
		extra := s.j - len(s.p) + 1
		s.p = append(s.p, make([]backc, extra)...)
	}
}

func (s *reParserStateT[T]) goparse() {
	s.wg.Add(1)
	defer s.wg.Done()
	defer close(s.tokenChan)

	for {
		// Get the next rune
		ok, r, eof := s.input.getNextRune()
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
			s.parseLBracket2()

		case '.':
			s.parseAny()

		default:
			s.emitErrorf("Syntax error at pos %d starting with '%c'", s.input.pos, r)
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

func (s *reParserStateT[T]) parseLParen() {

	s.groupNumsAllocated++

	// Then do the regular LParen logic
	if s.natom > 1 {
		s.natom--
		s.emitConcatenation()
	}

	s.p[s.j].nbin = s.nbin
	s.p[s.j].natom = s.natom
	s.p[s.j].beforeGroupNum = s.groupNumsAllocated
	//dlog.Printf("pstack %d => %+v", s.j, s.p[s.j])
	s.j++
	s.ensure_stack_space()
	s.nbin = 0
	s.natom = 0

}

func (s *reParserStateT[T]) parsePipe() {
	if s.natom == 0 {
		s.emitErrorf("'|' at pos %d is not allowed", s.input.pos)
		return
	}
	for s.natom--; s.natom > 0; s.natom-- {
		s.emitConcatenation()
	}
	s.nbin++
}

func (s *reParserStateT[T]) parseRParen() {
	// First emit the regular RParen stuff
	if s.j == 0 || s.natom == 0 {
		s.emitErrorf("Close paren ')' at pos %d doesn't follow an opening paren.", s.input.pos)
		return
	}

	dlog.Printf(") => atoms %d nbins %d", s.natom, s.nbin)

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
	s.tokenChan <- tokenT[T]{
		ttype:  tEndRegister,
		pos:    s.input.pos,
		regNum: s.p[s.j].beforeGroupNum,
	}
}

func (s *reParserStateT[T]) parseGlob(r rune) {
	if s.natom == 0 {
		s.emitErrorf("Cannot have glob '%c' at pos %d with no preceding item",
			r, s.input.pos)
		return
	}

	switch r {
	case '*':
		s.tokenChan <- tokenT[T]{
			ttype: tGlobStar,
			pos:   s.input.pos,
		}
	case '+':
		s.tokenChan <- tokenT[T]{
			ttype: tGlobPlus,
			pos:   s.input.pos,
		}
	case '?':
		s.tokenChan <- tokenT[T]{
			ttype: tGlobQuestion,
			pos:   s.input.pos,
		}
	default:
		panic(fmt.Sprintf("Unexpected '%c' at pos %d", r, s.input.pos))
	}
}

func (s *reParserStateT[T]) parseAny() {
	if s.natom > 1 {
		s.natom--
		s.emitConcatenation()
	}
	s.tokenChan <- tokenT[T]{
		ttype: tAny,
		pos:   s.input.pos,
	}
	s.natom++
}

func (s *reParserStateT[T]) parseLBracket2() {
	// Look for the RBracket, but take into consideration
	// that the class name can have a RBracket in it.
	startPos := s.input.pos
	endPos := -1

	needColon := false
	for {
		// set endPos here; if we get ']' it will be correct.
		endPos = s.input.pos
		ok, r, eof := s.input.getNextRune()
		if eof {
			s.emitUnexpectedEOF()
			return
		}
		if !ok {
			return
		}
		if needColon {
			if r == ':' {
				needColon = false
				continue
			}
		} else {
			if r == ']' {
				break
			}
		}
	}

	text := s.input.getStringSlice(startPos, endPos)

	// Do we have just 1 class name, or more?
	negation := false
	numColons := 0
	fcPos := 0
	scPos := 0
	for i, c := range text {
		if numColons == 0 && c == '!' {
			negation = true
		}
		if c == ':' {
			if numColons == 0 {
				fcPos = i
			} else if numColons == 1 {
				scPos = i
			}
			numColons++
			if numColons == 3 {
				break
			}
		}
	}

	if s.natom > 1 {
		s.natom--
		s.emitConcatenation()
	}

	if numColons > 2 {
		dynClass, err := newDynClassT[T](text, startPos, s.compiler)
		if err != nil {
			s.emitErrorf("Parsing class string at pos %d: %s",
				startPos, err)
			return
		}

		s.tokenChan <- tokenT[T]{
			ttype:    tDynClass,
			pos:      startPos,
			name:     text,
			dynClass: dynClass,
		}
	} else {
		s.tokenChan <- tokenT[T]{
			ttype:    tClass,
			pos:      startPos,
			name:     text[fcPos+1 : scPos],
			negation: negation,
		}
	}
	s.natom++
}

func (s *reParserStateT[T]) parseLBracket() {
	ok, r, eof := s.input.getNextRune()
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
		ok, r, eof = s.input.getNextRune()
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
		s.emitErrorf("Expected ':' to start a class name at pos %d", s.input.pos)
		return
	}

	// Read rune names until the ending colon
	nameRunes := make([]rune, 0, 20)

	classPos := s.input.pos
	for {
		ok, r, eof := s.input.getNextRune()
		if eof {
			s.emitUnexpectedEOF()
			return
		}
		if !ok {
			return
		}

		// Allow anything except non-visible code point glyphs.
		// But do allow spaces
		if !unicode.IsGraphic(r) && r != ' ' {
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
	ok, r, eof = s.input.getNextRune()
	if eof {
		s.emitUnexpectedEOF()
		return
	}
	if !ok {
		return
	}

	if r != ']' {
		s.emitErrorf("Expected ] to end a class name at pos %d", s.input.pos)
		return
	}

	if s.natom > 1 {
		s.natom--
		s.emitConcatenation()
	}

	s.tokenChan <- tokenT[T]{
		ttype:    tClass,
		pos:      classPos,
		name:     string(nameRunes),
		negation: negation,
	}
	s.natom++
}

func (s *reParserStateT[T]) emitConcatenation() {
	// Add a concatention
	s.tokenChan <- tokenT[T]{
		ttype: tConcat,
		pos:   -1,
	}
}
func (s *reParserStateT[T]) emitAlternation() {
	s.tokenChan <- tokenT[T]{
		ttype: tAlternate,
	}
}

func (s *reParserStateT[T]) emitErrorf(f string, args ...any) {
	s.tokenChan <- tokenT[T]{
		ttype: tError,
		pos:   s.input.pos,
		err:   fmt.Errorf(f, args...),
	}
	s.emittedError = true
}

func (s *reParserStateT[T]) emitRuneError() {
	s.emitErrorf("Bytes starting at position %d aren't valid UTF-8", s.input.pos)
}
func (s *reParserStateT[T]) emitUnexpectedEOF() {
	s.emitErrorf("Unexpected end of string")
}

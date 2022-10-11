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
// error

type tokenType string

// The token types
const (
	tError        tokenType = "E"
	tClass                  = "C" // [:alpha:]
	tConcat                 = "." // conctenate, for internal postfix notation
	tGlobStar               = "*" // *
	tGlobPlus               = "+" // +
	tGlobQuestion           = "?" // ?
)

type tokenT struct {
	ttype tokenType
	pos   int
	value string

	name     string
	negation bool

	int1 int
	int2 int

	err error
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

	insideClass   bool
	classStartPos int

	insideParen   bool
	parenStartPos int
	parenName     string

	firstAtomEmitted            bool
	firstInsideParenAtomEmitted bool
}

func (s *reParserState) Initialize() {
	s.tokenChan = make(chan tokenT)
}

func (s *reParserState) goparse() {
	s.wg.Add(1)
	defer s.wg.Done()
	defer close(s.tokenChan)

	// Start at the initial state, and get the next state,
	// over and over again, entil we reach the final state
	// (nil)
	var state stateFunc
	for state = s.stateNewAtom; state != nil; {
		state = state()
	}

	// Are we in an incomplete state?
	/*
		if s.insideParen {
			if s.parenName == "" {
				s.emitErrorf("Objregexp finished before clsoing the opening paren at position %d",
					s.groupStartPos)
			} else {
				s.emitErrorf("Objregexp finished before finishing group-list '%s' starting at position %d",
					s.groupListName, s.groupListStartPos)
			}
		}

		if s.insideParen {
			s.emitErrorf("Objregexp finished before finishing opening parentheis at position %d",
				s.parenStartPos)
		}
	*/
}

func (s *reParserState) emitConcatenation() {
	// Add a concatention
	s.tokenChan <- tokenT{
		ttype: tConcat,
		pos:   -1,
	}
}

func (s *reParserState) emitErrorf(f string, args ...any) {
	s.tokenChan <- tokenT{
		ttype: tError,
		pos:   s.pos,
		err:   fmt.Errorf(f, args...),
	}
}

func (s *reParserState) emitRuneError() {
	s.emitErrorf("Bytes starting at position %d aren't valid UTF-8", s.pos)
}
func (s *reParserState) emitUnexpectedEOF() {
	s.emitErrorf("Unexpected end of string")
}

// a parser state is a function which returns the next parser state
type stateFunc func() stateFunc

// Return ok, rune, eof, and advances the pointer
func (s *reParserState) getNextRune() (bool, rune, bool) {
	ok, r, size, eof := s._peekNextRune()
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
func (s *reParserState) consumeNextRune() (bool, bool) {
	ok, _, size, eof := s._peekNextRune()
	if ok && !eof {
		s.pos += size
	}
	return ok, eof
}

// Internal helper for get/peek/consume- NextRune()
func (s *reParserState) _peekNextRune() (bool, rune, int, bool) {
	// EOS?
	if s.pos == len(s.input) {
		return true, utf8.RuneError, 0, true
	}
	r, size := utf8.DecodeRuneInString(s.input[s.pos:])
	if r == utf8.RuneError {
		return false, utf8.RuneError, size, false
	}
	return true, r, size, false
}

// At the beginning of a new item
func (s *reParserState) stateNewAtom() stateFunc {
	startPos := s.pos
	ok, r, eof := s.getNextRune()
	if eof {
		return nil
	}
	if !ok {
		s.emitRuneError()
		return nil
	}

	// We can expect a new paren or group, or whitespace.

	switch r {
	case '(':
		s.insideParen = true
		s.firstInsideParenAtomEmitted = false
		return s.stateOpenParen
	case '[':
		s.insideClass = true
		s.classStartPos = startPos
		return s.stateClass
	case ' ':
		return s.stateNewAtom
	case '\n':
		return s.stateNewAtom
	case '\t':
		return s.stateNewAtom
	default:
		s.emitErrorf("At position %d '%c' is illegal",
			s.pos, r)
		return nil
	}
}

func (s *reParserState) stateClass() stateFunc {

	ok, r, eof := s.getNextRune()
	if eof {
		s.emitUnexpectedEOF()
		return nil
	}
	if !ok {
		s.emitRuneError()
		return nil
	}

	var negation bool
	// We can start with a negation
	if r == '!' {
		negation = true
		ok, r, eof = s.getNextRune()
		if eof {
			s.emitUnexpectedEOF()
			return nil
		}
		if !ok {
			s.emitRuneError()
			return nil
		}
	}

	// This must be a ':'
	if r != ':' {
		s.emitErrorf("Expected : to start a class name at pos %d", s.pos)
		return nil
	}

	// Read rune names until the ending colon
	nameRunes := make([]rune, 0, 20)

	classPos := s.pos
	for {
		ok, r, eof := s.getNextRune()
		if eof {
			s.emitUnexpectedEOF()
			return nil
		}
		if !ok {
			s.emitRuneError()
			return nil
		}
		if r == ' ' {
			s.emitErrorf("The class name starting at pos %d has a space in it",
				classPos)
			return nil
		}

		if !unicode.IsGraphic(r) {
			s.emitErrorf("The class name starting at pos %d has a non-graphic Unicode code point in it",
				classPos)
			return nil
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
		return nil
	}
	if !ok {
		s.emitRuneError()
		return nil
	}

	if r != ']' {
		s.emitErrorf("Expected ] to end a class name at pos %d", s.pos)
		return nil
	}

	s.tokenChan <- tokenT{
		ttype:    tClass,
		pos:      s.classStartPos,
		name:     string(nameRunes),
		negation: negation,
	}

	if s.insideParen {
		if s.firstInsideParenAtomEmitted {
			s.emitConcatenation()
		} else {
			s.firstInsideParenAtomEmitted = true
		}
		return s.stateInsideParen
	} else {
		return s.stateAfterClass
	}
}

// After a class is emitted (outside of parens), we might see a count modifer
func (s *reParserState) stateAfterClass() stateFunc {
	ok, r, eof := s.peekNextRune()
	if eof {
		// It's okay to end the string here
		// but if we do, ensure that a tConcat is emitted
		// if needed.
		if s.firstAtomEmitted {
			s.emitConcatenation()
		} else {
			s.firstAtomEmitted = true
		}
		return nil
	}
	if !ok {
		s.emitRuneError()
		return nil
	}
	// If we already emitted something, we can see
	// a count modifier
	switch r {
	case '*':
		_, _ = s.consumeNextRune()
		s.tokenChan <- tokenT{
			ttype: tGlobStar,
			pos:   s.pos,
		}
	case '+':
		_, _ = s.consumeNextRune()
		s.tokenChan <- tokenT{
			ttype: tGlobPlus,
			pos:   s.pos,
		}
	case '?':
		_, _ = s.consumeNextRune()
		s.tokenChan <- tokenT{
			ttype: tGlobQuestion,
			pos:   s.pos,
		}
	default:
		// no-op
	}
	if s.firstAtomEmitted {
		s.emitConcatenation()
	} else {
		s.firstAtomEmitted = true
	}
	return s.stateNewAtom
}

func (s *reParserState) stateOpenParen() stateFunc {
	return nil
}

func (s *reParserState) stateInsideParen() stateFunc {
	return nil
}

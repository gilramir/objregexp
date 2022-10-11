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
// glob			* or + or ?
// error

type tokenType string

// The token types
const (
	tError tokenType = "E"
	tClass           = "C" // [:alpha:]
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
	for state = s.stateBeginning; state != nil; {
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

// Return ok, rune, eof
func (s *reParserState) getNextRune() (bool, rune, bool) {
	// EOS?
	if s.pos == len(s.input) {
		return true, utf8.RuneError, true
	}
	r, size := utf8.DecodeRuneInString(s.input[s.pos:])
	if r == utf8.RuneError {
		return false, utf8.RuneError, false
	}
	s.pos += size
	return true, r, false
}

// At the beginning of a new item
func (s *reParserState) stateBeginning() stateFunc {
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
		return s.stateOpenParen
	case '[':
		s.insideClass = true
		s.classStartPos = startPos
		return s.stateClass
	case ' ':
		return s.stateBeginning
	case '\n':
		return s.stateBeginning
	case '\t':
		return s.stateBeginning
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

	return nil
}

func (s *reParserState) stateOpenParen() stateFunc {
	return nil
}

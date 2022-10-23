// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

// Convert infix to postfix using Shuting yard algorithm:
// https://en.wikipedia.org/wiki/Shunting_yard_algorithm

import (
	"fmt"
	"sync"
	"unicode"
)

type dynClassT[T comparable] struct {
	ops []dynClassOpT[T]
}

type dcopTypeT string

const (
	dcClass       dcopTypeT = "C"
	dcIdentity              = "I"
	dcNot                   = "!"
	dcAssertTrue            = "?"
	dcJumpIfTrue            = "T"
	dcJumpIfFalse           = "F"
)

type dynClassOpT[T comparable] struct {
	opType dcopTypeT

	// oClass is set if opType is dcClass
	oClass *Class[T]

	// iObj is set if opType is dcIdentiy
	iObj T

	// cName is set if opType is dcClass or dcIdentity
	// TODO - do I need this?
	cName string

	// If this is a jump, which instruction to jump to
	jmpTo int
}

func newDynClassT[T comparable](text string, compiler *Compiler[T]) (*dynClassT[T], error) {
	s := &dynClassT[T]{
		ops: make([]dynClassOpT[T], 0),
	}
	err := s.parse(text, compiler)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type jmpInfoT struct {
	insnPos   int
	jmpTarget int
}

func (s *dynClassT[T]) parse(text string, compiler *Compiler[T]) error {

	// Produce a slice of postfix-ordered tokens
	tokens, err := s.tokenize(text)
	if err != nil {
		return err
	}
	for ti, tok := range tokens {
		dlog.Printf("token #%d: %+v\n", ti, tok)
	}

	// convert to opcodes
	ops := make([]dynClassOpT[T], 0, len(tokens))

	jmpInfos := make([]jmpInfoT, 0)

	// Key = target, Value = pos
	jmpAt := make(map[int]int)

	for ti, tok := range tokens {
		op := dynClassOpT[T]{}

		switch tok.ttype {
		case dctClass:

			/*
				if tok.name[0] != ':' {
					panic(fmt.Sprintf("Token %s doesn't begin with :", tok.name))
				}
				if tok.name[len(tok.name)-1] != ':' {
					panic(fmt.Sprintf("Token %s doesn't end with :", tok.name))
				}
				// We konw that each ":" is only 1 byte long
				name := tok.name[1 : len(tok.name)-2]
			*/
			name := tok.name

			ctype, has := compiler.namespace[name]
			if !has {
				return fmt.Errorf("Class :%s: at pos %d is unknown",
					name, tok.pos)
			}
			switch ctype {
			case ccClass:
				op.opType = dcClass
				op.oClass = compiler.classMap[tok.name]
			case ccIdentity:
				op.opType = dcIdentity
				op.iObj = compiler.identityObj[tok.name]
			default:
				panic(fmt.Sprintf("Unexpected class type %v for %s", tok.ttype, name))
			}
			op.cName = name

		case dctNot:
			op.opType = dcNot

		case dctAssertTrue:
			op.opType = dcAssertTrue
			if tok.jmpTarget > 0 {
				jmpAt[tok.jmpTarget] = ti
			}

		case dctJumpIfFalse:
			op.opType = dcJumpIfFalse
			if tok.jmpTarget > 0 {
				jmpInfos = append(jmpInfos, jmpInfoT{insnPos: ti, jmpTarget: tok.jmpTarget})
			}

		case dctJumpIfTrue:
			op.opType = dcJumpIfTrue
			if tok.jmpTarget > 0 {
				jmpInfos = append(jmpInfos, jmpInfoT{insnPos: ti, jmpTarget: tok.jmpTarget})
			}

		default:
			panic(fmt.Sprintf("Unexpected token type %v", tok.ttype))
		}
		ops = append(ops, op)
	}

	// Fix up the jumps
	for _, ji := range jmpInfos {

		jmpToPos := jmpAt[ji.jmpTarget]
		if jmpToPos == 0 {
			panic(fmt.Sprintf("jmp at pos %d to target %d not found", ji.insnPos, ji.jmpTarget))
		}
		ops[ji.insnPos].jmpTo = jmpToPos
	}

	s.ops = ops
	return nil
}

func (s *dynClassT[T]) tokenize(input string) ([]dcTokenT, error) {
	var pstate dcParserStateT[T]
	pstate.Initialize(input)

	// Produce a slice of postfix-ordered tokens
	tokens, err := pstate.parse()
	return tokens, err
}

type dcTokenTypeT string

// The token types
const (
	dctError       dcTokenTypeT = "E"
	dctClass                    = "C" // :alpha:
	dctAssertTrue               = "?" // Final check for && and ||
	dctNot                      = "!"
	dctLParen                   = "("
	dctRParen                   = ")"
	dctJumpIfFalse              = "F" // short-circuit for &&
	dctJumpIfTrue               = "T" // short-circuit for ||
)

const (
	notPrecedence = 1
	andPrecedence = 2
	orPrecedence  = 3
)

type dcTokenT struct {
	ttype dcTokenTypeT

	// position in the dynamic class string; used for reporting syntax
	// errors to the user.
	pos int

	// For dctClass, name is the name of the class
	name string

	// where to jump to for JumpIfFalse and JumpIfTrue
	// if on an Assert operand, and jmpTarget is set, the next insns
	// is the target
	jmpTarget int

	precedence int

	// An error caught during parsing, to cause the
	// parse to fail, and to be reported to the user.
	err error
}

type dcParserStateT[T comparable] struct {
	input runeBufferT
	ops   []dynClassOpT[T]
	stack Stack[dcTokenT]

	nextJumpTarget int

	tokenChan chan dcTokenT
	wg        sync.WaitGroup

	// Was a dctError emitted?
	emittedError bool
}

func (s *dcParserStateT[T]) Initialize(input string) {
	s.input.Initialize(input)
	s.input.runeErrorCb = s.emitRuneError
	s.tokenChan = make(chan dcTokenT)
	s.stack = NewStack[dcTokenT]()
}

func (s *dcParserStateT[T]) parse() ([]dcTokenT, error) {
	tokens := make([]dcTokenT, 0)
	go s.goparse()

	for token := range s.tokenChan {
		switch token.ttype {
		case dctError:
			// There better not be any tokens being written
			// to the channel, and thereby not closing the
			// channel, and leaving the goroutine still running,
			// otherwise we'll be waiting here forever.
			s.wg.Wait()
			return nil, token.err
		default:
			tokens = append(tokens, token)
		}
	}
	s.wg.Wait()

	// Here can we clean up the start/end register tokens?
	//printTokens(tokens)
	//dlog.Printf("")

	return tokens, nil

}

func (s *dcParserStateT[T]) goparse() {
	s.wg.Add(1)
	defer s.wg.Done()
	defer close(s.tokenChan)

	allowClass := true
	allowAndOr := false

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
			allowClass = true
			allowAndOr = false

		case '|':
			if allowAndOr {
				s.parsePipe()
				allowClass = true
			} else {
				s.emitErrorf("|| is not allowed at pos %d", s.input.pos)
				return
			}

		case '&':
			if allowAndOr {
				s.parseAmpersand()
				allowClass = true
			} else {
				s.emitErrorf("&& is not allowed at pos %d", s.input.pos)
				return
			}

		case ')':
			s.parseRParen()
			allowClass = false
			allowAndOr = true

		case ':':
			if allowClass {
				s.parseColon()
				allowClass = false
				allowAndOr = true
			} else {
				s.emitErrorf("Class name not allowed at pos %d", s.input.pos)
				return
			}

		case '!':
			s.parseBang()
			allowClass = true
			allowAndOr = false

		default:
			s.emitErrorf("Syntax error at pos %d starting with '%c'", s.input.pos, r)
			return
		}

		if s.emittedError {
			return
		}
	}

	for s.stack.Size() > 0 {
		tok := s.stack.Top()
		if tok.ttype == dctLParen {
			s.emitErrorf("Unbalanced left paren starting at pos %d", tok.pos)
			return
		} else {
			s.tokenChan <- tok
			s.stack.Pop()
		}
	}
}

func (s *dcParserStateT[T]) parseLParen() {
	s.stack.Push(dcTokenT{
		ttype: dctLParen,
		pos:   s.input.pos,
	})
}
func (s *dcParserStateT[T]) parseRParen() {
	for tok := s.stack.Top(); tok.ttype != dctLParen; {
		s.tokenChan <- tok
		s.stack.Pop()
		if s.stack.Size() == 0 {
			s.emitErrorf("Unbalanced right paren at pos %d", s.input.pos)
			return
		}
		tok = s.stack.Top()
	}
	// the tok on the top of a stack was a LParen; pop & discard it
	s.stack.Pop()
}

func (s *dcParserStateT[T]) parsePipe() {
	startPos := s.input.pos
	ok, r, eof := s.input.getNextRune()
	if eof {
		s.emitUnexpectedEOF()
		return
	}
	if !ok {
		return
	}
	if r != '|' {
		s.emitErrorf("Expected 2 |'s at pos %d", startPos)
		return
	}

	s.nextJumpTarget++
	jmpTarget := s.nextJumpTarget

	for s.stack.Size() > 0 {
		tok := s.stack.Top()
		if tok.ttype == dctLParen || tok.precedence >= orPrecedence {
			break
		} else {
			s.tokenChan <- tok
			s.stack.Pop()
		}
	}
	s.tokenChan <- dcTokenT{
		ttype:     dctJumpIfTrue,
		pos:       startPos,
		jmpTarget: jmpTarget,
	}

	s.stack.Push(dcTokenT{
		ttype:      dctAssertTrue,
		pos:        s.input.pos,
		jmpTarget:  jmpTarget,
		precedence: orPrecedence,
	})
}

func (s *dcParserStateT[T]) parseAmpersand() {
	startPos := s.input.pos
	ok, r, eof := s.input.getNextRune()
	if eof {
		s.emitUnexpectedEOF()
		return
	}
	if !ok {
		return
	}
	if r != '&' {
		s.emitErrorf("Expected 2 &'s at pos %d", startPos)
		return
	}

	s.nextJumpTarget++
	jmpTarget := s.nextJumpTarget

	for s.stack.Size() > 0 {
		tok := s.stack.Top()
		if tok.ttype == dctLParen || tok.precedence >= andPrecedence {
			break
		} else {
			s.tokenChan <- tok
			s.stack.Pop()
		}
	}
	s.tokenChan <- dcTokenT{
		ttype:     dctJumpIfFalse,
		pos:       startPos,
		jmpTarget: jmpTarget,
	}

	s.stack.Push(dcTokenT{
		ttype:      dctAssertTrue,
		pos:        s.input.pos,
		jmpTarget:  jmpTarget,
		precedence: andPrecedence,
	})
}
func (s *dcParserStateT[T]) parseBang() {
	for s.stack.Size() > 0 {
		tok := s.stack.Top()
		if tok.ttype == dctLParen || tok.precedence >= notPrecedence {
			break
		} else {
			s.tokenChan <- tok
			s.stack.Pop()
		}
	}

	s.stack.Push(dcTokenT{
		ttype:      dctNot,
		pos:        s.input.pos,
		precedence: notPrecedence,
	})
}
func (s *dcParserStateT[T]) parseColon() {
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

	s.tokenChan <- dcTokenT{
		ttype: dctClass,
		pos:   classPos,
		name:  string(nameRunes),
	}

}

func (s *dcParserStateT[T]) emitErrorf(f string, args ...any) {
	s.tokenChan <- dcTokenT{
		ttype: dctError,
		pos:   s.input.pos,
		err:   fmt.Errorf(f, args...),
	}
	s.emittedError = true
}

func (s *dcParserStateT[T]) emitRuneError() {
	s.emitErrorf("Bytes starting at position %d aren't valid UTF-8", s.input.pos)
}
func (s *dcParserStateT[T]) emitUnexpectedEOF() {
	s.emitErrorf("Unexpected end of string")
}

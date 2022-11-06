package objregexp

import (
	"unicode/utf8"
)

type runeBufferT struct {
	input string
	pos   int

	runeErrorCb func()
}

func (s *runeBufferT) Initialize(input string) {
	s.input = input
	s.pos = 0
}

// Return ok, rune, eof, and advances the pointer
// If not ok, this function calls runeErrorCb, via _peekNextRune
func (s *runeBufferT) getNextRune() (bool, rune, bool) {
	ok, r, size, eof := s._peekNextRune(true)
	if !ok {
		return false, 0, false
	}

	if ok && !eof {
		s.pos += size
	}
	return ok, r, eof
}

// Return ok, rune, eof, but does not advance the pointer
// on error, this does *not* call runeErrorCb.
func (s *runeBufferT) peekNextRune() (bool, rune, bool) {
	ok, r, _, eof := s._peekNextRune(false)
	return ok, r, eof
}

// Advances the pointer by 1 rune
// If not ok, this function calls runeErrorCb, via _peekNextRune
func (s *runeBufferT) consumeNextRune() (bool, bool) {
	ok, _, size, eof := s._peekNextRune(true)
	if ok && !eof {
		s.pos += size
	}
	return ok, eof
}

// Internal helper for get/peek/consume- NextRune()
// If not ok, this function calls runeErrorCb
func (s *runeBufferT) _peekNextRune(useCb bool) (bool, rune, int, bool) {
	// EOS?
	if s.pos == len(s.input) {
		return true, utf8.RuneError, 0, true
	}
	r, size := utf8.DecodeRuneInString(s.input[s.pos:])
	if r == utf8.RuneError {
		if useCb && s.runeErrorCb != nil {
			s.runeErrorCb()
		}
		return false, utf8.RuneError, size, false
	}
	return true, r, size, false
}

func (s *runeBufferT) getStringSlice(start int, end int) string {
	return string(s.input[start:end])
}

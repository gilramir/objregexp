package objregexp

import (
	. "gopkg.in/check.v1"
)

// Test Match vs FullMatch
func (s *MySuite) TestRegexp01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_vowel, err := compiler.Compile("[:vowel:]")
	c.Assert(err, IsNil)

	input := []rune{'A'}
	m := re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A'}
	m = re_vowel.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 1)

	input = []rune{'A', 'A'}
	m = re_vowel.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 1)
}

// Test the "any" meta character (".")
func (s *MySuite) TestRegexp02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(DigitClass)
	compiler.Finalize()

	re_vowel, err := compiler.Compile("[:digit:] . [:digit:]")
	c.Assert(err, IsNil)

	input := []rune{'9', '0', '1'}
	m := re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'9', 'A', '1'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'9', 'X', '1'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'9', 'X', 'Y'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'M', 'X', '1'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'M', 'X', 'Z'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

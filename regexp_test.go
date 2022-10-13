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

// Test Search
func (s *MySuite) TestRegexp03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_vowel, err := compiler.Compile("[:vowel:]")
	c.Assert(err, IsNil)

	input := []rune{'B', 'B'}
	m := re_vowel.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B', 'B'}
	m = re_vowel.Search(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B', 'A'}
	m = re_vowel.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B', 'A'}
	m = re_vowel.Search(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 1)
	c.Check(m.Range.End, Equals, 2)
}

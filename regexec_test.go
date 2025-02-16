// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	. "gopkg.in/check.v1"
)

// Test Match vs FullMatch
func (s *MySuite) TestRegexp01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
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
func (s *MySuite) TestRegexp02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
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

// Simple concatenation
func (s *MySuite) TestRegexp03a(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:] [:consonant:]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8'}
	m := re.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'C'}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'8', 'C'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
}

// Simple alternation
func (s *MySuite) TestRegexp03b(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:] | [:consonant:]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'C'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A'}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)
}

// Simple parens, no alternation
func (s *MySuite) TestRegexp04a(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:] ( [:vowel:] ) [:consonant:]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8', 'A', 'B'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 3)
	c.Check(m.Group(1).Start, Equals, 1)
	c.Check(m.Group(1).End, Equals, 2)

	/*
		c.Assert(len(tokens), Equals, 5)

		c.Check(tokens[0].ttype, Equals, tokenType(tClass))
		c.Check(tokens[0].name, Equals, "foo")

		c.Check(tokens[1].ttype, Equals, tokenType(tClass))
		c.Check(tokens[1].name, Equals, "alpha")

		c.Check(tokens[2].ttype, Equals, tokenType(tClass))
		c.Check(tokens[2].name, Equals, "bar")

		c.Check(tokens[3].ttype, Equals, tokenType(tConcat))
		c.Check(tokens[4].ttype, Equals, tokenType(tConcat))
	*/
}

// Parens with fixed-sized alternation
func (s *MySuite) TestRegexp04b(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:] ( [:vowel:] | [:consonant:] ) [:digit:]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8', 'A', '8'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 3)
	c.Check(m.Group(1).Start, Equals, 1)
	c.Check(m.Group(1).End, Equals, 2)
}

// Nested parens
func (s *MySuite) TestRegexp04c(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:] ( [:digit:] ( [:vowel:] | [:consonant:] ) )?"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)
	//	err = re.WriteDot("TestRegexp04c.dot")
	//	c.Assert(err, IsNil)

	input := []rune{'8', '7', 'A'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 3)
	c.Check(m.Group(1).Start, Equals, 1)
	c.Check(m.Group(1).End, Equals, 3)
	c.Check(m.Group(2).Start, Equals, 2)
	c.Check(m.Group(2).End, Equals, 3)

	/*
		c.Assert(len(tokens), Equals, 5)

		c.Check(tokens[0].ttype, Equals, tokenType(tClass))
		c.Check(tokens[0].name, Equals, "foo")

		c.Check(tokens[1].ttype, Equals, tokenType(tClass))
		c.Check(tokens[1].name, Equals, "alpha")

		c.Check(tokens[2].ttype, Equals, tokenType(tClass))
		c.Check(tokens[2].name, Equals, "bar")

		c.Check(tokens[3].ttype, Equals, tokenType(tConcat))
		c.Check(tokens[4].ttype, Equals, tokenType(tConcat))
	*/
}

// Parens with variable-size alternation
func (s *MySuite) TestRegexp04d(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:] ( [:vowel:]+ | [:consonant:] ) [:digit:]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8', 'A', 'E', 'I', '8'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 5)
	c.Check(m.Group(1).Start, Equals, 1)
	c.Check(m.Group(1).End, Equals, 4)

	input = []rune{'8', 'A', 'E', '8'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 4)
	c.Check(m.Group(1).Start, Equals, 1)
	c.Check(m.Group(1).End, Equals, 3)

	input = []rune{'8', 'X', '8'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 3)
	c.Check(m.Group(1).Start, Equals, 1)
	c.Check(m.Group(1).End, Equals, 2)
}

// Test Glob * for greediness
func (s *MySuite) TestRegexpGlob01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:]*"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8', '9', '0'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 3)

	input = []rune{'7'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 1)

	/*
		TODO - handle this

		input = []rune{}
		m = re.Match(input)
		c.Check(m.Success, Equals, true)
		c.Check(m.Range.Start, Equals, 0)
		c.Check(m.Range.End, Equals, 0)
	*/

	input = []rune{'A'}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)
}

// Test Glob + for greediness
func (s *MySuite) TestRegexpGlob02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:]+"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8', '9', '0'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 3)

	input = []rune{'7'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 1)

	input = []rune{'A'}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)
}

// Test Glob ? for greediness
// Test Glob ? for greediness
func (s *MySuite) TestRegexpGlob03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:][:digit:]?"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'8', '9', '0'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 2)

	input = []rune{'7', '8'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 2)

	input = []rune{'8'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 1)
}

func (s *MySuite) TestRegexpGroup01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "([:digit:]|[:vowel:][:consonant:]) ([:digit:])"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'a', 'm', '9', '0'}
	m := re.MatchAt(input, 2)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 2)
	c.Check(m.Range.End, Equals, 4)

	g := m.Group(1)
	c.Check(g.Start, Equals, 2)
	c.Check(g.End, Equals, 3)

	g = m.Group(2)
	c.Check(g.Start, Equals, 3)
	c.Check(g.End, Equals, 4)
}

// Correctly match an empty sequence
func (s *MySuite) TestRegexpMatchEmpty01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "[:vowel:]*"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)
}

// Correctly match an empty sequence with $
func (s *MySuite) TestRegexpMatchEmpty02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "[:vowel:]*$"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)
}

// Correctly FullMatch an empty sequence
func (s *MySuite) TestRegexpMatchEmpty03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "[:vowel:]*"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)
}

// Correctly FullMatch an empty sequence, with $
func (s *MySuite) TestRegexpMatchEmpty04(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "[:vowel:]*$"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)
}

// Correctly match an empty sequence, with group
func (s *MySuite) TestRegexpMatchEmpty05(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "([:vowel:]*)"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)

	c.Check(m.Group(1).Start, Equals, -1)
	c.Check(m.Group(1).End, Equals, -1)
}

// Correctly match an empty sequence with $, with group
func (s *MySuite) TestRegexpMatchEmpty06(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "([:vowel:]*)$"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)

	c.Check(m.Group(1).Start, Equals, -1)
	c.Check(m.Group(1).End, Equals, -1)
}

// Correctly FullMatch an empty sequence, with group
func (s *MySuite) TestRegexpMatchEmpty07(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "([:vowel:]*)"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)

	c.Check(m.Group(1).Start, Equals, -1)
	c.Check(m.Group(1).End, Equals, -1)
}

// Correctly FullMatch an empty sequence, with $, with group
func (s *MySuite) TestRegexpMatchEmpty08(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	text := "([:vowel:]*)$"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 0)

	c.Check(m.Group(1).Start, Equals, -1)
	c.Check(m.Group(1).End, Equals, -1)
}

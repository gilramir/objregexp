// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	. "gopkg.in/check.v1"
)

var vowels = []rune{'A', 'E', 'I', 'O', 'U'}
var consonants = []rune{'B', 'C', 'D', 'F', 'G', 'H', 'J', 'K',
	'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z'}
var digits = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

var VowelClass = &Class[rune]{
	"vowel",
	func(r rune) bool {
		for _, t := range vowels {
			if r == t {
				return true
			}
		}
		return false
	},
}

var ConsonantClass = &Class[rune]{
	"consonant",
	func(r rune) bool {
		for _, t := range consonants {
			if r == t {
				return true
			}
		}
		return false
	},
}

var DigitClass = &Class[rune]{
	"digit",
	func(r rune) bool {
		for _, t := range digits {
			if r == t {
				return true
			}
		}
		return false
	},
}

/*
var ioMap = map[string]string{
	"A": "A",
	"B": "B",
	"C": "C",
}

func ioMapper(name string, inputTarget string) *TASTNode {
	values := make([]string, 1)
	values[0] = string(inputTarget)
	return &TASTNode{
		values:   values,
		consumed: 1,
	}
}

*/

// SImple one-class test
func (s *MySuite) TestClass01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_vowel, err := compiler.Compile("[:vowel:]")
	c.Assert(err, IsNil)

	input := []rune{'A'}
	m := re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'B'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)

	// Check that state is re-set properly and that
	// a match can happen again.
	input = []rune{'A'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)
}

// Test negation
func (s *MySuite) TestClass02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_not_vowel, err := compiler.Compile("[!:vowel:]")
	c.Assert(err, IsNil)

	input := []rune{'A'}
	m := re_not_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B'}
	m = re_not_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)

	// Check that state is re-set properly and that
	// a failed-match can happen again.
	input = []rune{'A'}
	m = re_not_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Test two-classes in sequence
func (s *MySuite) TestClass03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.RegisterClass(ConsonantClass)
	compiler.Finalize()

	re_vc, err := compiler.Compile("[:vowel:] [:consonant:]")
	c.Assert(err, IsNil)

	input := []rune{'A'}
	m := re_vc.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B', 'A'}
	m = re_vc.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'E', '9'}
	m = re_vc.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'E', 'T'}
	m = re_vc.FullMatch(input)
	c.Check(m.Success, Equals, true)
}

// Test glob *
func (s *MySuite) TestGlob01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_vg, err := compiler.Compile("[:vowel:]*")
	c.Assert(err, IsNil)

	input := []rune{'B'}
	m := re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A', 'B'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Test glob +
func (s *MySuite) TestGlob02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_vg, err := compiler.Compile("[:vowel:]+")
	c.Assert(err, IsNil)

	input := []rune{'B'}
	m := re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A', 'B'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Test glob ?
func (s *MySuite) TestGlob03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_vg, err := compiler.Compile("[:vowel:]?")
	c.Assert(err, IsNil)

	input := []rune{'B'}
	m := re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	// TODO - this needs to change to true + consumed
	input = []rune{'A', 'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Test paren with no glob
func (s *MySuite) TestParen01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.RegisterClass(DigitClass)
	compiler.Finalize()

	re_vg, err := compiler.Compile("[:vowel:] ([:digit:][:digit:])")
	c.Assert(err, IsNil)

	input := []rune{'B'}
	m := re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A', '9'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A', '9', '8'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)
}

// Test paren with glob
func (s *MySuite) TestParen02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.RegisterClass(DigitClass)
	compiler.Finalize()

	re_vg, err := compiler.Compile("[:vowel:] ([:digit:][:vowel:])?")
	c.Assert(err, IsNil)

	input := []rune{'B'}
	m := re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', '9'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A', '9', '8'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A', '9', '9'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Test the "any" meta character (".")
func (s *MySuite) TestMetaAny(c *C) {
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

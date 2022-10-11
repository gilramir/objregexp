package objregexp

import (
	"fmt"

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
				fmt.Printf("returning true\n")
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
func (s *MySuite) TestOClass01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_vowel, err := compiler.Compile("[:vowel:]")
	c.Assert(err, IsNil)

	input := []rune{'A'}
	m := re_vowel.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'B'}
	m = re_vowel.Match(input)
	c.Check(m.Success, Equals, false)

	// Check that state is re-set properly and that
	// a match can happen again.
	input = []rune{'A'}
	m = re_vowel.Match(input)
	c.Check(m.Success, Equals, true)
}

// Test negation
func (s *MySuite) TestOClass02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.Finalize()

	re_not_vowel, err := compiler.Compile("[!:vowel:]")
	c.Assert(err, IsNil)

	input := []rune{'A'}
	m := re_not_vowel.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B'}
	m = re_not_vowel.Match(input)
	c.Check(m.Success, Equals, true)

	// Check that state is re-set properly and that
	// a failed-match can happen again.
	input = []rune{'A'}
	m = re_not_vowel.Match(input)
	c.Check(m.Success, Equals, false)
}

// Test two-classes in sequence
func (s *MySuite) TestOClass03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(VowelClass)
	compiler.RegisterClass(ConsonantClass)
	compiler.Finalize()

	re_vc, err := compiler.Compile("[:vowel:] [:consonant:]")
	c.Assert(err, IsNil)

	input := []rune{'A'}
	m := re_vc.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B', 'A'}
	m = re_vc.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'E', '9'}
	m = re_vc.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'E', 'T'}
	m = re_vc.Match(input)
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
	m := re_vg.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A'}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A'}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A', 'B'}
	m = re_vg.Match(input)
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
	m := re_vg.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A'}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A'}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', 'A', 'B'}
	m = re_vg.Match(input)
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
	m := re_vg.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A'}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, true)

	// TODO - this needs to change to true + consumed
	input = []rune{'A', 'A'}
	m = re_vg.Match(input)
	c.Check(m.Success, Equals, false)
}

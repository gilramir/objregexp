// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	. "gopkg.in/check.v1"
)

var vowels = []rune{'A', 'E', 'I', 'O', 'U', 'a', 'e', 'i', 'o', 'u'}
var consonants = []rune{'B', 'C', 'D', 'F', 'G', 'H', 'J', 'K',
	'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z',
	'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r',
	's', 't', 'v', 'w', 'x', 'y', 'z'}
var digits = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

var uppers = []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K',
	'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}

var lowers = []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k',
	'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

var UpperClass = &Class[rune]{
	"upper",
	func(r rune) bool {
		for _, t := range uppers {
			if r == t {
				return true
			}
		}
		return false
	},
}

var LowerClass = &Class[rune]{
	"lower",
	func(r rune) bool {
		for _, t := range lowers {
			if r == t {
				return true
			}
		}
		return false
	},
}

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

// SImple one-class test
func (s *MySuite) TestClass01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.AddClass(VowelClass)
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

	compiler.AddClass(VowelClass)
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

	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
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

	compiler.AddClass(VowelClass)
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

	compiler.AddClass(VowelClass)
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

	compiler.AddClass(VowelClass)
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

	input = []rune{'A', 'A'}
	m = re_vg.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Test glob ? grediness
func (s *MySuite) TestGlob04(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.Finalize()

	re, err := compiler.Compile("[:consonant:][:vowel:]?")
	c.Assert(err, IsNil)

	input := []rune{'B'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 1)

	input = []rune{}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'B', 'A'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Range.Start, Equals, 0)
	c.Check(m.Range.End, Equals, 2)
}

// Test paren with no glob
func (s *MySuite) TestParen01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.AddClass(VowelClass)
	compiler.AddClass(DigitClass)
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

	compiler.AddClass(VowelClass)
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	re, err := compiler.Compile("[:vowel:] ([:digit:][:vowel:])?")
	c.Assert(err, IsNil)
	//err = re.WriteDot("TestParen02.dot")
	//c.Assert(err, IsNil)

	input := []rune{'B'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'A', '9'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A', '9', '8'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'A', '9', '9'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Test the "any" meta character (".")
func (s *MySuite) TestMetaAny(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.AddClass(DigitClass)
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

func (s *MySuite) TestIdentity01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.AddClass(DigitClass)
	compiler.AddIdentity("lower x", 'x')
	compiler.Finalize()

	re_vowel, err := compiler.Compile("[:digit:] [:lower x:] [:digit:]")
	c.Assert(err, IsNil)

	input := []rune{'9', 'X', '1'}
	m := re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'9', 'x', '1'}
	m = re_vowel.FullMatch(input)
	c.Check(m.Success, Equals, true)
}

func (s *MySuite) TestDynClass01(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddClass(VowelClass)
	compiler.AddClass(UpperClass)
	compiler.AddClass(LowerClass)
	compiler.Finalize()

	text := "[:digit: || ((:consonant: && :lower:) || (:vowel: && :upper:))]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'e'}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'E'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'m'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'M'}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)
}

func (s *MySuite) TestDynClass02(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddClass(VowelClass)
	compiler.AddClass(UpperClass)
	compiler.AddClass(LowerClass)
	compiler.Finalize()

	text := "[:digit: || ((:consonant: && :lower:) || (:vowel: && :upper:))]+"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9', 'E', 'm'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'e'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'9', 'E', 'M'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'9', 'e', 'M'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

func (s *MySuite) TestNamedGroup01(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddClass(LowerClass)
	compiler.Finalize()

	text := "[:digit:] (?P<con>[:consonant: && :lower:])"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9', 'm'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	rs := string(input[m.Group(1).Start:m.Group(1).End])
	c.Check(rs, Equals, "m")
	rs = string(input[m.GroupName("con").Start:m.GroupName("con").End])
	c.Check(rs, Equals, "m")

	c.Check(m.GroupName("none").Start, Equals, -1)
	c.Check(m.GroupName("none").End, Equals, -1)
}

func (s *MySuite) TestNamedGroup02(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddClass(LowerClass)
	compiler.Finalize()

	text := "(?P<all> ([:digit:]) (?P<con>[:consonant: && :lower:]))"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9', 'm'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	/*
		dlog.Printf("g1: %+v\n", m.Group(1))
		dlog.Printf("gAll: %+v\n", m.GroupName("all"))
		dlog.Printf("g2: %+v\n", m.Group(2))
		dlog.Printf("g3: %+v\n", m.Group(3))
		dlog.Printf("gCon: %+v\n", m.GroupName("con"))
	*/
	rs := string(input[m.Group(1).Start:m.Group(1).End])
	c.Check(rs, Equals, "9m")
	rs = string(input[m.Group(2).Start:m.Group(2).End])
	c.Check(rs, Equals, "9")
	rs = string(input[m.Group(3).Start:m.Group(3).End])
	c.Check(rs, Equals, "m")

	rs = string(input[m.GroupName("all").Start:m.GroupName("all").End])
	c.Check(rs, Equals, "9m")

	rs = string(input[m.GroupName("con").Start:m.GroupName("con").End])
	c.Check(rs, Equals, "m")
}

func (s *MySuite) TestAssertBegin01(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "^[:digit:]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
}

func (s *MySuite) TestAssertBegin02(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(DigitClass)
	compiler.AddClass(LowerClass)
	compiler.Finalize()

	text := "(^|[:digit:]) [:lower:]"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9'}
	m := re.Match(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'9', 'm'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Group(1).Start, Equals, 0)
	c.Check(m.Group(1).End, Equals, 1)

	input = []rune{'m'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Group(1).Start, Equals, -1)
	c.Check(m.Group(1).End, Equals, -1)
}

func (s *MySuite) TestAssertEnd01(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(DigitClass)
	compiler.Finalize()

	text := "[:digit:]$"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
}

func (s *MySuite) TestAssertEnd02(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(DigitClass)
	compiler.AddClass(LowerClass)
	compiler.Finalize()

	text := "[:digit:] ([:lower:]|$)"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)

	input := []rune{'9'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Group(1).Start, Equals, -1)
	c.Check(m.Group(1).End, Equals, -1)

	input = []rune{'9', 'm'}
	m = re.Match(input)
	c.Check(m.Success, Equals, true)
	c.Check(m.Group(1).Start, Equals, 1)
	c.Check(m.Group(1).End, Equals, 2)

	input = []rune{'m'}
	m = re.Match(input)
	c.Check(m.Success, Equals, false)
}

package objregexp

// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

import (
	"fmt"

	. "gopkg.in/check.v1"
)

// Found this during development of paasaathai module
// The open paren at position 0 caused a crash
func (s *MySuite) TestNfa01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddIdentity("e", 'e')
	compiler.AddIdentity("ae", 'A')
	compiler.AddIdentity("o", 'o')
	compiler.AddIdentity("a", 'a')
	compiler.Finalize()

	text := "([:e:] | [:ae:] | [:o:])? " +
		"([:consonant:] [:consonant:]?) " +
		"([:a:])"

	_, err := compiler.Compile(text)
	c.Assert(err, IsNil)
}

// Found this during development of paasaathai module
// Register 2 had End == -1, later
// Register 2 had Start == -1
func (s *MySuite) TestNfa02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddIdentity("o", 'o')
	compiler.AddIdentity("a", 'a')
	compiler.Finalize()

	// Enter Regex: (c)(d?a|o)
	// postfix: cd?a.o|.
	text := "([:consonant:]) ([:digit:]? " +
		"[:a:] | [:o:])"

	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)
	//err = re.WriteDot("TestNfa02.dot")
	//c.Assert(err, IsNil)

	input := []rune{'B', 'a'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'B', 'o'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'B', '2', 'a'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	// The alternation (|) takes precedence over the glob (?)
	input = []rune{'B', '2', 'o'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)

	input = []rune{'2', '1', '0'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)
}

// Found this during development of paasaathai module
func (s *MySuite) TestNfa03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddIdentity("o", 'o')
	compiler.AddIdentity("a", 'a')
	compiler.Finalize()

	text := "([:consonant:]) ([:digit:])? " +
		"([:a:] | [:o:])"

	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)
	//err = re.WriteDot("TestNfa03.dot")
	//c.Assert(err, IsNil)

	input := []rune{'B', 'a'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'B', 'o'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'B', '2', 'a'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'B', '2', 'o'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	input = []rune{'2', '1', '0'}
	m = re.FullMatch(input)
	c.Check(m.Success, Equals, false)

}

// Found this during development of paasaathai module
func (s *MySuite) TestNfa04(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddIdentity("o", 'o')
	compiler.AddIdentity("e", 'e')
	compiler.AddIdentity("a", 'a')
	compiler.AddIdentity("ae", 'A')
	compiler.Finalize()

	text := "([:e:] | [:ae:] | [:o:])? " +
		"([:consonant:]) ([:digit:])? " +
		"([:a:])"

	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)
	//err = re.WriteDot("TestNfa04.dot")
	//c.Assert(err, IsNil)

	input := []rune{'B', 'a'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	reg1 := m.Register(1)
	fmt.Printf("reg1: %+v\n", reg1)
	c.Assert(reg1.Empty(), Equals, true)
}

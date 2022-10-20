package objregexp

// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

import (
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
	c.Assert(reg1.Empty(), Equals, true)
}

// Found this during development of paasaathai module
func (s *MySuite) TestNfa05(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddIdentity("o", 'o')
	compiler.AddIdentity("e", 'e')
	compiler.AddIdentity("y", 'y')
	compiler.AddIdentity("aa", 'A')
	compiler.Finalize()

	text := "([:e:]) " +
		"([:consonant:]) ([:digit:])?" +
		"([:aa:] | [:o:] | [:y:])"

	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)
	//	err = re.WriteDot("TestNfa05.dot")
	//	c.Assert(err, IsNil)

	input := []rune{'e', 'B', '9', 'A'}
	m := re.FullMatch(input)
	c.Check(m.Success, Equals, true)

	reg1 := m.Register(1)
	c.Assert(reg1.Start, Equals, 0)
	c.Assert(reg1.End, Equals, 1)

	reg2 := m.Register(2)
	c.Assert(reg2.Start, Equals, 1)
	c.Assert(reg2.End, Equals, 2)

	reg3 := m.Register(3)
	c.Assert(reg3.Start, Equals, 2)
	c.Assert(reg3.End, Equals, 3)

	reg4 := m.Register(4)
	c.Assert(reg4.Start, Equals, 3)
	c.Assert(reg4.End, Equals, 4)
}

// ( is only allowed at pos 0
// ) is a postfix which adds SR to stack0 start, and
// ER is added to frag so outs can be tagged

func (s *MySuite) TestNfa06(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddIdentity("o", 'o')
	compiler.AddIdentity("e", 'e')
	compiler.AddIdentity("y", 'y')
	compiler.AddIdentity("aa", 'A')
	compiler.Finalize()

	//text := "([:e:]?)"
	//text := "([:e:])?"
	//text := "([:o:])?([:e:]|[:y:])"
	text := "([:o:]?)([:e:]|[:y:])"
	//text := "([:o:])*([:e:]|[:y:])"

	//text := "([:y:]?[:aa:]) ([:e:]* | [:o:])"
	//text := "([:y:]?[:aa:]) ([:e:]* | [:o:])"
	re, err := compiler.Compile(text)
	c.Assert(err, IsNil)
	//	err = re.WriteDot("TestNfa06.dot")
	//	c.Assert(err, IsNil)

	input := []rune{'o', 'e', 'y'}
	m := re.Match(input)
	c.Check(m.Success, Equals, true)

	reg1 := m.Register(1)
	c.Assert(reg1.Start, Equals, 0)
	c.Assert(reg1.End, Equals, 1)

	reg2 := m.Register(2)
	c.Assert(reg2.Start, Equals, 1)
	c.Assert(reg2.End, Equals, 2)

	/*
		input := []rune{'e', 'B', '9', 'A'}
		m := re.FullMatch(input)
		c.Check(m.Success, Equals, true)
			reg1 := m.Register(1)
			c.Assert(reg1.Start, Equals, 0)
			c.Assert(reg1.End, Equals, 1)

			reg2 := m.Register(2)
			c.Assert(reg2.Start, Equals, 1)
			c.Assert(reg2.End, Equals, 2)

			reg3 := m.Register(3)
			c.Assert(reg3.Start, Equals, 2)
			c.Assert(reg3.End, Equals, 3)

			reg4 := m.Register(4)
			c.Assert(reg4.Start, Equals, 3)
			c.Assert(reg4.End, Equals, 4)
	*/
}

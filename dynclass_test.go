package objregexp

// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestDynParse01(c *C) {
	var parser dcParserStateT[rune]

	text := ":foo:"
	parser.Initialize(text)
	tokens, err := parser.parse()
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 1)

	c.Check(tokens[0].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[0].name, Equals, "foo")
}

func (s *MySuite) TestDynParse02(c *C) {
	var parser dcParserStateT[rune]

	text := "! :foo:"
	parser.Initialize(text)
	tokens, err := parser.parse()
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, dcTokenTypeT("!"))
}

func (s *MySuite) TestDynParse03(c *C) {
	var parser dcParserStateT[rune]

	text := "! ! :foo:"
	parser.Initialize(text)
	tokens, err := parser.parse()
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 3)

	c.Check(tokens[0].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, dcTokenTypeT("!"))
	c.Check(tokens[2].ttype, Equals, dcTokenTypeT("!"))
}

func (s *MySuite) TestDynParse04(c *C) {
	var parser dcParserStateT[rune]

	text := "! ! :foo:"
	parser.Initialize(text)
	tokens, err := parser.parse()
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 3)

	c.Check(tokens[0].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, dcTokenTypeT("!"))
	c.Check(tokens[2].ttype, Equals, dcTokenTypeT("!"))
}

func (s *MySuite) TestDynParse05(c *C) {
	var parser dcParserStateT[rune]

	text := "! :foo: && :bar: "
	parser.Initialize(text)
	tokens, err := parser.parse()
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 5)

	c.Check(tokens[0].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, dcTokenTypeT("!"))

	c.Check(tokens[2].ttype, Equals, dcTokenTypeT("F"))
	c.Check(tokens[2].jmpTarget, Equals, 1)

	c.Check(tokens[3].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[3].name, Equals, "bar")

	c.Check(tokens[4].ttype, Equals, dcTokenTypeT("?"))
	c.Check(tokens[4].jmpTarget, Equals, 1)
}

func (s *MySuite) TestDynParse06(c *C) {
	var parser dcParserStateT[rune]

	text := ":foo: && ( :bar: || :baz: )"
	parser.Initialize(text)
	tokens, err := parser.parse()
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 7)

	c.Check(tokens[0].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, dcTokenTypeT("F"))
	c.Check(tokens[1].jmpTarget, Equals, 1)

	c.Check(tokens[2].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[2].name, Equals, "bar")

	c.Check(tokens[3].ttype, Equals, dcTokenTypeT("T"))
	c.Check(tokens[3].jmpTarget, Equals, 2)

	c.Check(tokens[4].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[4].name, Equals, "baz")

	c.Check(tokens[5].ttype, Equals, dcTokenTypeT("?"))
	c.Check(tokens[5].jmpTarget, Equals, 2)

	c.Check(tokens[6].ttype, Equals, dcTokenTypeT("?"))
	c.Check(tokens[6].jmpTarget, Equals, 1)
}

func (s *MySuite) TestDynParse07(c *C) {
	var parser dcParserStateT[rune]

	text := "( (!:foo:) || ( :bar: && :baz: && :a:) )"
	parser.Initialize(text)
	tokens, err := parser.parse()
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 11)

	c.Check(tokens[0].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, dcTokenTypeT("!"))

	c.Check(tokens[2].ttype, Equals, dcTokenTypeT("T"))
	c.Check(tokens[2].jmpTarget, Equals, 1)

	c.Check(tokens[3].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[3].name, Equals, "bar")

	c.Check(tokens[4].ttype, Equals, dcTokenTypeT("F"))
	c.Check(tokens[4].jmpTarget, Equals, 2)

	c.Check(tokens[5].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[5].name, Equals, "baz")

	c.Check(tokens[6].ttype, Equals, dcTokenTypeT("F"))
	c.Check(tokens[6].jmpTarget, Equals, 3)

	c.Check(tokens[7].ttype, Equals, dcTokenTypeT("C"))
	c.Check(tokens[7].name, Equals, "a")

	c.Check(tokens[8].ttype, Equals, dcTokenTypeT("?"))
	c.Check(tokens[8].jmpTarget, Equals, 3)

	c.Check(tokens[9].ttype, Equals, dcTokenTypeT("?"))
	c.Check(tokens[9].jmpTarget, Equals, 2)

	c.Check(tokens[10].ttype, Equals, dcTokenTypeT("?"))
	c.Check(tokens[10].jmpTarget, Equals, 1)
}

func (s *MySuite) TestDynMatch01(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddIdentity("a", 'a')
	compiler.Finalize()

	text := ":a: && :consonant:"
	dynClass, err := newDynClassT[rune](text, &compiler)
	c.Assert(err, IsNil)

	for i, op := range dynClass.ops {
		dlog.Printf("op #%d: %+v", i, op)
	}

	m := dynClass.Matches('a')
	c.Check(m, Equals, false)

	m = dynClass.Matches('B')
	c.Check(m, Equals, false)
}

func (s *MySuite) TestDynMatch02(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(LowerClass)
	compiler.AddIdentity("e", 'e')
	compiler.Finalize()

	text := ":lower: && (:consonant: || :e:)"
	dynClass, err := newDynClassT[rune](text, &compiler)
	c.Assert(err, IsNil)

	for i, op := range dynClass.ops {
		dlog.Printf("op #%d: %+v", i, op)
	}

	m := dynClass.Matches('a')
	c.Check(m, Equals, false)

	m = dynClass.Matches('e')
	c.Check(m, Equals, true)

	m = dynClass.Matches('m')
	c.Check(m, Equals, true)
}

func (s *MySuite) TestDynMatch03(c *C) {

	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(ConsonantClass)
	compiler.AddClass(DigitClass)
	compiler.AddClass(VowelClass)
	compiler.AddClass(UpperClass)
	compiler.AddClass(LowerClass)
	compiler.Finalize()

	text := ":digit: || ((:consonant: && :lower:) || (:vowel: && :upper:))"
	dynClass, err := newDynClassT[rune](text, &compiler)
	c.Assert(err, IsNil)

	for i, op := range dynClass.ops {
		dlog.Printf("op #%d: %+v", i, op)
	}

	m := dynClass.Matches('9')
	c.Check(m, Equals, true)

	m = dynClass.Matches('e')
	c.Check(m, Equals, false)

	m = dynClass.Matches('E')
	c.Check(m, Equals, true)

	m = dynClass.Matches('m')
	c.Check(m, Equals, true)

	m = dynClass.Matches('M')
	c.Check(m, Equals, false)
}

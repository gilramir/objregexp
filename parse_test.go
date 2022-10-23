package objregexp

// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>
import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParser01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:foo:]"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 1)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")
}

func (s *MySuite) TestParser02(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:foo:"
	_, err := parseRegex[rune](text, &compiler)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Unexpected end of string")
}

func (s *MySuite) TestParser03(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:foo"
	_, err := parseRegex[rune](text, &compiler)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Unexpected end of string")
}

func (s *MySuite) TestParser04a(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[: spaces are legal :]"
	_, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)
}

func (s *MySuite) TestParser04b(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:방탄소년단:]"
	_, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)
}

func (s *MySuite) TestParser05(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:foo:][:bar:]*"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 4)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[1].name, Equals, "bar")

	c.Check(tokens[2].ttype, Equals, tokenTypeT(tGlobStar))
	c.Check(tokens[3].ttype, Equals, tokenTypeT(tConcat))
}

func (s *MySuite) TestParser06(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:foo:]*"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tGlobStar))
}

func (s *MySuite) TestParser07(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:foo:]+"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tGlobPlus))
}

func (s *MySuite) TestParser08(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.Finalize()

	text := "[:foo:]?"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tGlobQuestion))
}

func (s *MySuite) TestParser09(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.AddIdentity("alpha", 'A')
	compiler.AddIdentity("bar", 'b')
	compiler.Finalize()

	text := "[:foo:] ([:alpha:][:bar:])"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	printTokens(tokens)
	c.Assert(len(tokens), Equals, 6)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[1].name, Equals, "alpha")

	c.Check(tokens[2].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[2].name, Equals, "bar")

	c.Check(tokens[3].ttype, Equals, tokenTypeT(tConcat))
	c.Check(tokens[4].ttype, Equals, tokenTypeT(tEndRegister))
	c.Check(tokens[5].ttype, Equals, tokenTypeT(tConcat))
}

func (s *MySuite) TestParser10(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.AddIdentity("bar", 'b')
	compiler.Finalize()

	text := "[:foo:] | [:bar:]"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 3)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[1].name, Equals, "bar")

	c.Check(tokens[2].ttype, Equals, tokenTypeT(tAlternate))
}

func (s *MySuite) TestParser11(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.AddIdentity("baz", 'z')
	compiler.AddIdentity("bar", 'b')
	compiler.Finalize()

	text := "[:foo:] | [:bar:][:bar:]"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 5)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[1].name, Equals, "bar")

	c.Check(tokens[2].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[2].name, Equals, "bar")

	c.Check(tokens[3].ttype, Equals, tokenTypeT(tConcat))
	c.Check(tokens[4].ttype, Equals, tokenTypeT(tAlternate))
}

func (s *MySuite) TestParser12(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.AddIdentity("baz", 'z')
	compiler.AddIdentity("bar", 'b')
	compiler.Finalize()

	text := "[:foo:] | [:bar:] | [:baz:]"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 5)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[1].name, Equals, "bar")

	c.Check(tokens[2].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[2].name, Equals, "baz")

	c.Check(tokens[3].ttype, Equals, tokenTypeT(tAlternate))
	c.Check(tokens[4].ttype, Equals, tokenTypeT(tAlternate))
}

func (s *MySuite) TestParser13(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("foo", 'a')
	compiler.AddIdentity("bar", 'b')
	compiler.Finalize()

	text := "( [:foo:] | [:bar:] )"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 4)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[1].name, Equals, "bar")

	c.Check(tokens[2].ttype, Equals, tokenTypeT(tAlternate))

	c.Check(tokens[3].ttype, Equals, tokenTypeT(tEndRegister))
}

// Nested parens
func (s *MySuite) TestParser14(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("a", 'a')
	compiler.AddIdentity("b", 'b')
	compiler.AddIdentity("c", 'c')
	compiler.AddIdentity("d", 'd')
	compiler.Finalize()

	text := "[:a:] ( [:b:] ( [:c:] | [:d:] ) )?"
	_, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	//printTokens(tokens)
	//  abcd|.?.
}

// Imbalanced nested parens which at one point caused an
// infinite loop in the logic
func (s *MySuite) TestParser15(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("a", 'a')
	compiler.AddIdentity("b", 'b')
	compiler.AddIdentity("c", 'c')
	compiler.AddIdentity("d", 'd')
	compiler.Finalize()

	text := "[:a:] ( [:b:] ( [:c:] | [:d:] )?"
	_, err := parseRegex[rune](text, &compiler)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Unexpected end of string")
}

// Test the handling of registers in the parser
func (s *MySuite) TestParserRegs01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("aa", 'A')
	compiler.AddIdentity("e", 'e')
	compiler.AddIdentity("c", 'c')
	compiler.AddIdentity("d", 'd')
	compiler.AddIdentity("o", 'o')
	compiler.AddIdentity("y", 'y')
	compiler.Finalize()

	text := "([:e:]) ([:c:]) ([:d:])? ([:aa:] | [:o:] | [:y:])"
	tokens, err := parseRegex[rune](text, &compiler)
	c.Assert(err, IsNil)

	tokenString := makeTokensString(tokens)
	dlog.Printf("tokenString: %s", tokenString)
	c.Assert(tokenString, Equals, "C)C).C)?.CCC||).")
}

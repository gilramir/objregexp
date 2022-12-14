package objregexp

// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParser01(c *C) {
	text := "[:foo:]"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 1)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")
}

func (s *MySuite) TestParser02(c *C) {
	text := "[:foo:"
	_, err := parseRegex(text)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Unexpected end of string")
}

func (s *MySuite) TestParser03(c *C) {
	text := "[:foo"
	_, err := parseRegex(text)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals, "Unexpected end of string")
}

func (s *MySuite) TestParser04a(c *C) {
	text := "[: spaces are legal :]"
	_, err := parseRegex(text)
	c.Assert(err, IsNil)
}

func (s *MySuite) TestParser04b(c *C) {
	text := "[:방탄소년단:]"
	_, err := parseRegex(text)
	c.Assert(err, IsNil)
}

func (s *MySuite) TestParser05(c *C) {
	text := "[:foo:][:bar:]*"
	tokens, err := parseRegex(text)
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
	text := "[:foo:]*"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tGlobStar))
}

func (s *MySuite) TestParser07(c *C) {
	text := "[:foo:]+"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tGlobPlus))
}

func (s *MySuite) TestParser08(c *C) {
	text := "[:foo:]?"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tGlobQuestion))
}

func (s *MySuite) TestParser09(c *C) {
	text := "[:foo:] ([:alpha:][:bar:])"
	tokens, err := parseRegex(text)
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
	text := "[:foo:] | [:bar:]"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 3)

	c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
	c.Check(tokens[1].name, Equals, "bar")

	c.Check(tokens[2].ttype, Equals, tokenTypeT(tAlternate))
}

func (s *MySuite) TestParser11(c *C) {
	text := "[:foo:] | [:bar:][:bar:]"
	tokens, err := parseRegex(text)
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
	text := "[:foo:] | [:bar:] | [:baz:]"
	tokens, err := parseRegex(text)
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
	text := "( [:foo:] | [:bar:] )"
	tokens, err := parseRegex(text)
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
	// infloop text := "[:a:] ( [:b:] ( [:c:] | [:d:] )?"
	text := "[:a:] ( [:b:] ( [:c:] | [:d:] ) )?"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	printTokens(tokens)
	//  abcd|.?.
	/*
		c.Assert(len(tokens), Equals, 5)

		c.Check(tokens[0].ttype, Equals, tokenTypeT(tClass))
		c.Check(tokens[0].name, Equals, "foo")

		c.Check(tokens[1].ttype, Equals, tokenTypeT(tClass))
		c.Check(tokens[1].name, Equals, "alpha")

		c.Check(tokens[2].ttype, Equals, tokenTypeT(tClass))
		c.Check(tokens[2].name, Equals, "bar")

		c.Check(tokens[3].ttype, Equals, tokenTypeT(tConcat))
		c.Check(tokens[4].ttype, Equals, tokenTypeT(tConcat))
	*/
}

// Imbalanced nested parens which at one point caused an
// infinite loop in the logic
func (s *MySuite) TestParser15(c *C) {
	text := "[:a:] ( [:b:] ( [:c:] | [:d:] )?"
	_, err := parseRegex(text)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Unexpected end of string")
}

// Test the handling of registers in the parser
func (s *MySuite) TestParserRegs01(c *C) {
	text := "([:e:]) ([:c:]) ([:d:])? ([:aa:] | [:o:] | [:y:])"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	tokenString := makeTokensString(tokens)
	dlog.Printf("tokenString: %s", tokenString)
	c.Assert(tokenString, Equals, "C)C).C)?.CCC||).")
}

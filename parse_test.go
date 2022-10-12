package objregexp

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParser01(c *C) {
	text := "[:foo:]"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 1)

	c.Check(tokens[0].ttype, Equals, tokenType(tClass))
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

func (s *MySuite) TestParser04(c *C) {
	text := "[: space:]"
	_, err := parseRegex(text)
	c.Assert(err, NotNil)
	c.Check(err.Error(), Equals,
		"The class name starting at pos 2 has a space in it")
}

func (s *MySuite) TestParser05(c *C) {
	text := "[:foo:][:bar:]*"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 4)

	c.Check(tokens[0].ttype, Equals, tokenType(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenType(tClass))
	c.Check(tokens[1].name, Equals, "bar")

	c.Check(tokens[2].ttype, Equals, tokenType(tGlobStar))
	c.Check(tokens[3].ttype, Equals, tokenType(tConcat))
}

func (s *MySuite) TestParser06(c *C) {
	text := "[:foo:]*"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenType(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenType(tGlobStar))
}

func (s *MySuite) TestParser07(c *C) {
	text := "[:foo:]+"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenType(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenType(tGlobPlus))
}

func (s *MySuite) TestParser08(c *C) {
	text := "[:foo:]?"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 2)

	c.Check(tokens[0].ttype, Equals, tokenType(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenType(tGlobQuestion))
}

func (s *MySuite) TestParser09(c *C) {
	text := "[:foo:] ([:alpha:][:bar:])"
	tokens, err := parseRegex(text)
	c.Assert(err, IsNil)

	c.Assert(len(tokens), Equals, 5)

	c.Check(tokens[0].ttype, Equals, tokenType(tClass))
	c.Check(tokens[0].name, Equals, "foo")

	c.Check(tokens[1].ttype, Equals, tokenType(tClass))
	c.Check(tokens[1].name, Equals, "alpha")

	c.Check(tokens[2].ttype, Equals, tokenType(tClass))
	c.Check(tokens[2].name, Equals, "bar")

	c.Check(tokens[3].ttype, Equals, tokenType(tConcat))
	c.Check(tokens[4].ttype, Equals, tokenType(tConcat))
}

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

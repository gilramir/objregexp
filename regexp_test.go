// Copyright 2022 by Gilbert Ramirez <gram@alumni.rice.edu>

package objregexp

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestOnlyMatchesAtBeginning(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.Finalize()

	re, err := compiler.Compile("[:vowel:]")
	c.Assert(err, IsNil)
	c.Check(re.onlyMatchesAtBeginning(), Equals, false)

	re, err = compiler.Compile("(^|[:vowel:])")
	c.Assert(err, IsNil)
	c.Check(re.onlyMatchesAtBeginning(), Equals, false)

	re, err = compiler.Compile("([:vowel:]|^)")
	c.Assert(err, IsNil)
	c.Check(re.onlyMatchesAtBeginning(), Equals, false)

	re, err = compiler.Compile("^[:vowel:]")
	c.Assert(err, IsNil)
	c.Check(re.onlyMatchesAtBeginning(), Equals, true)

	re, err = compiler.Compile("(^[:vowel:])")
	c.Assert(err, IsNil)
	c.Check(re.onlyMatchesAtBeginning(), Equals, true)

	re, err = compiler.Compile("(^|^)[:vowel:]")
	c.Assert(err, IsNil)
	c.Check(re.onlyMatchesAtBeginning(), Equals, true)
}

func (s *MySuite) TestMustStartWith(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddClass(VowelClass)
	compiler.AddClass(ConsonantClass)
	compiler.AddIdentity("e", 'e')
	compiler.Finalize()

	re, err := compiler.Compile("[:vowel:]")
	c.Assert(err, IsNil)
	nfa := re.mustStartWith()
	c.Assert(nfa, NotNil)
	c.Check(nfa.c, Equals, ntClass)

	re, err = compiler.Compile("[:e:]")
	c.Assert(err, IsNil)
	nfa = re.mustStartWith()
	c.Assert(nfa, NotNil)
	c.Check(nfa.c, Equals, ntIdentity)

	re, err = compiler.Compile("[:e: || :vowel:]")
	c.Assert(err, IsNil)
	nfa = re.mustStartWith()
	c.Assert(nfa, NotNil)
	c.Check(nfa.c, Equals, ntDynClass)

	re, err = compiler.Compile("[:e:] | [:vowel:]")
	c.Assert(err, IsNil)
	c.Check(re.mustStartWith(), IsNil)

	re, err = compiler.Compile("[:e:]?")
	c.Assert(err, IsNil)
	c.Check(re.mustStartWith(), IsNil)

	re, err = compiler.Compile("[:e:]*")
	c.Assert(err, IsNil)
	c.Check(re.mustStartWith(), IsNil)

	// There must be at least one 'e', so this is not nil
	re, err = compiler.Compile("[:e:]+")
	c.Assert(err, IsNil)
	nfa = re.mustStartWith()
	c.Assert(nfa, NotNil)
	c.Check(nfa.c, Equals, ntIdentity)

	// AssertBegin doesn't qualify
	re, err = compiler.Compile("^[:e:]*")
	c.Assert(err, IsNil)
	c.Check(re.mustStartWith(), IsNil)
}

func (s *MySuite) TestCompileNestedGroupNames(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()
	compiler.AddIdentity("a", 'a')
	compiler.AddIdentity("b", 'b')
	compiler.AddIdentity("c", 'c')
	compiler.Finalize()

	rt := "(?P<xtra>[:a:]) ([:a:]) "
	_, err := compiler.Compile(rt)
	c.Assert(err, IsNil)
}

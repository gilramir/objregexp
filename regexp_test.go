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
	c.Check(re.OnlyMatchesAtBeginning(), Equals, false)

	re, err = compiler.Compile("(^|[:vowel:])")
	c.Assert(err, IsNil)
	c.Check(re.OnlyMatchesAtBeginning(), Equals, false)

	re, err = compiler.Compile("([:vowel:]|^)")
	c.Assert(err, IsNil)
	c.Check(re.OnlyMatchesAtBeginning(), Equals, false)

	re, err = compiler.Compile("^[:vowel:]")
	c.Assert(err, IsNil)
	c.Check(re.OnlyMatchesAtBeginning(), Equals, true)

	re, err = compiler.Compile("(^[:vowel:])")
	c.Assert(err, IsNil)
	c.Check(re.OnlyMatchesAtBeginning(), Equals, true)

	re, err = compiler.Compile("(^|^)[:vowel:]")
	c.Assert(err, IsNil)
	c.Check(re.OnlyMatchesAtBeginning(), Equals, true)
}

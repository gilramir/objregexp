package objregexp

import (
	"fmt"

	. "gopkg.in/check.v1"
)

var vowels = []rune{'A', 'E', 'I', 'O', 'U'}
var consonants = []rune{'B', 'C', 'D', 'F', 'G', 'H', 'J', 'K',
	'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z'}
var digits = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

var ClassVowel = &Class[rune]{
	"vowel",
	func(r rune) bool {
		fmt.Printf("matching %c for vowel\n", r)
		for _, t := range vowels {
			if r == t {
				fmt.Printf("returning true\n")
				return true
			}
		}
		fmt.Printf("returning false\n")
		return false
	},
}

var ClassConsonant = &Class[rune]{
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

/*
var ioMap = map[string]string{
	"A": "A",
	"B": "B",
	"C": "C",
}

func ioMapper(name string, inputTarget string) *TASTNode {
	values := make([]string, 1)
	values[0] = string(inputTarget)
	return &TASTNode{
		values:   values,
		consumed: 1,
	}
}

*/

func (s *MySuite) TestOClass01(c *C) {
	var compiler Compiler[rune]
	compiler.Initialize()

	compiler.RegisterClass(ClassVowel)
	compiler.Finalize()

	re_vowel, err := compiler.Compile("[:vowel:]")
	c.Assert(err, IsNil)
	fmt.Printf("re_vowel: %+v\n", re_vowel)

	fmt.Printf("Testing A\n")
	input := []rune{'A'}
	m := re_vowel.Match(input)
	c.Check(m.Success, Equals, true)

	fmt.Printf("Testing B\n")
	input = []rune{'B'}
	m = re_vowel.Match(input)
	c.Check(m.Success, Equals, false)
}

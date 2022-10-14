# Introduction
Regular expressions are convenient mechanisms for doing some types of parsing
of strings. But what if you aren't analyzing string?
"Object" regular expressions are a way to use regular
expressions on slices of arbitrary objects, instead of just strings.

This library uses Go generics.  The objects involved in the regular
expression must implement the "comparable" Golang contstraint. "comparable
in Go means that "==" and "!=" work with the type.  If it's a struct,
every field in the struct must the "comparable".

To make this work, you have to define the "vocabulary" of the regular
expression compiler. You create "object classes" (similar "character classes")
which will declare if an object belongs to the class.

Given the classes, you can write regular expressions using basic
regular expression syntax.

In this syntax:

* All class names are inside a pair of brackets and colons. For example,
        a class named "vowel" is used in a regular expression as "[:vowel:]"

* A class name can have any graphic (visible) Unicode character, or space,
        in it, except for ":". The name is not limited to ASCII or Latin
        code points.

* A "!" between the "[" and ":" before a class name negates the test.
        The test looks for non-membership in the class.

* A "." matches any one object.

* Parens can be used for grouping.

* Alternate choices are given via the vertical pipe: |

* These "repeat" indicators are supported: +, \*, ?

* Whitespace has no meaning and can be used liberally throughout
        your reggex to make it more readable.

# Examples

Here are some sample regexes, assuming that "vowel" and "consonant" are
object classes;

---
    # Match any object
    .

    # Match an object which is part of the "vowel" class.
    [:vowel:]

    # Match an object which is part of the "vowel" class.
    ([:vowel:])

    # Match an object which is part of the "vowel" class,
    # followed by an object which is part of the "consonant" class.
    [:vowel:] [:consonant:]

    # Match an object which is not part of the "vowel" class.
    [!:vowel:]

    # Match either a "vowel" object, or a "consonant" object
    # followed by a "vowel" object.
    ([:vowel:] | [:consonant:] [:vowel:])

    # Match one or more vowel objects
    [:vowel:]+

    # Match zero or more vowel objects
    [:vowel:]*

    # Match one or zero vowel objects
    ?
---

# Using objregexp

## Instantiate the Compiler

First instantiate a Compiler object. The Compiler object
will hold the object classes you define.

In this example, the object type we consider is a rune, but
it can be any object type which satisfies the "comparable" constraint.

---
    import (
            "github.com/gilramir/objregexp"
    )

    var vowels = []rune{'A', 'E', 'I', 'O', 'U'}

    var VowelClass = &Class[rune]{
            "vowel",
            func(r rune) bool {
                    // Is this rune in the list of vowel runes?
                    for _, t := range vowels {
                            if r == t {
                                    return true
                            }
                    }
                    return false
            },
    }


    func main() {
            rc := Compiler[MyType]()
            compiler.RegisterClass(VowelClass)
            compiler.Finalize()
    }
---


## Compile the Regexp

Once you have defined
all your object classes, use the Compiler object to compile
regular expressions into Regexp objects.

---
        pattern = "[:vowel:]*"
        regex, err := compiler.Compile(pattern)
---

## Use the Regexp on a slice of objects

---
        objects := []rune{ 'A', 'E', 'I' }

        m = regex.Match(objects)
        if m.Success {
            fmt.Println("Success!")
        }
---

The Regexp class has a few different methods for trying
the regular expression on a slice of objects:

* *Match* - Find the first sub-slice that matches
* *MatchAt* - Like above, but starting at a specific index
* *FullMatch* - Check if the regex match the entire slice
* *FullMatchAt* - Like above, but starting at a specific index
* *Search* - Find the first match, starting at any index
* *SearchAt* - Find the first match, starting at a specific index

## The Match object

The Match object not only tells you if the regex succesfully
matched, but it gives you the index position of the start and
end of the match.

If you used parens in your regular expressions,
you can get the start and end positions for each of these "registers",
based on their number, starting with 1.

For example:

---
        pattern = "[:vowel:] ([!:vowel:])"
        regex, err := compiler.Compile(pattern)

        objects := []rune{ 'A', 'X' }

        m = regex.Match(objects)
        if m.Success {
            fmt.Println("The non-vowel group is at pos %d - %d",
                m.Register(1).Start, m.Register(1).End)
        }
---

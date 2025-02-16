# Introduction

Regular expressions are convenient mechanisms for parsing
strings. But what if you aren't analyzing strings?
"Object regular expressions" are a way to use regular
expressions on slices of arbitrary objects instead of strings.

# Usage

This library uses Go generics.  The objects involved in the regular
must satisfy the "comparable" constraint.

To use this module, you have to define the "vocabulary" of the regular
expression compiler. You create "object classes" (similar "character classes").
One type of class is implemented by a function which declares if an object
belongs to the class.
Another type of class is an identity; a name is given to an object,
and if the input object compares to it equally, then it is a member of the
class.

Once you define the classes, you can write regular expressions using syntax
similar to basic regular expression syntax.

Objects do not have to belong to only one category. They can belong to as
many categories as you want. The regular expression can also check that
a single object matches (or doesn't match) more than one category.

See the [module documentation](https://pkg.go.dev/github.com/gilramir/objregexp)

# The Syntax

In this syntax:

* All class names are inside a pair of brackets and colons. For example,
        a class named "vowel" is used in a regular expression as "[:vowel:]"

* A class name can have any graphic (visible) Unicode character, or space,
        in it, except for ":". The name is not limited to ASCII or Latin
        code points.

* Inside the "[" and "]" brackets of a class name:
    * A "!" before the ":" of a class name
        negates the test.  It looks for non-membership in the class.
    * A "&&" between 2 class names tests that the object belongs
        to both classes.
    * A "||" between 2 class names tests that the object belongs
        to either class.
    * "(" and ")" can be used to group the "&&" and "||" tests.

* A "." matches any one object.

* "^" matches the beginning of the input.

* "$" matches the end of the input.

* Parens can be used for grouping, both for the globs and for retrieving
  the range of objects after the match is successful.

* Capture groups can be named, using the same syntax that Python regexes use:
        (?P<name>.*)

* Alternate choices are given via the vertical pipe: |

* These "glob" patterns are supported: "+", "\*", and "?". They are greedy;
  they will match as many objects as they can.

* Whitespace has no meaning and can be used liberally throughout
        your reggex to make it more readable.

# Examples


Here are some sample regexes, assuming that "vowel", "consonant",
"lower" and "upper" are object classes.

```
    # Match any object
    .

    # Match an object which is part of the "vowel" class.
    [:vowel:]

    # Match an object which is part of the "vowel" class.
    # Save the value in a numbered "capture group"
    ([:vowel:])

    # Match an object which is part of the "vowel" class,
    # followed by an object which is part of the "consonant" class.
    [:vowel:] [:consonant:]

    # Match an object which is not part of the "vowel" class.
    [!:vowel:]

    # Match either a "vowel" object, or a "consonant" object
    # followed by a "vowel" object.
    # Save the value in a numbered "capture group"
    ([:vowel:] | [:consonant:] [:vowel:])

    # Name the matched group "foo"
    (?P<foo>[:vowel:] | [:consonant:] [:vowel:])

    # Match one or more vowel objects
    [:vowel:]+

    # Match zero or more vowel objects
    [:vowel:]*

    # Match one or zero vowel objects
    [:vowel:]?
```

You can test a single object against multiple classes, too.

```
    # Match an object which is part of the "vowel" class, and also
    # part of the "lower" class.
    [:vowel: && :lower:]

    # Match an object which is either a lower-case vowel, or
    # an upper-case consonant.
    [ (:vowel: && :lower:) || (:consonant: && :upper:) ]
```


# Using objregexp

## Instantiate the Compiler

First instantiate a Compiler object. The Compiler object
will hold the object classes you define.

In this example, the object type we consider is a rune, but
it can be any object type which satisfies the "comparable" constraint.

```
    import (
            "github.com/gilramir/objregexp"
    )

    var vowels = []rune{'A', 'E', 'I', 'O', 'U'}

    // A class is comprised of a name and a function
    // which accepts one argument of the type you used
    // to instantiate the Compiler, and returns a boolean.
    // The boolean indicates if the item is a member of
    // the class.
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
            rc := Compiler[rune]()

            // Add a class function
            compiler.AddClass(VowelClass)

            // Add an identity
            compiler.AddIdentity("lower x", 'x')

            // The compiler must be "finalized" before
            // it can compile regexes
            compiler.Finalize()
    }
```


## Compile the Regexp

Once you have defined
all your object classes, use the Compiler object to compile
regular expressions into Regexp objects.

```
        pattern = "[:vowel:]* [:lower x:]"
        regex, err := compiler.Compile(pattern)
```

## Use the Regexp on a slice of objects

```
        objects := []rune{ 'A', 'E', 'I', 'x' }

        m = regex.Match(objects)
        if m.Success {
            fmt.Println("Success!")
        }
```

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
you can get the start and end positions for each of these groups,
based on their number, starting with 1.

For example:

```
        pattern = "[:vowel:] ([!:vowel:])"
        regex, err := compiler.Compile(pattern)

        objects := []rune{ 'A', 'X' }

        m = regex.Match(objects)
        if m.Success {
            vSpan := m.Group(1)
            fmt.Println("The non-vowel group is at pos %d - %d",
                vSpan.Start, vSpan.End)
        }
```

# Concurrency

The regexes on this module are concurrent-safe. The same Regexp
object can be used to Match() (with any of the related methods)
in at the same time in different concurrent goroutines. Each
Match uses its own state; there is no internal locking of the
Regexp object.


# Internals

The code was inspired by
[Regular Expression Matching Can Be Simple and Fast](https://swtch.com/~rsc/regexp/regexp1.html)

Files:

* class.go - this defines the struct for Class
* dynclass.go - code for dynamically combining classes with boolean logic
* nfa.go - this generates the NFA (non-deterministic finite automata)
* objregexp.go - this defines the Compiler and its methods
* parse.go - this tokenizes the regex string
* regexec.go - this executes the regex
* regexp.go - this defines the Regexp class and its methods
* runebuffer.go - simple buffer of runes used by the parsers in parse.go and
  dynclass.go
* stack.go - generic stack implementation

Flow:

1.  When the Compiler is used to compile a regex, the regex string
is tokenzied by code in parse.go.

2. The tokens are then analyzed to produce an NFA, in nfa.go.

3. That NFA is inserted into a Regexp object, returned to the caller.

4. When the Regexp object is used to match a sequence, an executorT
object is created in regexec.go. That executorT object carries
the state used while traversing the sequence of objects.

# Bugs

None that are known.

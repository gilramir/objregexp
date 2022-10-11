Regular expressions are convenient mechanisms for doing some types of parsing
of strings. "Object" regular expressions are a way to use the regular
expressions on slices of arbitrary objects, instead of strings.


    [:vowel:]
    (?P<V>[:vowel:])
    [:vowel:][:consonant:]
    [!:vowel:]
    (?P<char>[:vowel:] [!:consonant:])
    ([:vowel:] [:consonant:])
    ([:vowel:] | [:consonant:])
    {2}
    {1,2}
    +
    *
    ?

Whitespace can be used liberally.

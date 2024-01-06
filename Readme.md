# Calc

A simple calculator language / REPL. This project is merely for code practicing, not intended to be used.

    1+2
    >  3
    a = 2 * ( 1+1)
    >  4
    a+a
    >  8

Supported features:

 - integer and floating point literals
 - variables
 - 4 arithmetic operations +, -, *, /
 - explicit evaluation order by parenthesis

All arithmetic operators are left associative thus following the natural notations. 1-2+1 is 0 and not -2. * and / are higher precedence than + and -.

## Type coercions

There are 3 value types: integers, floats and errors. Pure integer expressions result in integer, expressions containing floats result in floats. If there is an error, for example division by zero or undefined variable, the expression evaluates to the error and any further arithmetics using the error would result in the same error. Some examples:

    1/0.0
    >  +Inf
    1/0
    >  division by zero
    a=1/0
    >  division by zero
    a+a
    >  division by zero
    c = b+a
    >  variable b not defined
    c*2
    >  variable b not defined

## Language

The language has the following statement types:

 - arithmetic expression
 - variable assignment

Expressions evaluate to a value printed in the REPL loop as answers, assignments results in the value assigned, but they are not expressions so an assignment can't be used apart from top level.

### Tokens

The following tokens are valid (using usual regular expression notation)

 - integer literal /\d+/
 - float literal /\d+.\d+/
 - variable name /[a-z]+/
 - single character token /[+-*/=()]/

tokens are separated with white-spaces.

### Grammar

Support unary minus at the grammar level as opposed to lexer level for negative number literals. This means that "- 5" is minus five with white-space or 2+-(3+1) works. In the following BNF non-terminals are lower case, terminals are upper case.

    statement: expression | assignment
    assignment: VARIABLE '=' expression 
    expression: addsub
    addsub: addsub /[+-]/ divmul | divmul
    divmul: divmul /[*/]/ unary | unary
    unary: '-' top | top
    top: INTL | FLOATL | VARIABLE  | '(' expression ')'

## Approach

A finite state machine based hand written lexer combined with a hand written parser using the parser combinator style approach to effectively create a recursive descent parser.

                         tokens           AST
    input text -> lexer --------> parser -----> logic --> output
    var state ------------------------------------^ |
          ^                                         |
          `--- write back --------------------------'

### Lexer

State machine based lexer, reading a character at a time. The lexer also inserts an EOL (end of line) at the end of the input if there is none, as the parser uses that to assert that all input has been consumed.

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
 - 4 aritmetic operations +, -, *, /
 - explicit evaluation order by parenthesis
 - type coersions, pure integer expressions result in integer, expressions containing floats result in floats

The language is right associative and has 2 precedence groups: +, - is lower than * and /. It is planned to make the language left associative.

## Language

The language has the following statement types:

 - arithmetic expression
 - variable assignment

Expressions evaluate to a value printed in the repl loop as answers, assignments results in the value assigned, but they are not expressions so an assignment can't be used apart from top level.

### Tokens

The following tokens are valid:

 - integer literal /\d+/
 - float literal /-?\d*.\d*/
 - variable name /[a-z]+/
 - operator plus, minus /[+-]/
 - operator multiply, divide /[*/]/
 - assign /=/
 - paren /[()]/

tokens are spearated by white-space.

### Grammar

Support unary minus at the grammar level as opposed to lexer level for negative number literals.
This means that "- 5" is minus five with white-space or 2+-(3+1) works.

    statement: expression | assignment
    assignment: VARIABLE ASSIGN expression 
    expression: addsub
    addsub: divmul ADDSUB addsub | divmul
    divmul: unary DIVMUL divmul | unary
    unary: UNARY top | top
    top: INTL | FLOATL | VARIABLE  | '(' expression ')'

## Approach

A finite state machine based hand written lexer combined with a hand written parser using the parser combinator style approach to effectively create a recursive descent parser.

                         tokens           AST
    input text -> lexer --------> parser -----> logic --> output
    var state ------------------------------------^ |
          ^                                         |
          `--- write back --------------------------'

### Lexer

State machine based lexer, reading a character at a time. The mechanism is basically 

    state <- start
    WHILE NOT EOL
      c <- NEXT CHAR
      state <- FSM(state, c)

The states are as followos, with the next character implying the next state.

    Start 
      - whitespace: Start
      - digit: Intl
      - [a-z]: Variable
      - '(', ')': Paren 
      - '+', '*', '/', '-': Op

    Intl
      - whitespace: emit Intl, Start
      - digit: Intl
      - [a-z]: ERROR
      - '(', ')': ERROR
      - '+', '-', '*', '/': emit Intl, Op
      - '.': Float

    Float
      - whitespace: emit Floatl, Start
      - digit: Floatl
      - [a-z]: ERROR
      - '(', ')': ERROR
      - '+', '-', '*', '/': emit Floatl, Op
      - '.': Float

    Variable 
      - whitespace: emit Variable, Start
      - digit: ERROR
      - [a-z]: Variable
      - '(', ')': ERROR 
      - '+', '*', '/', '-': emit Variable, Op

    Paren 
      emit Paren and
      - whitespace: Start
      - digit: Intl
      - [a-z]: Variable
      - '(', ')': Paren 
      - '+', '*', '/', '-': Op

    Op 
      emit Op and
      - whitespace: Start
      - digit: Intl
      - [a-z]: Variable
      - '(', ')': Paren 
      - '+', '*', '/', '-': Op



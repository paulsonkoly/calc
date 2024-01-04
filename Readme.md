# Calc

A Simple calculator repl. Supported features:

 - integer and floating point literals
 - variables
 - 4 aritmetic operations +, -, *, /
 - explicit evaluation order by parenthesis

The language is right associative and has 2 precedence groups: +, - is lower than * and /

## Language

The language has the following statement types:

 - arithmetic expression
 - variable assignment

Expressions evaluate to a value printed in the repl loop as answers, assignments don't result in a value,
 - they are not expressions so assignment can't be used apart from top level.

### Tokens

The following tokens are valid:

 - integer literal /-?\d+/
 - float literal /-?\d*.\d*/
 - variable name /[a-z]+/
 - operator plus, minus /[+-]/
 - operator multiply, divide /[*/]/
 - assign /=/
 - paren /[()]/

tokens are spearated by white-space.

### Grammar

Support unary minus at the grammar level as opposed to lexer level for negative number literals.
This means that "- 5" is minus five with white-space.

    statement: expression | assignment
    assignment: VARIABLE ASSIGN expression 
    expression: addsub
    addsub: divmul ADDSUB addsub | divmul
    divmul: unary DIVMUL divmul | unary
    unary: UNARY top | top
    top: INTL | FLOATL | VARIABLE  | '(' expression ')'

## Approach
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



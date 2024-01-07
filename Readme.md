# Calc

A simple calculator language / REPL. This project is merely for code practicing, not intended to be used.

The language can be used in a REPL or instructions can be read from a file. The REPL outputs its answer after '>' character.

    fib = (n) -> if (n <= 1) 1 else rec(n-1) + rec(n-2) 
    > function fib
    

    sum = (n) -> {
        i = 1
        s = 0
        while (i <= n) {
           s = s + i
        }
    }
    > function sum

    TODO SHOW USE

Supported features:

 - integer, floating point and boolean literals
 - variables
 - conditionals
 - loops
 - functions
 - arithmetic operations +, -, *, /; relational operations <, <=, >, >=, ==, !=; boolean operations & |.
 - explicit evaluation order by parenthesis
 - being functional in the sense that functions are first class values

## Type coercions

There are 5 value types: integers, floats, booleans, functions and errors.

Arithmetic operations and relational operations work on numbers, an expression containing only integers results in integer (or error), an expression containing a float results in a float. Relational operations work both on numbers and booleans, logic operations work only on booleans.

There are 5 precedence groups (from lowest to highest): 
    - relational
    - logic
    - + or -
    - * or /
    - unary minus

Incorrect operations result in error, any further calculation with an error results in the same error. Functions as values used in calculations result in error.

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

The token { defines a multi-line block, until the corresponding } is found. The REPL doesn't output for lines inside a block. The top level non-terminal is the program, a program consists of statements. Statements are normally a single line but if they contain a block they can span across mulitple lines.

The language has the following statement types:

 - expression
 - assignment
 - loop
 - conditional

### Expressions

All operators are left associative thus following the natural notations. 1-2+1 is 0 and not -2. Unary minus is supported as an operator, not part of a number literal, thus work with any expression.

TODO EXAMPLE

### Assignment

Any value type can be assigned to a variable. The variable name is not defined in the scope of the assigment right hand side, so when writing a recursive function one cannot use the variable the function is assigned to to do the recursive call. Instead the rec keyword can be used. Although assigments return the assigned value, they cannot be used in expressions, only as a statement.

### Loop and Conditionals

The only loop syntax is the while loop. Conditional code can be written with the if or the if .. else .. structures.


### Tokens

The following tokens are valid (using usual regular expression notation)

 - integer literal /\d+/
 - float literal /\d+.\d+/
 - variable name /[a-z]+/
 - single character token /[+-*/=()<>{}]/
 - double character token /<=|>=|==|!=|->/
 - new line /\n/

tokens are separated with white-spaces.

### Grammar

Support unary minus at the grammar level as opposed to lexer level for negative number literals. This means that "- 5" is minus five with white-space or 2+-(3+1) works. In the following BNF non-terminals are lower case, terminals are upper case.

    program: block "\n" program | block EOF
    block: "{" "\n" statements "\n" "}" | statement
    statements: statement "\n" statements | statement
    statement: assigment | loop | conditional | expression

    assignment: VARIABLE '=' block 
    loop: "while" expression block
    conditional: "if" expression block "else" block | "if" expression block

    expression: relational
    relational: logic /<|>|<=|>=|==|!=/ logic | logic
    logic: logic /[|&]/ addsub | addsub
    addsub: addsub /[+-]/ divmul | divmul
    divmul: divmul /[*/]/ unary | unary
    unary: '-' atom | atom
    atom: INTL | FLOATL | VARIABLE  | '(' expression ')'

## Approach

A finite state machine based hand written lexer combined with a hand written parser using the parser combinator style approach to effectively create a recursive descent parser.

                         tokens           AST
    input text -> lexer --------> parser -----> logic --> output
    var state ------------------------------------^ |
          ^                                         |
          `--- write back --------------------------'

### Lexer

State machine based lexer, reading a character at a time. The lexer also inserts an EOL (end of line) at the end of the input if there is none, as the parser uses that to assert that all input has been consumed.

# Calc

A simple calculator language / REPL. This project is merely for code practicing, not intended to be used.

The language can be used in a REPL or instructions can be read from a file. The REPL outputs its answer after '>' character.

    divides = (a, b) -> {
      r = false
      s = a
      while s <= b {
        if s == b {
          r = true
        }
        s = s + a
      }
      r
    }
    > function
  
    isprime = (n) -> {
      if n < 2 {
        false
      } else {
        i = 2
        r = true
        while i <= n / 2 {
          if divides(i, n) {
            r = false
          }
          i = i + 1	
        }
        r
      }
    }
    > function
  
    isprime(13)
    > true

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

## Variable lookup, shadowing

The language has very simple variable lookup rules. Function calls create new stack frames, function returns pop stack frames. Variables on read are looked up starting at the current frame, travesing each frame upwards in the stack until the variable is found. Variable writes set the variable in the current frame.

    a = 13
    >  13
    f = (n) -> {
        a = a+1
    }
    >  function
    f(1)
    >  14
    a
    >  13

We set a to 13 at the top frame. We set f to a function value. We call f passing argument 1. This does the following steps:

   1. push a new empty frame on the stack
   2. set n in the new frame to 1
   3. evaluate the function body, which sets a in the current frame to 14 (reading 13 from the frame above).
   4. pops the last frame from the stack

Now a is 13 as the variable was shadowed in the function call.

## Language

The token { defines a multi-line block, until the corresponding } is found. The REPL doesn't output for lines inside a block. The top level non-terminal is the program, a program consists of statements. Statements are on a single line up to the first new line character, blocks span across mulitple lines.

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
 - single character token /[+-*/=()<>{}],/
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
    atom: function | call | INTL | FLOATL | VARIABLE  | '(' expression ')'

    function: '(' parameters ')' '->' block
    parameters: VARIABLE ',' parameters | VARIABLE
    call: VARIABLE '(' arguments ')'
    arguments: expression ',' arguments | expression


## Approach

A finite state machine based hand written lexer combined with a hand written parser using the parser combinator style approach to effectively create a recursive descent parser.

                         tokens           AST
    input text -> lexer --------> parser -----> logic --> output
    var state ------------------------------------^ |
          ^                                         |
          `--- write back --------------------------'

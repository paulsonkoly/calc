# Calc

A simple calculator language / REPL. This project is merely for code practicing, not intended to be used.

The language can be used in a REPL or instructions can be read from a file. The REPL outputs its answer after '>' character.

    divides = (a, b) -> {
      s = a
      while s <= b {
        if s == b {
          return true
        }
        s = s + a
      }
      false
    }
    > function
  
    isprime = (n) -> {
      if n < 2 return false
      i = 2
      while i <= n / 2 {
        if divides(i, n) return false
        i = i + 1	
      }
      true
    }
    > function
  
    isprime(13)
    > true

Supported features:

 - integer, floating point and boolean literals
 - variables
 - conditionals
 - loops
 - functions and closures
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

## Variable lookup, shadowing, closures

Function calls create new stack frames, function returns pop stack frames. Variables on read are looked up starting at the current frame, traversing each frame upwards in the stack until the variable is found. Variable writes always set the variable in the current frame.

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

   1. create a new frame with the arguments populated, setting n to 1
   2. push the frame
   3. evaluate the function body, which sets a in the current frame to 14 (reading 13 from the frame above).
   4. pop the last frame from the stack

Now a is 13 as the variable was shadowed in the function call.

When a function is not a top level function but defined within a function, it becomes a closure. This is done by the call and the return pushing and popping 2 frames respectively. The first frame pushed is the frame the function was defined in, the second frame contains the arguments.

    f = (n) -> {
      a = 1
      (b) -> a + b + n
    }
    >  function

    foo = f(2)
    >  function

    foo(3)
    >  6

In this example the function returned from f holds reference to the frame that was pushed on the call of f. This frame contains both a=1 and n=2. The anonymous function is assigned to foo later, and at the call of foo, we push this frame, and a second frame containing b=3.

This allows us to implement Currying.

    curry = (f, a) -> (b) -> f(a, b)
    >  function

    sum = (a, b) -> a + b
    >  function

    plusthree = curry(sum, 3)
    >  function

    plusthree(5)
    >  8

Note however that only the top level frame of the function definition is retained, thus the following results in error:

    f = (x) -> {
      (y) -> {
        (z) -> x + y + z
      }
    }
    >  function

    first = f(1)
    >  function

    second = first(2)
    >  function

    second(3) 
    >  x not defined

## Language

The token { defines a multi-line block, until the corresponding } is found. The REPL doesn't output for lines inside a block. The top level non-terminal is the program, a program consists of statements. Statements are on a single line up to the first new line character, blocks span across multiple lines.

The language has the following statement types:

 - expression
 - assignment
 - loop
 - conditional
 - return

### Expressions

All operators are left associative thus following the natural notations. 1-2+1 is 0 and not -2. Unary minus is supported as an operator, not part of a number literal, thus work with any expression.

### Assignment

Any value type can be assigned to a variable. The variable name is not defined in the scope of the assignment right hand side, so when writing a recursive function one cannot use the variable the function is assigned to to do the recursive call. Instead the rec keyword can be used. Although assignments return the assigned value, they cannot be used in expressions, only as a statement.

### Loop and Conditionals

The only loop syntax is the while loop. Conditional code can be written with the if or the if .. else .. structures. As these are statements they end at the first newline, but one can use blocks to write multi line body loops and conditionals. This should explain why the first two examples are valid, but the third one is not.

    if true 1 else 2
    >  1

    if true {
    1
    } else {
    2
    }
    >  1

    if true 1
    >  1
    else 2
    Parser: end of file expected, got <"2" IntLit>

### Return

Returns from the current function call or block. Returns are valid outside of a function and they produce the returned value. They have an effect on the containing structure. For blocks the containing block evaluates to the return value without evaluating subsequent lines. For loops, encountering a return breaks out of the loop (and the result of the loop will be the return value.

### Tokens

The following tokens are valid (using usual regular expression notation)

 - integer literal /\d+/
 - float literal /\d+.\d+/
 - variable name /[a-z]+/
 - single character token /[+-*/=()<>{}],/
 - double character token /<=|>=|==|!=|->/
 - new line /\n/

tokens are separated with white-spaces.

### BNF

In the following BNF non-terminals are lower case, terminals are upper case or quoted strings.

    program: block "\n" program | block EOF
    block: "{" "\n" statements "\n" "}" | statement
    statements: statement "\n" statements | statement
    statement: assigment | loop | conditional | returning | expression

    assignment: VARIABLE '=' block 
    loop: "while" expression block
    conditional: "if" expression block "else" block | "if" expression block
    returning: "return" expression

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

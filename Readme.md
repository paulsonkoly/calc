# Calc

A simple calculator language / REPL.

The language can be used in a REPL or instructions can be read from a file. The REPL outputs its answer after '>' character.

    isprime = (n) -> {
      if n < 2 return false
      for i <- fromto(2, n/2+1) {
        if n % i == 0 return false
      }
      true
    }
    > function
  
    isprime(13)
    > true

Functional programming / Currying

    curry = (f, a) -> (b) -> f(a, b)
    >  function

    sum = (a, b) -> a + b
    >  function

    plusthree = curry(sum, 3)
    >  function

    plusthree(5)
    >  8

Iterators / generators

    evens = () -> {
      yield 0
      yield 2
      yield 4
    }
    > function

    for i <- evens() {
      write(i)
      i+3
    }
    0
    2
    4
    > 7 

## Editor support

There is syntax highlighting based on tree-sitter, and a small nvim plugin that enables neovim to download the treesitter parser and adds file type detection (ssuming .calc extension). Add [paulsonkoly/calc.nvim](https://github.com/paulsonkoly/calc.nvim) to your neovim package manager and require("calc") to add language support.

## Running calc

The language is meant to be a calculator REPL, and as such takes care of input/output automatically, but given it can also read source from a file it also supports some basic input output primitives. The calc program can run in 3 modes: reading a single line expression from its command line argument, running code from a REPL or reading code from a file.

### REPL

If there is no input file given and no command line argument to evaluate then the input is assumed to come from a terminal and we assume REPL mode. In this mode readline library is used to ease line editing. The token { defines a multi-line block, until the corresponding } is found. The REPL doesn't evaluate until multi line blocks are closed, and it automatically outputs the result after each evaluation.

### Command line argument

A single line statement can be passed as a command line argument:

    % ./calc -eval "1+2"
    3

Calc doesn't prefix the answer with '> ' in this case.

### File evaluation

If a single file name is provided on the command line the input is redirected from this file, in this case calc doesn't output evaluation results at all, for any output the program has to use the write function.

    % cat x.calc
    write(3)
    % ./calc x.calc
    3

## Builtin functions

Built in functions are loaded in the top level frame on the interpreter start up. They provide functionality that cannot be implemented in calc itself, or convinience functions. These are just regular function values defined in the global lexical scope.

| function | arity | returns                    | description                             |
|----------|-------|----------------------------|-----------------------------------------|
| read     | 0     | string                     | Reads a string from the stdin           |
| write    | 1     | no result error            | Writes the given value to the output    |
| aton     | 1     | int/float/conversion error | Converts a string to an int or a float  |
| toa      | 1     | string                     | Converts a value to a string            |
| error    | 1     | error                      | Converts a string to an error           |
| exit     | 1     | doesn't return/type error  | Exits the interpreter with exit code    |
| fromto   | 2     | iterator/type error        | fromto(a, b) iterates from a to b-1     |
| elems    | 1     | iterator/type error        | elems(ary) iterates the array elements  |
| indices  | 1     | iterator/type error        | indices(ary) iterates the array indices |

## Type coercions

There are 7 value types: integers, floats, booleans, functions, strings, arrays and errors.

Arithmetic operations and relational operations work on numbers, an expression containing only integers results in integer (or error), an expression containing a float results in a float. Relational operations work both on numbers, booleans and strings, logic operations work only on booleans.

Arrays are dynamic container of any type.

    funs = [ ["+", (a, b) -> a+b ], ["-", (a, b) -> a - b ] ]
    >  [["+", function], ["-", function]]

There are 6 precedence groups (from lowest to highest): 

    - <, >, <=, >=, ==, !=
    - &, |
    - + or -
    - *, / and %
    - unary -, # and !

### Length operator

    #[1,2,3]
    > 3

### Indexing

Array and string indexing has 2 forms: "apple"[1] results in "p"; "apple"[1:3] results in "pp". Indexing outside, or using a lower value for the upper index than the lower index results in index error.

In an expression array indexing binds stronger than any operator, thus

    #[[1,1,1]][0]
    >  3

### Errors

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

### Strings

String literals can be written using double quotes ("). Within a string a double quote has to be escaped: "\\"" is a string with a single element containing a double quote. Line breaks and any other character can be inserted within a string normally. Strings can be concatenated and indexed.

## Iterators and generators, yield and for

Assuming we have the following definition of `fromto` (available as a builtin function):

    fromto = (n, m) -> {
        while n < m {
            yield n
            n = n + 1
        }
    }

One can replace the following while loop

    i = 0
    while i < 10 {
       write(i)
       i = i + 1
    }

with the more concise

    for i <- fromto(0, 10) write(i)

`elems` and `indices` can also be implemented in similar fashion but also provided as builtin functions. One can write number generators or other iterators using yield.

An iterator or generator is an expression that when evaluated calls the yield keyword with some value. The syntax for a for loop is

    for <variable> <- <iterator> <for_loop_body>

yield is a keyword that is used to give flow control back to the for loop across function calls given the yield was invoked in the iterator part of the for construct. When yield yields a value the for loop resumes and when the loop body finishes the execution continues from the point the yield happened. This is until there are no more items to yield or the for loop body executes a return statement. Yielding in a call stack that doesn't have a containing for loop would have no effect, yield itself evaluates to the yielded value.

If one wants to observe side effects from the iterator it can lead to confusing results, all instructions will be executed in the iterator to the point when it returns to the for loop, unless the for loop body returns early.

    strangeiter = () -> {
        yield 1
        write("in iteration")
        yield 2
        yield 3
        yield 4
        write("end of the iterator")
    }
    > function

    for i <- strangeiter() {
       write(i)
       if i == 2 {
           return "done"
       }
    }
    1
    "in iteration"
    2
    > "done"

## Variable lookup, shadowing, closures

There are 3 types of variables, depending on the lexical scope, but their syntax is identical.

A variable at the global lexical scope is a global variable and visible in every scope where it's not shadowed, but is only writable from the global lexical scope.

A variable defined withing a function or a parameter to a function is local variable in the function.

A variable that's local variable in some function f that defines function g, becomes a closure variable within the call of g.

Variable reads look up variables in the order of local, closure, global. Variable writes write variables as local, shadowing previously visible variables by the same name.

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

`a` is a global variable shadowed in the function `f`.

When a function is not a top level function but defined within a function, it becomes a closure. This is done by it holding a reference to the stack frame that belonged to the function call that defined it.

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

Closures frames are saved and attached to the function value at the point of time when a function is created. Different calls to a function can have different values of the same variable:

    weird = (n) -> {
      i = n
      if n/2*2 == n {
         () -> i* 10
      } else {
         () -> i * 100
      }
    }
    >  function
    f = weird(2)
    >  function
    g = weird(3)
    >  function
    f()
    >  20
    g()
    >  300

Note however that only immediately containing lexical scope of the function definition is retained, thus the following results in error:

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

One can make this example work by making an explicit copy of x:

    f = (x) -> {
      (y) -> {
        x = x
        (z) -> x + y + z
      }
    }
    >  function

## Language

The top level non-terminal is the program, a program consists of statements. Statements are on a single line up to the first new line character, blocks span across multiple lines.

The language has the following statement types:

 - expression
 - assignment
 - loop
 - conditional
 - return

The followings are keywords: if, else, while, true, false. A variable name cannot be one of the keywords.

### Expressions

All operators are left associative thus following the natural notations. 1-2+1 is 0 and not -2. Unary minus is supported as an operator, not part of a number literal, thus work with any expression.

### Assignment

Any value type can be assigned to a variable. The variable name is not defined in the scope of the assignment right hand side. Although assignments return the assigned value, they cannot be used in expressions, only as a statement.

Recursive call to a function using the variable name the function is assigned to works however; because the function body is only evaluated when the function is called.

    f = (n) -> if n <= 0 0 else n + f(n-1)
    > function

    f(5)
    > 15

### Loop and Conditionals

while loops are a simple construct of a loop condition and a loop body. for loops have to be used with iterators. Conditional code can be written with the if or the if .. else .. structures. As these are statements they end at the first newline, but one can use blocks to write multi line body loops and conditionals. This should explain why the first two examples are valid, but the third one is not.

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

Returns from the current function call or block. Returns are valid outside of a function and they produce the returned value. They have an effect on the containing structure. For blocks the containing block evaluates to the return value without evaluating subsequent lines. For loops, encountering a return breaks out of the loop and the result of the loop will be the return value.

### Tokens

The following tokens are valid (using usual regular expression notation)

 - integer literal `/\d+/`
 - float literal `/\d+(\.\d+)?/`
 - string literal `/"([^"]|\\")*"/`
 - variable name `/[a-z]+/`
 - non sticky chars `/[(){},\[\]:]/`
 - sticky chars `/[+*/=<>!%-&|@]/`
 - new line `/\n/`

Tokens are separated with white-spaces. Sticky chars together are returned from the lexer as single lexeme. For example "<=" is a single lexeme.

### BNF

In the following BNF non-terminals are lower case, terminals are upper case or quoted strings.

    program: block "\n" program | block EOF
    block: "{" "\n" statements "\n" "}" | statement
    statements: statement "\n" statements | statement
    statement: loop | conditional | returning | assignment| expression

    assignment: VARIABLE '=' block 
    loop: "while" expression block
    conditional: "if" expression block "else" block | "if" expression block
    returning: "return" expression

    expression: relational
    relational: logic /<|>|<=|>=|==|!=/ logic | logic
    logic: logic /[|&]/ addsub | addsub
    addsub: addsub /[+-]/ divmul | divmul
    divmul: divmul /[*/%]/ unary | unary
    unary: /[-#!]/ index | index
    index: atom "[" expression ":" expression "]" | atom "[" expression "]" | atom
    atom: function | call | INTL | FLOATL | BOOLL | STRINGL | VARIABLE  | '(' expression ')'

    function: "()" "->" block | '(' parameters ')' "->" block
    parameters: VARIABLE ',' parameters | VARIABLE
    call: VARIABLE "()" | VARIABLE '(' arguments ')'
    arguments: expression ',' arguments | expression

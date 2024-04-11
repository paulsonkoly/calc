# Calc

Calc is an interpreted language / REPL. It has dynamic typing, generally strict evaluation semantics with pass by value calls, and lazy iterator semantics. It supports closures, first class functions, composable iterators as functions. It only violates referential transparency due to IO. A function always returns the same value for the same input, and always behaves the same way, except if it does input/output. The syntax allows mixing procedural and functional programming.

```scheme
; all asserts that a predicate f is true for all iterated values
all = (iter, f) -> {
  for e <- iter() if !f(e) return false
  true
}
```
> function

```scheme
isprime = (n) -> {
  if n < 2 return false
  all(() -> fromto(2, n/2+1), (i) -> n % i != 0)
}
```
> function

```scheme
isprime(13)
```
> true

Further code examples can be found in the [examples directory](https://github.com/paulsonkoly/calc/tree/main/examples).


## Types

There are 6 value types: integers, floats, booleans, functions, strings and arrays.

There is no automatic type conversion between types, except in an arithmetic expression integers are converted to floats if the expression contains floats. Equality check works between any types. Function equality always result in false. Invalid operations like type errors, division by zero etc. result in runtime error.

If an expression doesn't hold a value, it evaluates to nil. Calculation with nil or assigning nil results in runtime error.

### Arrays and strings

Arrays are dynamic containers of any type.

```scheme
funs = [ ["+", (a, b) -> a+b ], ["-", (a, b) -> a - b ] ]
```
> [["+", function], ["-", function]]

Array and string indexing has 2 forms: "apple"[1] results in "p"; "apple"[1:3] results in "pp". Indexing outside, or using a lower value for the upper index than the lower index results in index error.

String literals can be written using double quotes ("). Within a string a double quote has to be escaped: "\\"" is a string with a single element containing a double quote. Line breaks and any other character can be inserted within a string normally. Strings can be concatenated and indexed.

In an expression array indexing binds stronger than any operator, thus

```scheme
#[[1,1,1]][0]
```
>  3

\# is the length operator

```scheme
#[1,2,3]
```
> 3

## Iterators and generators, yield and for

Assuming we have the following definition of `fromto` (available as a builtin function):

```scheme
fromto = (n, m) -> {
    while n < m {
        yield n
        n = n + 1
    }
}
```

One can replace the following while loop

```scheme
i = 0
while i < 10 {
   write(i)
   i = i + 1
}
```

with the more concise

```scheme
for i <- fromto(0, 10) write(i)
```

`elems` and `indices` can also be implemented in a similar fashion but also provided as builtin functions. One can write number generators or other iterators using yield.

An iterator or generator is an expression that when evaluated calls the yield keyword with some value. The syntax for a for loop is

    for <variable> <- <iterator> <for_loop_body>

yield is a keyword that is used to give flow control back to the for loop across function calls given the yield was invoked in the iterator part of the for construct. When yield yields a value the for loop resumes and when the loop body finishes the execution continues from the point the yield happened. This is until there are no more items to yield or the for loop body executes a return statement. Yielding in a call stack that doesn't have a containing for loop would have no effect, yield itself evaluates to the yielded value.

### Composing iterators

Iterators can be used in the definition of a new iterator:

```scheme
map = (f, iter) -> for e <- iter() yield f(e)
```

Embedding for loops iterate the cross product of the iterators:

```scheme
for i <- fromto(1,3) {
  for j <- elems("ab") {
    write(toa(i) + " " + j + "\n")
  }
}
```

> 1 a  
> 1 b  
> 2 a  
> 2 b  
> > nil  

Listing multiple iterators in a for loop zips iterations. The iteration stops when any of the iterators terminate.

```scheme
for i, j <- fromto(1, 3), elems("ab") write(toa(i) + " " + j + "\n")
```

> 1 a  
> 2 b  
> > nil  

## Variable lookup, shadowing, closures

There are 3 types of variables, depending on the lexical scope, but their syntax is identical.

A variable at the global lexical scope is a global variable and visible in every scope where it's not shadowed but is only writable from the global lexical scope.

A variable defined within a function or a parameter to a function is a local variable in the function.

A variable that's a local variable in some function f that defines function g, becomes a closure variable within the call of g.

Variable reads look up variables in the order of local, closure and global. Variable writes write variables as local, shadowing previously visible variables by the same name.

```scheme
a = 13
```
>  13

```scheme
f = (n) -> {
    a = a+1
}
```
>  function

```scheme
f(1)
```
>  14

```scheme
a
```
>  13

`a` is a global variable shadowed in the function `f`.

When a function is not a top level function but defined within a function, it becomes a closure. This is done by holding a reference to the stack frame that belonged to the function call that defined it.

```scheme
f = (n) -> {
  a = 1
  (b) -> a + b + n
}
```
>  function

```scheme
foo = f(2)
```
>  function


```scheme
foo(3)
```
>  6

In this example, the function returned from f holds reference to the frame that was pushed on the call of f. This frame contains both a=1 and n=2. The anonymous function is assigned to foo later, and at the call of foo, we push this frame, and a second frame containing b=3.

Closure variables are shared with the defining function until the defining function returns. Updates to these variables are visible in the closure, but the closure cannot write these variables, as they are not local.

```scheme
f = () -> {
  x = 1
  g = () -> x
  x = 2
  g
}
```
> function

```scheme
g()
```
> 2

For a closure only the immediately containing lexical scope of the function definition is retained, thus the following results in error:

```scheme
f = (x) -> {
  (y) -> {
    (z) -> x + y + z
  }
}
```
>  function

```scheme
first = f(1)
```
>  function

```scheme
second = first(2)
```
>  function

```scheme
second(3) 
```
> nil error

One can make this example work by making an explicit copy of x:

```scheme
f = (x) -> {
  (y) -> {
    x = x
    (z) -> x + y + z
  }
}
```

## Editor support

There is syntax highlighting based on tree-sitter, and a small nvim plugin that enables neovim to download the treesitter parser and adds file type detection (assuming .calc extension). Add [paulsonkoly/calc.nvim](https://github.com/paulsonkoly/calc.nvim) to your neovim package manager and require("calc") to add language support.

## Running calc

The language is meant to be a calculator REPL, and as such takes care of input/output automatically, but given it can also read source from a file it also supports some basic input output primitives. The calc program can run in 3 modes: reading a single line expression from its command line argument, running code from a REPL or reading code from a file.

### REPL

If there is no input file given and no command line argument to evaluate then the input is assumed to come from a terminal and we assume REPL mode. In this mode, readline library is used to ease line editing. The token { defines a multi-line block, until the corresponding } is found. The REPL doesn't evaluate until multi line blocks are closed, and it automatically outputs the result after each evaluation.

### Command line argument

A single line statement can be passed as a command line argument:

    % ./calc -eval "1+2"
    3

Calc doesn't prefix the answer with '> ' in this case.

### File evaluation

If a single file name is provided on the command line the input is schemeirected from this file, in this case calc doesn't output evaluation results at all, for any output the program has to use the write function.

    % cat x.calc
    write(3)
    % ./calc x.calc
    3

## Builtin functions

Built in functions are loaded in the top level frame on the interpreter start up. They provide functionality that cannot be implemented in calc itself, or convenience functions. These are just regular function values defined in the global lexical scope.

| function | arity | returns                    | description                             |
|----------|-------|----------------------------|-----------------------------------------|
| read     | 0     | string                     | Reads a string from the stdin           |
| write    | 1     | nil                        | Writes the given value to the output    |
| aton     | 1     | int/float/conversion error | Converts a string to an int or a float  |
| toa      | 1     | string                     | Converts a value to a string            |
| exit     | 1     | doesn't return/type error  | Exits the interpreter with exit code    |
| fromto   | 2     | iterator                   | fromto(a, b) iterates from a to b-1     |
| elems    | 1     | iterator                   | elems(ary) iterates the array elements  |
| indices  | 1     | iterator                   | indices(ary) iterates the array indices |


### Binary operators

There are 5 precedence groups (from lowest to highest): 

| Operator     | Precedence | Types                                                 | Description                                  |
|--------------|------------|-------------------------------------------------------|----------------------------------------------|
| &&, \|\|     | 0          | int/int, bool/bool                                    | bitwise, or boolean and, or - low precedence |
| <, >, <=, >= | 1          | int or float/int or float                             | relational                                   |
| ==, !=       | 1          | any/any                                               | equality check                               |
| &, \|        | 2          | int/int, bool/bool                                    | bitwise, or boolean and or - high precedence |
| +            | 3          | int or float/int or float, array/array, string/string | addition                                     |
| -            | 3          | int or float/int or float                             | substraction                                 |
| *, /         | 4          | int or float/int or float                             | division/mulitplication                      |
| <<, >>       | 4          | int/int                                               | bitshift                                     |
| %            | 4          | int/int                                               | modulo                                       |

### Unary operators

Unary operators bind stronger than binary operators. All unary operators are prefix.

| Operator | Types         | Description |
|----------|---------------|-------------|
| -        | int, float    | negation    |
| #        | array, string | length      |
| !        | bool          | not         |
| ~        | int           | binary flip |


# Errors

Given incorrect source code 3 types of errors can happen: lexer errors, parser errors and runtime errors. In case of lexer or parser errors calc simply outputs the error without evaluating any code. In case of a runtime error the code is run, and the runtime error is outputted with the current state of the interpreter including the failing instruction, the arguments to the failing instruction, and the stack backtrace in each memory context that are active at the point of the error.

```
calc repl
12££12
Lexer: unexpected char £ in integer literal
12££12
^~^
```

```
1+)
Parser: variable name expected, got )
1+)
 ^^
 ```

```
f = () -> {
  yield 1
  1/0
  yield 2
}
> function

g = (x) -> {
  for i <- f() {
    write(toa(i+x) + "\n")
  }
}
> function

h = () -> g(13)
> function

h()
14
RUNTIME ERROR : division by zero
    85: 0X3201000000000007 : JMP 7 
    86: 0X4406000000000018 : YIELD DS[24] 
    87: 0X0400000000000000 : POP 
--> 88: 0X0E3600000019001A : DIV DS[25] DS[26] ; 1, 0
    89: 0X0400000000000000 : POP 
    90: 0X440600000000001B : YIELD DS[27] 
memory context 0x140002eda40
= stack =============================================
IP: 98 f() args: 
IP: 116 g() args: arg[0]: 13
=====================================================
memory context 0x14000108ab0
= stack =============================================
IP: 116 g() args: arg[0]: 13
IP: 121 h() args: 
=====================================================
 
```

## Language

Comments start with ; until the end of the line.

The top level non-terminal is the program, a program consists of statements. Statements are on a single line up to the first new line character, whereas blocks span across multiple lines.

The language has the following statement types:

 - expression
 - assignment
 - loop
 - conditional
 - return

The followings are keywords: if, else, while, for, return, yield, true, false. A variable name cannot be one of the keywords.

### Expressions

All operators are left associative thus following the natural notations. 1-2+1 is 0 and not -2. Unary minus is supported as an operator, not part of a number literal, and thus work with any expression.

### Assignment

Any value type can be assigned to a variable. The variable name is not defined in the scope of the assignment right hand side. Although assignments return the assigned value, they cannot be used in expressions, only as a statement.

Recursive call to a function using the variable name the function is assigned to works, however; because the function body is only evaluated when the function is called.

```scheme
f = (n) -> if n <= 0 0 else n + f(n-1)
```
> function

```scheme
f(5)
```
> 15

### Loop and Conditionals

while loops are a simple construct of a loop condition and a loop body. for loops have to be used with iterators. Conditional code can be written with the if or the if .. else .. structures. As these are statements they end at the first newline, but one can use blocks to write multi line body loops and conditionals. This should explain why the first two examples are valid, but the third one is not.

```scheme
if true 1 else 2
```
>  1

```scheme
if true {
   1
} else {
   2
}
```
>  1

```scheme
if true 1
```
>  1

```scheme
else 2
```
> Parser: end of file expected, got <"2" IntLit>

### Return

Returns from the current function call or block. Returns are valid outside of a function and they produce the returned value. They have an effect on the containing structure. For blocks, the containing block evaluates to the return value without evaluating subsequent lines. For loops, encountering a return breaks out of the loop and the result of the loop will be the return value.

### Tokens

The following tokens are valid (using the usual regular expression notation)

 - integer literal `/\d+/`
 - float literal `/\d+(\.\d+)?/`
 - string literal `/"([^"]|\\")*"/`
 - variable name `/[a-z]+/`
 - non sticky chars `/[(){},\[\]:]/`
 - sticky chars `/[+*/=<>!%-&|@]/`
 - new line `/\n/`

Tokens are separated with white-spaces. Sticky chars together are returned from the lexer as a single lexeme. For example "<=" is a single lexeme.

### BNF

In the following BNF non-terminals are lower case, and terminals are upper case or quoted strings.

    program: block "\n" program | block EOF
    block: "{" "\n" statements "\n" "}" | statement
    statements: statement "\n" statements | statement
    statement: whileLoop | forLoop | conditional | returning | yield | assignment| expression

    assignment: VARIABLE '=' block 
    whileLoop: "while" expression block
    forLoop: "for" VARIABLE "<-" expression block
    conditional: "if" expression block "else" block | "if" expression block
    returning: "return" expression
    yield: "yield" expression

    expression: lowprec
    lowprec: relational /&&|[|]{2}/ relational | relational
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

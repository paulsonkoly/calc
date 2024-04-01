// Package node is defines the abstract syntax tree (AST) node.
package node

// Type is AST node type.
type Type interface {
	STRewriter
	graphvizzer
	ByteCoder
	Constanter
}

// Invalid is an invalid AST node.
type Invalid struct{}

// Call is function call.
type Call struct {
	Name      Type // Variable referencing function
	Arguments List // Arguments passed to the function
}

// Function is a function definition.
type Function struct {
	Parameters List // Parameters of the function
	Body       Type // Body of the function
	LocalCnt   int  // count of local variables
}

// Int is integer literal.
type Int int

// Float is float literal.
type Float float64

// String is string literal.
type String string

// Bool is boolean literal.
type Bool bool

// BinOp is a binary operator of any kind, anything from "=", etc.
type BinOp struct {
	Op    string // Op is the operator string
	Left  Type   // Left operand
	Right Type   // Right operand
}

// UnOp is a unary operator of any kind, ie. '-'.
type UnOp struct {
	Op     string // Op is the operator string
	Target Type   // Target is the operand
}

type IndexAt struct {
	Ary Type // Ary is the indexed node
	At  Type // At is the index
}

type IndexFromTo struct {
	Ary  Type // Ary is the indexed node
	From Type // From is the start of the range
	To   Type // To is the end of the range
}

// If is a conditional construct without an else case.
type If struct {
	Condition Type // Condition is the condition for the if statement
	TrueCase  Type // TrueCase is executed if condition evaluates to true
}

// IfElse is a conditional construct.
type IfElse struct {
	Condition Type // Condition is the condition for the if statement
	TrueCase  Type // TrueCase is executed if condition evaluates to true
	FalseCase Type // FalseCase is executed if condition evaluates to false
}

// While is a loop construct.
type While struct {
	Condition Type // Condition is the condition for the loop
	Body      Type // Body is the loop body
}

// For is a loop for iterators ans generators.
type For struct {
	VarRef   Type // VarRef is variable reference
	Iterator Type // Value is assigned value
	Body     Type // Body is the loop body
}

// Return is a return statement.
type Return struct {
	Target Type // Target is the returned value
}

// Yield statement.
type Yield struct {
	Target Type // Target is the yielded value
}

// Variable name.
type Name string

// Local variable reference.
type Local struct {
	Ix      int    // Ix is the index in the call frame
	VarName string // VarName is variable name
}

// Closure variable reference.
type Closure struct {
	Ix      int    // Ix is the index in the call frame
	VarName string // VarName is variable name
}

type Assign struct {
	VarRef Type // VarRef is variable reference
	Value  Type // Value is assigned value
}

// Block is a code block / sequence that was in '{', '}'.
type Block struct {
	Body []Type // Body is the block body
}

// List is a list of arguments or parameters depending on whether it's a function call or definition.
type List struct {
	Elems []Type // Elems are the parameters or arguments
}

// builtins

// Read reads a string from stdin.
type Read struct{}

// Write writes a value to stdout.
type Write struct{ Value Type }

// Aton converts a string to a number type.
type Aton struct{ Value Type }

// Toa converts a valye to a string.
type Toa struct{ Value Type }

// Exit exits the interpreter with an os exit code.
type Exit struct{ Value Type }

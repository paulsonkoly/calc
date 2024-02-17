// node is an abstract syntax tree (AST) node
package node

// Type is AST node type
type Type interface {
	PrettyPrinter
	Token() string
}

// Invalid is an invalid AST node
type Invalid struct{}

// Call is function call
type Call struct {
	Name      string // Name of the function called
	Arguments List   // Arguments passed to the function
}

// Function is a function definition
type Function struct {
	Parameters List // Parameters of the function
	Body       Type // Body of the function
}

// Int is integer literal
type Int string

// Float is float literal
type Float string

// String is string literal
type String string

// BinOp is a binary operator of any kind, anything from "=", etc.
type BinOp struct {
	Op    string // Op is the operator string
	Left  Type   // Left operand
	Right Type   // Right operand
}

// UnOp is a unary operator of any kind, ie. '-'
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

// If is a conditional construct without an else case
type If struct {
	Condition Type // Condition is the condition for the if statement
	TrueCase  Type // TrueCase is executed if condition evaluates to true
}

// IfElse is a conditional construct
type IfElse struct {
	Condition Type // Condition is the condition for the if statement
	TrueCase  Type // TrueCase is executed if condition evaluates to true
	FalseCase Type // FalseCase is executed if condition evaluates to false
}

// While is a loop construct
type While struct {
	Condition Type // Condition is the condition for the loop
	Body      Type // Body is the loop body
}

// Return is a return statement
type Return struct {
	Target Type // Target is the returned value
}

// Variable name (also "true", "false" etc.)
type Name string

// Block is a code block / sequence that was in '{', '}'
type Block struct {
	Body []Type // Body is the block body
}

// List is a list of arguments or parameters depending on whether it's a function call or definition
type List struct {
	Elems []Type // Elems are the parameters or arguments
}

// builtins
type Read struct {
	Target Name // Target is the variable name the read will be assigned to
}

type Write struct {
	Value Type // Value is the written value
}

type Repl struct{}

func (i Invalid) Token() string     { return "" }
func (c Call) Token() string        { return "" }
func (f Function) Token() string    { return "" }
func (i Int) Token() string         { return string(i) }
func (f Float) Token() string       { return string(f) }
func (s String) Token() string      { return string(s) }
func (b BinOp) Token() string       { return b.Op }
func (u UnOp) Token() string        { return u.Op }
func (u IndexAt) Token() string     { return "" }
func (u IndexFromTo) Token() string { return "" }
func (i If) Token() string          { return "" }
func (i IfElse) Token() string      { return "" }
func (w While) Token() string       { return "" }
func (r Return) Token() string      { return "" }
func (r Read) Token() string        { return "" }
func (w Write) Token() string       { return "" }
func (n Name) Token() string        { return string(n) }
func (b Block) Token() string       { return "" }
func (l List) Token() string        { return "" }
func (r Repl) Token() string        { return "" }

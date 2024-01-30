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
	Name      string
	Arguments List
}

// Function is a function definition
type Function struct {
	Parameters List
	Body       Type
}

// Int is integer literal
type Int string

// Float is float literal
type Float string

// BinOp is a binary operator of any kind, anything from "=", etc.
type BinOp struct {
	Op    string
	Left  Type
	Right Type
}

// UnOp is a unary operator of any kind, ie. '-'
type UnOp struct {
	Op     string
	Target Type
}

// If is a conditional construct without an else case
type If struct {
	Condition Type
	TrueCase  Type
}

// IfElse is a conditional construct
type IfElse struct {
	Condition Type
	TrueCase  Type
	FalseCase Type
}

// While is a loop construct
type While struct {
	Condition Type
	Body      Type
}

// Return is a return statement
type Return struct {
	Target Type
}

// Variable name (also "true", "false" etc.)
type Name string

// Block is a code block / sequence that was in '{', '}'
type Block struct {
	Body []Type
}

// List is a list of arguments or parameters
type List struct {
	Elems []Type
}

func (i Invalid) Token() string  { return "" }
func (c Call) Token() string     { return "" }
func (f Function) Token() string { return "" }
func (i Int) Token() string      { return string(i) }
func (f Float) Token() string    { return string(f) }
func (b BinOp) Token() string    { return b.Op }
func (u UnOp) Token() string     { return u.Op }
func (i If) Token() string       { return "" }
func (i IfElse) Token() string   { return "" }
func (w While) Token() string    { return "" }
func (r Return) Token() string   { return "" }
func (n Name) Token() string     { return string(n) }
func (b Block) Token() string    { return "" }
func (l List) Token() string     { return "" }

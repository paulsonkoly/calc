package lexer

import "fmt"

type TokenType int

const (
	InvalidToken = TokenType(iota)
	IntLit
	FloatLit
	VarName
	Op
	Assign
	Paren
)

type Token struct {
	Value string    // the slice into the input containing the token value
	Type  TokenType // the type of the token
}

func (t Token) String() string {
	return fmt.Sprintf("<\"%v\" %v>", t.Value, t.Type)
}

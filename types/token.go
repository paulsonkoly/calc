package types

import (
	"fmt"

	"github.com/phaul/calc/combinator"
)

type TokenType int

const (
	InvalidToken = TokenType(iota)
	EOL
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
	if t.Type == EOL {
		return fmt.Sprintf("<%v>", t.Type)
	}
	return fmt.Sprintf("<\"%v\" %v>", t.Value, t.Type)
}

func (t Token) Node() combinator.Node {
	return Node{Token: t}
}

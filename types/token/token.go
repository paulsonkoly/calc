package token

import (
	"fmt"
)

type TokenType int

const (
	// Invalid should not be produced by the lexer, however the parser uses it in compound AST nodes
	Invalid = TokenType(iota)
	EOL          // end of line
	EOF          // end of file (end of input)
	IntLit       // integer literal
	FloatLit     // float literal
	Name         // sequence of alphabeth chars, variable names or keywords
	Sticky       // one of +, -, *, /, =, <, >, ! a sequence of these stick together in a single lexeme
	NotSticky    // one of (, ), {, }, `,` a sequence of these gives a sequence of single char lexemes
)

// Token as produced by the lexer
type Type struct {
	Value string    // the slice into the input containing the token value
	Type  TokenType // the type of the token
}

func (t Type) String() string {
	switch t.Type {

	case EOL, EOF:
		return fmt.Sprintf("<%v>", t.Type)

	case Sticky, NotSticky:
		return t.Value

	default:
		return fmt.Sprintf("<\"%v\" %v>", t.Value, t.Type)
	}
}

package token

import (
	"fmt"
)

type Kind int

const (
	Invalid   = Kind(iota) // Invalid token
	EOL                    // EOL is end of line
	EOF                    // EOF is end of file
	IntLit                 // IntLit is integer literal
	FloatLit               // FloatLit is float literal
	StringLit              // StringLit is String literal
	Name                   // Name is a variable name or keyword
	Sticky                 // one of +, -, *, /, =, <, >, ! a sequence of these stick together in a single lexeme
	NotSticky              // one of (, ), {, }, `,` a sequence of these gives a sequence of single char lexemes
)

// Token as produced by the lexer
type Type struct {
	Value string // Value is the token string contained withing the input stream
	Type  Kind   // Type of the token
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

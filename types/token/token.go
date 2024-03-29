// Package token defines the lexer tokens.
package token

import (
	"fmt"
)

// Kind is the token kind.
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

// Token as produced by the lexer.
type Type struct {
	from  int
	to    int
	Value string // Value is the token string contained within the input stream
	Type  Kind   // Type of the token
}

// WithFromTo returns a Type with the given value and indices into the input stream.
func WithFromTo(typ Kind, value string, from int, to int) Type {
	return Type{from: from, to: to, Value: value, Type: typ}
}

// String converts a token into a string.
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

// From is the index of the start of the token in the input stream.
func (t Type) From() int { return t.from }

// To is the index of the end of the token in the input stream.
func (t Type) To() int { return t.to }

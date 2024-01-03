// My parser combinator library
package combinator

import "fmt"

// seems one can write haskell in every language

type Token interface {
	Node() Node // lift a lexer token to an AST node that holds that single token
}
type Node any // AST Node

// A lexer that produces a stream of tokens. Next() advances the lexer and
// returns true until all tokens are returned. Err() and Token() do not modify
// the lexer and start returning values after the first Next() call
type Lexer interface {
	Next() bool   // Is there a next token
	Err() error   // signal lexing error
	Token() Token // The next Token
}

type Rollback interface {
	Snapshot() // Push current state on a stack so it can be recovered
	Rollback() // Recover last state that was pushed
}

type RollbackLexer interface {
	Lexer
	Rollback
}

// A parser that accepts input and returns sequence of parsed nodes or error
// It's the users responsibility to combine compound nodes into one, ie. the
// sub-parser results can be combined into a Node that has tree pointers to the
// sub expressions
type Parser[L RollbackLexer, N Node] func(input L) ([]N, error)

// Parses with a and if that fails does what b would do
func Or[L RollbackLexer, N Node](a, b Parser[L, N]) Parser[L, N] {
	return func(input L) ([]N, error) {
		input.Snapshot()
		aRes, aErr := a(input)
		if aErr != nil {
			input.Rollback()
		} else {
			return aRes, nil
		}
		bRes, bErr := b(input)
		if bErr != nil {
			return bRes, bErr
		}
		return bRes, nil
	}
}

// Appends the given parsers together with Or. In effect parses with the first
// succeeding parser and returns what that would return
func Any[L RollbackLexer, N Node](args ...Parser[L, N]) Parser[L, N] {
	if len(args) < 1 {
		panic("Parser: Any needs at least one parser")
	}
	r := args[0]
	for _, p := range args[1:] {
		r = Or(r, p)
	}
	return r
}

// Parses with a and then continues parsing with b. Only succeeds if both a and
// b succeeds and returns the concatenated result from both
func And[L RollbackLexer, N Node](a, b Parser[L, N]) Parser[L, N] {
	return func(input L) ([]N, error) {
		aRes, aErr := a(input)
		if aErr != nil {
			return aRes, aErr
		}
		bRes, bErr := b(input)
		return append(aRes, bRes...), bErr
	}
}

// Parses with a sequence of parsers, returns the concatenated result from all
// if all are successful
func Seq[L RollbackLexer, N Node](args ...Parser[L, N]) Parser[L, N] {
	if len(args) < 1 {
		panic("Parser: Seq needs at least one parser")
	}
	r := args[0]

	for _, p := range args {
		r = And(r, p)
	}
	return r
}

// given a predicate p on a lexer Token, parses successfully if the predicate is
// true and returns the node holding the parsed token
func Accept[L RollbackLexer, N Node](p func(Token) bool) Parser[L, N] {
	return func(input L) ([]N, error) {
		if !input.Next() {
			return nil, fmt.Errorf("Parser: unexpected end of input")
		}
		if input.Err() != nil {
			return nil, input.Err()
		}
		tok := input.Token()
		if !p(tok) {
			return nil, fmt.Errorf("Parser: %v failed", tok)
		}
		return []N{tok.Node().(N)}, nil
	}
}

func Fmap[L RollbackLexer, N Node](f func([]N) []N, p Parser[L, N]) Parser[L, N] {
	return func(input L) ([]N, error) {
		r, err := p(input)
		if err != nil {
			return nil, err
		}
		return f(r), nil
	}
}

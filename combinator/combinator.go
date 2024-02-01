// package combinator parser combinator library
package combinator

import "fmt"

// seems one can write haskell in every language

// Token represents a lexeme from a lexer
type Token any

// Node is an AST node, with pointers to sub-tree nodes and potentially some
// parse information data
type Node any

// TokenWrapper wraps a token in a single AST node containing the token
type TokenWrapper interface {
	Wrap(Token) Node
}

// Lexer produces a stream of tokens. Next() advances the lexer and
// returns true until all tokens are returned. Err() and Token() do not modify
// the lexer and start returning values after the first Next() call
type Lexer interface {
	Next() bool   // Is there a next token
	Err() error   // signal lexing error
	Token() Token // The next Token
}

// Transaction adds snapshot and rollback functionality to some API.

// It's expected to be stack based just like a database transaction
type Transaction interface {
	Snapshot() // Push current state on a stack so it can be recovered
	Rollback() // Recover last state that was pushed
	// Commit state. After commit the previous snapshot point is removed, the
	// next Rollback returns to the snapshot prior to that
	Commit()
}

type RollbackLexer interface {
	Lexer
	Transaction
}

// Parser accepts input and returns sequence of parsed nodes or error.
//
// It's the users responsibility to combine compound nodes into one, ie. the
// sub-parser results can be combined into a Node that has tree pointers to the
// sub expressions
type Parser func(input RollbackLexer) ([]Node, error)

// Or is a choice between two parsers
//
// parses with a and if that fails does what b would do
func Or(a, b Parser) Parser {
	return func(input RollbackLexer) ([]Node, error) {
		input.Snapshot()
		aRes, aErr := a(input)
		if aErr == nil {
			input.Commit()
			return aRes, nil
		}
		input.Rollback()
		bRes, bErr := b(input)
		return bRes, bErr
	}
}

// OneOf is a choice between many parsers
//
// Appends the given parsers together with Or. In effect parses with the first
// succeeding parser and returns what that would return. Fails if all parsers
// fail.
func OneOf(args ...Parser) Parser {
	if len(args) < 1 {
		panic("Parser: OneOf needs at least one parser")
	}
	r := args[0]
	for _, p := range args[1:] {
		r = Or(r, p)
	}
	return r
}

// And is sequencing two parsers
//
// Parses with a and then continues parsing with b. Only succeeds if both a and
// b succeed and returns the concatenated result from both a and b.
func And(a, b Parser) Parser {
	return func(input RollbackLexer) ([]Node, error) {
		aRes, aErr := a(input)
		if aErr != nil {
			return aRes, aErr
		}
		bRes, bErr := b(input)
		return append(aRes, bRes...), bErr
	}
}

// Seq is sequencing many parsers
//
// Parses with a sequence of parsers, returns the concatenated result from all
// if all are successful. Fails if any of the parsers fail.
func Seq(args ...Parser) Parser {
	if len(args) < 1 {
		panic("Parser: Seq needs at least one parser")
	}
	r := args[0]

	for _, p := range args[1:] {
		r = And(r, p)
	}
	return r
}

// Some runs the given parser a as many times as it would succeed
//
// Some fails if a doesn't succeed at least once and succeeds otherwise. It
// returns the concatenated result of all successful runs.
func Some(a Parser) Parser {
	return And(a, Any(a))
}

// Any runs the given parser a as many times as it would succeed
//
// Any never fails, and it returns the concatenated result of all successful
// runs which can be potentially empty.
func Any(a Parser) Parser {
	return func(input RollbackLexer) ([]Node, error) {
		r := make([]Node, 0)

		for {
			input.Snapshot()
			aRes, aErr := a(input)
			if aErr == nil {
				input.Commit()
				r = append(r, aRes...)
			} else {
				input.Rollback()
				break
			}
		}
		return r, nil
	}
}

// SeparatedBy parses with a sequence of a, separated by b.
//
// It never fails, but result might be empty. It asserts that a sequence of a
// is interspersed with b, the sequence not ending with b. The parse results of
// b are thrown away, it returns the sequenced results of a.
func SeparatedBy(a, b Parser) Parser {
  return func(input RollbackLexer) ([]Node, error) {
    input.Snapshot()
    r, aErr := a(input)
    if aErr != nil {
      input.Rollback()
      return []Node{}, nil
    }
    input.Commit()
    for {
      input.Snapshot()
      _, bErr := b(input)
      if bErr != nil {
        input.Rollback()
        return r, nil
      }
      aRes, aErr := a(input)
      if aErr != nil {
        input.Rollback()
        return r, nil
      }
      input.Commit()
      r = append(r, aRes...)
    }
  }
}

// JoinedWith parses with a sequence of a, separated by b.
//
// It fails if a doesn't succeed at least once. If there is only one a it
// doesn't assert a subsequent b. Otherwise the sequence has to end with b. The
// parse results of b are thrown away, it returns the sequenced results of a.
func JoinedWith(a, b Parser) Parser {
	return Or(
		Some(Fmap(func(ab []Node) []Node { return ab[0:1] }, And(a, b))),
		a,
	)
}

// SurroundedBy parses with a sequence of a, b, c but returns the parse result
// of b only
//
// It fails if any of a, b, c fails. Useful for asserting parenthesis style rules.
func SurroundedBy(a, b, c Parser) Parser {
	return func(input RollbackLexer) ([]Node, error) {
		_, aErr := a(input)
		if aErr != nil {
			return nil, aErr
		}
		bRes, bErr := b(input)
		if bErr != nil {
			return nil, bErr
		}
		_, cErr := c(input)

		return bRes, cErr
	}
}

// Accept asserts the next token
//
// given a predicate p on a lexer Token, parses successfully if the predicate
// is true for the next token provided by l, and then returns the node holding
// that token.
// If p is false or the lexer fails then Accept fails.
func Accept(p func(Token) bool, msg string, wrp TokenWrapper) Parser {
	return func(input RollbackLexer) ([]Node, error) {
		if !input.Next() {
			return nil, fmt.Errorf("Parser: unexpected end of input")
		}
		if input.Err() != nil {
			return nil, input.Err()
		}
		tok := input.Token()
		if !p(tok) {
			return nil, fmt.Errorf("Parser: %s expected, got %v", msg, tok)
		}
		return []Node{wrp.Wrap(tok)}, nil
	}
}

// Fmap maps f : []Node -> []Node function on parser p
//
// Returns a modified version of p that succeeds when p succeeds but if p
// returns r the modified version returns f(r)
func Fmap(f func([]Node) []Node, p Parser) Parser {
	return func(input RollbackLexer) ([]Node, error) {
		r, err := p(input)
		if err != nil {
			return nil, err
		}
		return f(r), nil
	}
}

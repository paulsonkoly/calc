// package combinator parser combinator library
package combinator

import (
	"fmt"
)

// seems one can write haskell in every language

// Token represents a lexeme from a lexer.
type Token interface {
	From() int
	To() int
}

// Node is an AST node, with pointers to sub-tree nodes and potentially some
// parse information data.
type Node any

// TokenWrapper wraps a token in a single AST node containing the token.
type TokenWrapper interface {
	Wrap(Token) Node
}

// Lexer produces a stream of tokens. Next() advances the lexer and
// returns true until all tokens are returned. Err() and Token() do not modify
// the lexer and start returning values after the first Next() call.
type Lexer interface {
	From() int
	To() int
	Next() bool   // Is there a next token
	Err() error   // signal lexing error
	Token() Token // The next Token
}

// Transaction adds snapshot and rollback functionality to some API.

// It's expected to be stack based just like a database transaction.
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
// sub expressions.
type Parser func(input RollbackLexer) ([]Node, *Error)

// Ok is a parser that doesn't do anything just returns a successful parse
// result.
func Ok() Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
		return []Node{}, nil
	}
}

// Assert asserts the given parser p would succeed without consuming input.
// It returns empty parse result.
func Assert(p Parser) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
		input.Snapshot()
		defer input.Rollback()

		_, pErr := p(input)

		return []Node{}, pErr
	}
}

// Not asserts that the given parser p will fail.
func Not(p Parser) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
		_, pErr := p(input)
		if pErr == nil {
			return nil, &Error{message: "expecting error"}
		}
		return []Node{}, nil
	}
}

// Conditional is a pair of parsers. Once Gate succeeds we don't roll back but
// we are committed to parse with OnSuccess.
//
// Conditional is meant to solve the following problem.
//
// OneOf is much simpler and simpler to use than Choose. However in the following
// language there is a problem.
//
//	S -> Ax
//	A -> OneOf(B, C)
//	B -> qbcd
//	C -> q
//
// qbczx is the input text, here using rule B was intended, giving the error
// that z doesn't match expected d. However in OneOf C matches, so in rule A
// there is no error. Rule S fails q (from rule C) is not followed by x,
// rolling back bcz part of the input. The reported error gets far from the
// actual error character.
//
// Do this instead:
//
//	S -> Ax
//	A -> Choose( Cond{G: qb, S: cd }, Cond{ G: Ok(), S : q } )
//
// Assuming choice rules can have identifying prefixes.
type Conditional struct {
	Gate      Parser // Gate gates Choose and Any from using Success if Predicate fails
	OnSuccess Parser // OnSuccess is the parser that runs after gate succeeds
}

// Choose is a choice between many parsers
//
// It uses Conditionals, asserting that the Gate succeeds and if so it commits
// to parsing with OnSuccess part of choice. It rolls back a failing Predicate
// automatically. A succeeding predicate will be prepended to the result of
// success.
func Choose(choices ...Conditional) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
		for _, c := range choices {
			input.Snapshot()
			pRes, pErr := c.Gate(input)
			if pErr == nil {
				input.Commit()
				sRes, sErr := c.OnSuccess(input)
				if sRes != nil {
					return append(pRes, sRes...), sErr
				}
				return nil, sErr
			}
			input.Rollback()
		}
		// Modify your grammar using Choose, so the last choice always passes
		//
		// we can't really create more meaningful error here, than the error from
		// the last choice, one can use Ok for Predicate in the last case
		panic("no predicates succeeded in choice")
	}
}

// OneOf is a choice between many parsers
//
// It tries each parser in turn, rolling back the input after each failed
// attempt It is meant to be used with terminal rules only, for complex
// language rules prefer Choose because it gives much closer syntax errors to
// the actual error location.
func OneOf(args ...Parser) Parser {
	if len(args) < 1 {
		panic("Parser: OneOf needs at least one parser")
	}
	return func(input RollbackLexer) ([]Node, *Error) {
		input.Snapshot()
		pRes, pErr := args[0](input)
		if pErr == nil {
			input.Commit()
			return pRes, nil
		}
		input.Rollback()
		for _, p := range args[1:] {
			input.Snapshot()
			pRes, pErr = p(input)
			if pErr == nil {
				input.Commit()
				return pRes, nil
			}
			input.Rollback()
		}
		return nil, pErr
	}
}

// And is sequencing two parsers
//
// Parses with a and then continues parsing with b. Only succeeds if both a and
// b succeed and returns the concatenated result from both a and b.
func And(a, b Parser) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
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

// Any parses with 0 or more execution of a
//
// It tries parsing with predicate and if that succeeds it commits to go one
// more step with success. The predicate failing is rolled back automatically.
// The succeeding predicate result is concatenated with the result
//
// If the rule doesn't have a meaningful prefix to condition on, one can place
// the rule in the gate, and just use Ok() for OnSuccess. This gives a
// behaviour where failures are rolled back.
func Any(a Conditional) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
		r, err := make([]Node, 0), (*Error)(nil)

		for {
			input.Snapshot()
			pRes, pErr := a.Gate(input)
			if pErr == nil {
				input.Commit()
				aRes, aErr := a.OnSuccess(input)
				r = append(append(r, pRes...), aRes...)
				err = aErr
				if err != nil {
					break
				}
			} else {
				input.Rollback()
				break
			}
		}
		return r, err
	}
}

// SeparatedBy parses with a sequence of a, separated by b.
//
// It never fails, but result might be empty. It asserts that a sequence of a
// is interspersed with b, the sequence not ending with b. The parse results of
// b are thrown away, it returns the sequenced results of a.
func SeparatedBy(a, b Parser) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
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

// SurroundedBy parses with a sequence of a, b, c but returns the parse result
// of b only
//
// It fails if any of a, b, c fails. Useful for asserting parenthesis style rules.
func SurroundedBy(a, b, c Parser) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
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
	return func(input RollbackLexer) ([]Node, *Error) {
		if !input.Next() {
			return nil, &Error{from: input.From(), to: input.To(), message: "Parser: unexpected end of input"}
		}
		if input.Err() != nil {
			return nil, &Error{from: input.From(), to: input.To(), message: input.Err().Error()}
		}
		tok := input.Token()
		if !p(tok) {
			msg := fmt.Sprintf("Parser: %s expected, got %v", msg, tok)
			return nil, &Error{from: tok.From(), to: tok.To(), message: msg}
		}
		return []Node{wrp.Wrap(tok)}, nil
	}
}

// Fmap maps f : []Node -> []Node function on parser p.
//
// Returns a modified version of p that succeeds when p succeeds but if p
// returns r the modified version returns f(r).
func Fmap(f func([]Node) []Node, p Parser) Parser {
	return func(input RollbackLexer) ([]Node, *Error) {
		r, err := p(input)
		if err != nil {
			return nil, err
		}
		return f(r), nil
	}
}

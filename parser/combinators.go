package parser

import "github.com/phaul/calc/lexer"

type Parser func(input lexer.Lexer) ([]ASTNode, error)

func Or(a, b Parser) Parser {
	return func(input lexer.Lexer) ([]ASTNode, error) {
		//input snapshot
		aRes, aErr := a(input)
		if aErr != nil {
			// input rollback
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

func Any(args ...Parser) Parser {
	if len(args) < 1 {
		panic("Parser: Any needs at least one parser")
	}
	r := args[0]
	for _, p := range args[1:] {
		r = Or(r, p)
	}
	return r
}

func And(a, b Parser) Parser {
	return func(input lexer.Lexer) ([]ASTNode, error) {
		aRes, aErr := a(input)
		if aErr != nil {
			return aRes, aErr
		}
		bRes, bErr := b(input)
		return append(aRes, bRes...), bErr
	}
}

func Seq(args ...Parser) Parser {
	if len(args) < 1 {
		panic("Parser: Seq needs at least one parser")
	}
	r := args[0]

	for _, p := range args {
		r = And(r, p)
	}
	return r
}

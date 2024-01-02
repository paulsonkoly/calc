package parser

import (
	"fmt"

	"github.com/phaul/calc/lexer"
)

// seems one can write haskell in every language

type ASTNode struct {
	Token    lexer.Token
	Children []ASTNode
}

func wrap(nodes []ASTNode) []ASTNode {
	return []ASTNode{{Children: nodes}}
}

func fmap(f func([]ASTNode) []ASTNode, p Parser) Parser {
	return func(input lexer.Lexer) ([]ASTNode, error) {
		r, err := p(input)
		if err != nil {
			return nil, err
		}
		return f(r), nil
	}
}

func term(tokType lexer.TokenType) Parser {
	return func(input lexer.Lexer) ([]ASTNode, error) {
		var result ASTNode
		if !input.Next() {
			return nil, fmt.Errorf("Parser: unexpected end of input")
		}
		if input.Err != nil {
			return nil, input.Err
		}
		if input.Token.Type != tokType {
			return nil, fmt.Errorf("Parser: %v failed", tokType)
		}
		result.Token = input.Token
		return []ASTNode{result}, nil
	}
}

func token(token string) Parser {
	return func(input lexer.Lexer) ([]ASTNode, error) {
		var result ASTNode
		if !input.Next() {
			return nil, fmt.Errorf("Parser: unexpected end of input")
		}
		if input.Err != nil {
			return nil, input.Err
		}
		if input.Token.Value != token {
			return nil, fmt.Errorf("Parser: %v failed", token)
		}
		result.Token = input.Token
		return []ASTNode{result}, nil
	}
}

// The grammar
var intLit = term(lexer.IntLit)
var floatLit = term(lexer.FloatLit)
var varName = term(lexer.VarName)

// these can't be defined as variables as they are self referencing
func paren(input lexer.Lexer) ([]ASTNode, error) {
	r, err := fmap(wrap, Seq(token("("), expression, token(")")))(input)
	return r, err
}

func top(input lexer.Lexer) ([]ASTNode, error) {
	r, err := Any(intLit, floatLit, varName, paren)(input)
	return r, err
}

func unary(input lexer.Lexer) ([]ASTNode, error) {
	r, err := Or(fmap(wrap, (Seq(token("-"), top))), top)(input)
	return r, err
}

func divmul(input lexer.Lexer) ([]ASTNode, error) {
	op := Or(token("*"), token("/"))
	r, err := Or(fmap(wrap, Seq(unary, op, divmul)), unary)(input)
	return r, err
}

func addsub(input lexer.Lexer) ([]ASTNode, error) {
	op := Or(token("+"), token("-"))
	r, err := Or(fmap(wrap, Seq(divmul, op, addsub)), divmul)(input)
	return r, err
}

func expression(input lexer.Lexer) ([]ASTNode, error) {
	r, err := addsub(input)
	return r, err
}

var assignment = fmap(wrap, Seq(varName, token("="), expression))
var statement = Or(assignment, expression)


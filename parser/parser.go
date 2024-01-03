package parser

import (
	"fmt"

	"github.com/phaul/calc/combinator"
	"github.com/phaul/calc/lexer"
)

type Token lexer.Token

func (t Token) Node() combinator.Node {
	return ASTNode{Token: t}
}

type Lexer struct {
	l lexer.Lexer
}

func (l *Lexer) Next() bool {
	return l.l.Next()
}

func (l Lexer) Err() error {
	return l.l.Err
}

func (l Lexer) Token() Token {
	return Token(l.l.Token)
}

func (l Lexer) Rollback() {
}

func (l Lexer) Snapshot() {
}

type ASTNode struct {
	Token    Token
	Children []ASTNode
}

var or = combinator.Or[*Lexer, ASTNode]

func wrap(nodes []combinator.Node) []combinator.Node {
	return []ASTNode{{Children: nodes.([]ASTNode)}}
}

func term(tokType lexer.TokenType) combinator.Parser {
	return combinator.Accept(
		func(t combinator.Token) bool {
			return t.(Token).Type == tokType
		},
  )
}

func token(token string) combinator.Parser {
  return combinator.Accept(
		func(t combinator.Token) bool {
			return t.(Token).Value == token
		},
  )
}

// The grammar
var intLit = term(lexer.IntLit)
var floatLit = term(lexer.FloatLit)
var varName = term(lexer.VarName)

// these can't be defined as variables as they are self referencing
func paren(input lexer.Lexer) ([]ASTNode, error) {
	r, err := combinator.Fmap(wrap, Seq(token("("), expression, token(")")))(input)
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

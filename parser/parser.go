package parser

import (
	c "github.com/phaul/calc/combinator"
	l "github.com/phaul/calc/lexer"
	t "github.com/phaul/calc/types"
)

func Parse(input string) ([]t.Node, error) {
	l := l.NewTLexer(input)
	rn := make([]t.Node, 0)

	r, err := statement(&l)
	for _, e := range r {
		rn = append(rn, e.(t.Node))
	}
	return rn, err
}

type node = t.Node
type token t.Token

type tokenType = t.TokenType

func wrap(nodes []c.Node) []c.Node {
	var children []node

	if len(nodes) > 0 {
		children = make([]node, 0)
	}
	for _, n := range nodes {
		children = append(children, n.(node))
	}
	r := []c.Node{c.Node(node{Children: children})}

	return r
}

func acceptTerm(tokType tokenType) c.Parser {
	return c.Accept(func(tok c.Token) bool { return tok.(t.Token).Type == tokType })
}

func acceptToken(str string) c.Parser {
	return c.Accept(func(tok c.Token) bool { return tok.(t.Token).Value == str })
}

// The grammar
var intLit = acceptTerm(t.IntLit)
var floatLit = acceptTerm(t.FloatLit)
var varName = acceptTerm(t.VarName)

// these can't be defined as variables as they are self referencing
func paren(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(wrap, c.Seq(acceptToken("("), expression, acceptToken(")")))(input)
	return r, err
}

func top(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Any(intLit, floatLit, varName, paren)(input)
	return r, err
}

func unary(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Or(c.Fmap(wrap, (c.Seq(acceptToken("-"), top))), top)(input)
	return r, err
}

func divmul(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("*"), acceptToken("/"))
	r, err := c.Or(c.Fmap(wrap, c.Seq(unary, op, divmul)), unary)(input)
	return r, err
}

func addsub(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("+"), acceptToken("-"))
	r, err := c.Or(c.Fmap(wrap, c.Seq(divmul, op, addsub)), divmul)(input)
	return r, err
}

func expression(input c.RollbackLexer) ([]c.Node, error) {
	r, err := addsub(input)
	return r, err
}

var assignment = c.Fmap(wrap, c.Seq(varName, acceptToken("="), expression))
var statement = c.Or(assignment, expression)

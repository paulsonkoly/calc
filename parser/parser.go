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

type tokenType = t.TokenType

func binOp(nodes []c.Node) []c.Node {
	if len(nodes) != 3 {
		panic("incorrect number of sub nodes for binary operator")
	}

	r := nodes[1].(t.Node)
	r.Children = []t.Node{nodes[0].(t.Node), nodes[2].(t.Node)}
	return []c.Node{r}
}

func unOp(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		panic("incorrect number of sub nodes for unary operator")
	}

	r := nodes[0].(t.Node)
	r.Children = []t.Node{nodes[1].(t.Node)}
	return []c.Node{r}
}

func first(nodes []c.Node) []c.Node  { return []c.Node{nodes[0]} }
func second(nodes []c.Node) []c.Node { return []c.Node{nodes[1]} }

func acceptTerm(tokType tokenType, msg string) c.Parser {
	return c.Accept(func(tok c.Token) bool { return tok.(t.Token).Type == tokType }, msg)
}

func acceptToken(str string) c.Parser {
	return c.Accept(func(tok c.Token) bool { return tok.(t.Token).Value == str }, str)
}

// The grammar
var intLit = acceptTerm(t.IntLit, "integer literal")
var floatLit = acceptTerm(t.FloatLit, "float literal")
var varName = acceptTerm(t.VarName, "variable name")

// these can't be defined as variables as there are cycles in their
// definitions, otherwise we could write:
//
//    var paren = c.Fmap(second, c.Seq(acceptToken("("), expression, acceptToken(")")))
//
func paren(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(second, c.Seq(acceptToken("("), expression, acceptToken(")")))(input)
	return r, err
}

func top(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Any(floatLit, intLit, varName, paren)(input)
	return r, err
}

func unary(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Or(c.Fmap(unOp, (c.Seq(acceptToken("-"), top))), top)(input)
	return r, err
}

func divmul(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("*"), acceptToken("/"))
	r, err := c.Or(c.Fmap(binOp, c.Seq(unary, op, divmul)), unary)(input)
	return r, err
}

func addsub(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("+"), acceptToken("-"))
	r, err := c.Or(c.Fmap(binOp, c.Seq(divmul, op, addsub)), divmul)(input)
	return r, err
}

func expression(input c.RollbackLexer) ([]c.Node, error) {
	r, err := addsub(input)
	return r, err
}

var assignment = c.Fmap(binOp, c.Seq(varName, acceptToken("="), expression))

var statement = c.Fmap(first, c.Or(c.And(assignment, acceptTerm(t.EOL, "end of line")),
	c.And(expression, acceptTerm(t.EOL, "end of line"))))

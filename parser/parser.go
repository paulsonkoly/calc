package parser

import (
	"log"

	c "github.com/phaul/calc/combinator"
	l "github.com/phaul/calc/lexer"
	t "github.com/phaul/calc/types"
)

func Parse(input string) ([]t.Node, error) {
	l := l.NewTLexer(input)
	rn := make([]t.Node, 0)

	r, err := program(&l)
	for _, e := range r {
		rn = append(rn, e.(t.Node))
	}
	return rn, err
}

type tokenType = t.TokenType

func unOp(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		log.Panicf("incorrect number of sub nodes for unary operator (%d)", len(nodes))
	}

	r := nodes[0].(t.Node)
	r.Children = []t.Node{nodes[1].(t.Node)}
	return []c.Node{r}
}

func leftChain(nodes []c.Node) []c.Node {
	if len(nodes) < 3 || len(nodes)%2 == 0 {
		log.Panicf("incorrect number of sub nodes for left chain (%d)", len(nodes))
	}
	r := nodes[0]
	for i := 1; i+1 < len(nodes); i += 2 {
		n := nodes[i].(t.Node)
		n.Children = []t.Node{r.(t.Node), nodes[i+1].(t.Node)}
		r = n
	}
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
var varName = acceptTerm(t.Name, "variable name")

// these can't be defined as variables as there are cycles in their
// definitions, otherwise we could write:
//
//	var paren = c.Fmap(second, c.Seq(acceptToken("("), expression, acceptToken(")")))
func paren(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(second, c.Seq(acceptToken("("), expression, acceptToken(")")))(input)
	return r, err
}

func atom(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.OneOf(floatLit, intLit, acceptToken("true"), acceptToken("false"), varName, paren)(input)
	return r, err
}

func unary(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Or(c.Fmap(unOp, (c.Seq(acceptToken("-"), atom))), atom)(input)
	return r, err
}

func divmul(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("*"), acceptToken("/"))
	r, err := c.Or(c.Fmap(leftChain, c.And(c.Some(c.And(unary, op)), unary)), unary)(input)
	return r, err
}

func addsub(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("+"), acceptToken("-"))
	r, err := c.Or(c.Fmap(leftChain, c.And(c.Some(c.And(divmul, op)), divmul)), divmul)(input)
	return r, err
}

var relOp = c.OneOf(
	acceptToken("=="),
	acceptToken("!="),
	acceptToken("<="),
	acceptToken(">="),
	acceptToken("<"),
	acceptToken(">"),
)

func relational(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Or(c.Fmap(leftChain, c.Seq(addsub, relOp, addsub)), addsub)(input)
	return r, err
}

func expression(input c.RollbackLexer) ([]c.Node, error) {
	r, err := relational(input)
	return r, err
}

var assignment = c.Fmap(leftChain, c.Seq(varName, acceptToken("="), expression))
var statement = c.Fmap(first, c.Or(assignment, expression))
var program = c.Some(c.Fmap(first, c.And(statement, acceptTerm(t.EOL, "end of line"))))

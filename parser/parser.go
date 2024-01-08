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

// unOp is used for unary operators
//
// It rewrites the pair of nodes putting the second under the first.
func unOp(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		log.Panicf("incorrect number of sub nodes for unary operator (%d)", len(nodes))
	}

	r := nodes[0].(t.Node)
	r.Children = []t.Node{nodes[1].(t.Node)}
	return []c.Node{r}
}

// leftChain rewrites a sequence of binary operators applied on operands in a
// left assictive structure
//
// In effect it arranges a+b+c sequence in:
//
//	+
//	|-+
//	| |-a
//	| `-b
//	`-c
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

func wrap(nodes []c.Node) []c.Node {
	r := t.Node{Children: make([]t.Node, 0)}
	for _, n := range nodes {
		r.Children = append(r.Children, n.(t.Node))
	}
	return []c.Node{r}
}

func control(nodes []c.Node) []c.Node {
	if len(nodes) != 3 && len(nodes) != 5 {
		log.Panicf("incorrect number of sub nodes for control (%d)", len(nodes))
	}
	r := nodes[0].(t.Node)
	r.Children = []t.Node{nodes[1].(t.Node), nodes[2].(t.Node)}
	if len(nodes) == 5 {
		r.Children = append(r.Children, nodes[4].(t.Node))
	}
	return []c.Node{r}
}

func allButLast(nodes []c.Node) []c.Node { return nodes[0 : len(nodes)-1] }

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
	r, err := c.SurroundedBy(acceptToken("("), expression, acceptToken(")"))(input)
	return r, err
}

func atom(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.OneOf(function, floatLit, intLit, acceptToken("true"), acceptToken("false"), varName, paren)(input)
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

func assignment(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(leftChain, c.Seq(varName, acceptToken("="), expression))(input)
	return r, err
}

func statement(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.OneOf(conditional, loop, assignment, expression)(input)
	return r, err
}

func conditional(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(control,
		c.Seq(acceptToken("if"), expression, c.Or(c.Seq(block, acceptToken("else"), block), block)))(input)
	return r, err
}

func loop(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(control, c.Seq(acceptToken("while"), expression, block))(input)
	return r, err
}

func function(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(leftChain, c.Seq(arguments, acceptToken("->"), block))(input)
	return r, err
}

var arguments = c.Fmap(wrap,
	c.SurroundedBy(
		acceptToken("("),
		c.SeparatedBy(varName, acceptToken(",")),
		acceptToken(")")))

var eol = acceptTerm(t.EOL, "end of line")
var eof = acceptTerm(t.EOF, "end of file")

func block(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Or(
		c.Fmap(wrap,
			c.SurroundedBy(
				c.And(acceptToken("{"), eol),
				c.JoinedWith(statement, eol),
				acceptToken("}"))),
		statement)(input)
	return r, err
}

var program = c.Fmap(allButLast, c.And(c.JoinedWith(block, eol), eof))

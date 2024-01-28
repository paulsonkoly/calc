package parser

import (
	"log"

	c "github.com/phaul/calc/combinator"
	l "github.com/phaul/calc/lexer"
	"github.com/phaul/calc/types/node"
	"github.com/phaul/calc/types/token"
)

func Parse(input string) ([]node.Type, error) {
	l := l.NewTLexer(input)
	rn := make([]node.Type, 0)

	r, err := program(&l)
	for _, e := range r {
		rn = append(rn, e.(node.Type))
	}
	return rn, err
}

type tokenType = token.TokenType

// unaryOp is used for unary operators
//
// It rewrites the pair of nodes putting the second under the first.
func unaryOp(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		log.Panicf("incorrect number of sub nodes for unary operator (%d)", len(nodes))
	}

	r := nodes[0].(node.Type)
	r.Children = []node.Type{nodes[1].(node.Type)}
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
		n := nodes[i].(node.Type)
		n.Children = []node.Type{r.(node.Type), nodes[i+1].(node.Type)}
		r = n
	}
	return []c.Node{r}
}

func wrap(nodes []c.Node) []c.Node {
	r := node.Type{Children: make([]node.Type, 0)}
	for _, n := range nodes {
		r.Children = append(r.Children, n.(node.Type))
	}
	return []c.Node{r}
}

func mkFCall(nodes []c.Node) []c.Node {
	r := node.Type{Token: token.Type{Type: token.Call}, Children: []node.Type{nodes[0].(node.Type), nodes[1].(node.Type)}}
	return []c.Node{r}
}

func control(nodes []c.Node) []c.Node {
	if len(nodes) != 3 && len(nodes) != 5 {
		log.Panicf("incorrect number of sub nodes for control (%d)", len(nodes))
	}
	r := nodes[0].(node.Type)
	r.Children = []node.Type{nodes[1].(node.Type), nodes[2].(node.Type)}
	if len(nodes) == 5 {
		r.Children = append(r.Children, nodes[4].(node.Type))
	}
	return []c.Node{r}
}

func allButLast(nodes []c.Node) []c.Node { return nodes[0 : len(nodes)-1] }

type tokenWrapper struct{}

// TODO rename once t doesn't conlict
func (_ tokenWrapper) Wrap(tok c.Token) c.Node {
	return node.Type{Token: tok.(token.Type)}
}

func acceptTerm(tokType tokenType, msg string) c.Parser {
	tokenWrap := tokenWrapper{}
	return c.Accept(func(tok c.Token) bool { return tok.(token.Type).Type == tokType }, msg, tokenWrap)
}

func acceptToken(str string) c.Parser {
	tokenWrap := tokenWrapper{}
	return c.Accept(func(tok c.Token) bool { return tok.(token.Type).Value == str }, str, tokenWrap)
}

// The grammar
var intLit = acceptTerm(token.IntLit, "integer literal")
var floatLit = acceptTerm(token.FloatLit, "float literal")
var varName = acceptTerm(token.Name, "variable name")

// these can't be defined as variables as there are cycles in their
// definitions, otherwise we could write:
//
//	var paren = c.Fmap(second, c.Seq(acceptToken("("), expression, acceptToken(")")))
func paren(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.SurroundedBy(acceptToken("("), expression, acceptToken(")"))(input)
	return r, err
}

func atom(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.OneOf(function, call, floatLit, intLit, acceptToken("true"), acceptToken("false"), varName, paren)(input)
	return r, err
}

func unary(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Or(c.Fmap(unaryOp, (c.Seq(acceptToken("-"), atom))), atom)(input)
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

func logic(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("&"), acceptToken("|"))
	r, err := c.Or(c.Fmap(leftChain, c.And(c.Some(c.And(addsub, op)), addsub)), addsub)(input)
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
	r, err := c.Or(c.Fmap(leftChain, c.Seq(logic, relOp, logic)), logic)(input)
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
	r, err := c.OneOf(conditional, loop, returning, assignment, expression)(input)
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

func returning(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(unaryOp, c.And(acceptToken("return"), expression))(input)
	return r, err
}

func function(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(leftChain, c.Seq(parameters, acceptToken("->"), block))(input)
	return r, err
}

var parameters = c.Fmap(wrap,
	c.SurroundedBy(
		acceptToken("("),
		c.SeparatedBy(varName, acceptToken(",")),
		acceptToken(")")))

func call(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkFCall, c.Seq(varName, arguments))(input)
	return r, err
}

func arguments(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(wrap,
		c.SurroundedBy(
			acceptToken("("),
			c.SeparatedBy(expression, acceptToken(",")),
			acceptToken(")")))(input)
	return r, err
}

var eol = acceptTerm(token.EOL, "end of line")
var eof = acceptTerm(token.EOF, "end of file")

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

package parser

import (
	c "github.com/phaul/calc/combinator"
	"github.com/phaul/calc/lexer"
	"github.com/phaul/calc/types/node"
	"github.com/phaul/calc/types/token"
)

func Parse(input string) ([]node.Type, error) {
	l := lexer.NewTLexer(input)
	rn := make([]node.Type, 0)

	r, err := program(&l)
	for _, e := range r {
		rn = append(rn, e.(node.Type))
	}
	return rn, err
}

func allButLast(nodes []c.Node) []c.Node { return nodes[0 : len(nodes)-1] }

func acceptTerm(tokType token.TokenType, msg string) c.Parser {
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
	r, err := c.Or(c.Fmap(mkUnaryOp, (c.Seq(acceptToken("-"), atom))), atom)(input)
	return r, err
}

func divmul(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("*"), acceptToken("/"))
	r, err := c.Or(c.Fmap(mkLeftChain, c.And(c.Some(c.And(unary, op)), unary)), unary)(input)
	return r, err
}

func addsub(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("+"), acceptToken("-"))
	r, err := c.Or(c.Fmap(mkLeftChain, c.And(c.Some(c.And(divmul, op)), divmul)), divmul)(input)
	return r, err
}

func logic(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("&"), acceptToken("|"))
	r, err := c.Or(c.Fmap(mkLeftChain, c.And(c.Some(c.And(addsub, op)), addsub)), addsub)(input)
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
	r, err := c.Or(c.Fmap(mkLeftChain, c.Seq(logic, relOp, logic)), logic)(input)
	return r, err
}

func expression(input c.RollbackLexer) ([]c.Node, error) {
	r, err := relational(input)
	return r, err
}

func assignment(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkLeftChain, c.Seq(varName, acceptToken("="), expression))(input)
	return r, err
}

func statement(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.OneOf(conditional, loop, returning, assignment, expression)(input)
	return r, err
}

func conditional(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkIf,
		c.Seq(acceptToken("if"), expression, c.Or(c.Seq(block, acceptToken("else"), block), block)))(input)
	return r, err
}

func loop(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkWhile, c.Seq(acceptToken("while"), expression, block))(input)
	return r, err
}

func returning(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkReturn, c.And(acceptToken("return"), expression))(input)
	return r, err
}

func function(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkFunction, c.Seq(parameters, acceptToken("->"), block))(input)
	return r, err
}

var parameters = c.Fmap(mkList,
	c.SurroundedBy(
		acceptToken("("),
		c.SeparatedBy(varName, acceptToken(",")),
		acceptToken(")")))

func call(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkFCall, c.Seq(varName, arguments))(input)
	return r, err
}

func arguments(input c.RollbackLexer) ([]c.Node, error) {
	r, err := c.Fmap(mkList,
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
		c.Fmap(mkBlock,
			c.SurroundedBy(
				c.And(acceptToken("{"), eol),
				c.JoinedWith(statement, eol),
				acceptToken("}"))),
		statement)(input)
	return r, err
}

var program = c.Fmap(allButLast, c.And(c.JoinedWith(block, eol), eof))

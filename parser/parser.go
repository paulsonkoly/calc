package parser

import (
	c "github.com/paulsonkoly/calc/combinator"
	"github.com/paulsonkoly/calc/lexer"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/token"
)

var Keywords = [...]string{"if", "else", "while", "return", "true", "false"}

type Type struct {}

func (t Type)Parse(input string) ([]node.Type, error) {
  return Parse(input)
}

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
var stringLit = acceptTerm(token.StringLit, "string literal")

func varName(input c.RollbackLexer) ([]c.Node, error) {
	return c.Accept(
		func(tok c.Token) bool {
			ctok := tok.(token.Type)
			if ctok.Type != token.Name {
				return false
			}
			for _, kw := range Keywords {
				if kw == ctok.Value {
					return false
				}
			}
			return true
		},
		"variable name",
		tokenWrapper{},
	)(input)
}

// these can't be defined as variables as there are cycles in their
// definitions, otherwise we could write:
//
//	var paren = c.Fmap(second, c.Seq(acceptToken("("), expression, acceptToken(")")))
func paren(input c.RollbackLexer) ([]c.Node, error) {
	return c.SurroundedBy(acceptToken("("), expression, acceptToken(")"))(input)
}

func arrayLit(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkList,
		c.SurroundedBy(
			acceptToken("["),
			c.SeparatedBy(expression, acceptToken(",")),
			acceptToken("]")))(input)
}

func atom(input c.RollbackLexer) ([]c.Node, error) {
	return c.OneOf(
		function,
		call,
		floatLit,
		intLit,
		acceptToken("true"),
		acceptToken("false"),
		stringLit,
		arrayLit,
		varName,
		paren)(input)
}

func index(input c.RollbackLexer) ([]c.Node, error) {
	return c.Or(
		c.Fmap(mkIndex,
			c.Or(
				c.Seq(atom, acceptToken("@"), unaryAtom, acceptToken(":"), unaryAtom),
				c.Seq(atom, acceptToken("@"), unaryAtom),
			),
		), atom)(input)
}

func unary(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("-"), acceptToken("#"))
	return c.Or(c.Fmap(mkUnaryOp, (c.And(op, index))), index)(input)
}

func unaryAtom(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("-"), acceptToken("#"))
	return c.Or(c.Fmap(mkUnaryOp, (c.And(op, atom))), atom)(input)
}

func divmul(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("*"), acceptToken("/"))
	return c.Fmap(mkLeftChain, c.And(unary, c.Any(c.And(op, unary))))(input)
}

func addsub(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("+"), acceptToken("-"))
	return c.Fmap(mkLeftChain, c.And(divmul, c.Any(c.And(op, divmul))))(input)
}

func logic(input c.RollbackLexer) ([]c.Node, error) {
	op := c.Or(acceptToken("&"), acceptToken("|"))
	return c.Fmap(mkLeftChain, c.And(addsub, c.Any(c.And(op, addsub))))(input)
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
	return c.Fmap(mkLeftChain, c.And(logic, c.Any(c.And(relOp, logic))))(input)
}

func expression(input c.RollbackLexer) ([]c.Node, error) {
	return relational(input)
}

func assignment(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkLeftChain, c.Seq(varName, acceptToken("="), expression))(input)
}

func statement(input c.RollbackLexer) ([]c.Node, error) {
	return c.OneOf(conditional, loop, returning, assignment, expression)(input)
}

func conditional(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkIf,
		c.Seq(acceptToken("if"), expression, c.Or(c.Seq(block, acceptToken("else"), block), block)))(input)
}

func loop(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkWhile, c.Seq(acceptToken("while"), expression, block))(input)
}

func returning(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkReturn, c.And(acceptToken("return"), expression))(input)
}

func function(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkFunction, c.Seq(parameters, acceptToken("->"), block))(input)
}

var parameters = c.Fmap(mkList,
	c.SurroundedBy(
		acceptToken("("),
		c.SeparatedBy(varName, acceptToken(",")),
		acceptToken(")")))

func call(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkFCall, c.Seq(varName, arguments))(input)
}

func arguments(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkList,
		c.SurroundedBy(
			acceptToken("("),
			c.SeparatedBy(expression, acceptToken(",")),
			acceptToken(")")))(input)
}

var eol = acceptTerm(token.EOL, "end of line")
var eof = acceptTerm(token.EOF, "end of file")

func block(input c.RollbackLexer) ([]c.Node, error) {
	return c.Or(
		c.Fmap(mkBlock,
			c.SurroundedBy(
				c.And(acceptToken("{"), eol),
				c.JoinedWith(statement, eol),
				acceptToken("}"))),
		statement)(input)
}

var program = c.Fmap(allButLast, c.And(c.JoinedWith(block, eol), eof))

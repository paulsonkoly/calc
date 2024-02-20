package parser

import (
	c "github.com/paulsonkoly/calc/combinator"
	"github.com/paulsonkoly/calc/lexer"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/token"
)

var Keywords = [...]string{"if", "else", "while", "return", "true", "false"}

type Type struct{}

func (t Type) Parse(input string) ([]node.Type, error) {
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
	return c.Choose(
		c.Conditional{Gate: c.Assert(c.And(parameters, acceptToken("->"))), OnSuccess: function},
		c.Conditional{Gate: c.Assert(c.And(varName, acceptToken("("))), OnSuccess: call},
		c.Conditional{Gate: floatLit, OnSuccess: c.Ok()},
		c.Conditional{Gate: intLit, OnSuccess: c.Ok()},
		c.Conditional{Gate: acceptToken("true"), OnSuccess: c.Ok()},
		c.Conditional{Gate: acceptToken("false"), OnSuccess: c.Ok()},
		c.Conditional{Gate: stringLit, OnSuccess: c.Ok()},
		c.Conditional{Gate: c.Assert(acceptToken("[")), OnSuccess: arrayLit},
		c.Conditional{Gate: c.Assert(acceptToken("(")), OnSuccess: paren},
		c.Conditional{Gate: c.Ok(), OnSuccess: varName})(input)
}

func index(input c.RollbackLexer) ([]c.Node, error) {
	return c.OneOf(
		c.Fmap(mkIndex,
			c.OneOf(
				c.Seq(atom, acceptToken("@"), unaryAtom, acceptToken(":"), unaryAtom),
				c.Seq(atom, acceptToken("@"), unaryAtom),
			),
		), atom)(input)
}

func unary(input c.RollbackLexer) ([]c.Node, error) {
	op := c.OneOf(acceptToken("-"), acceptToken("#"))
	return c.OneOf(c.Fmap(mkUnaryOp, (c.And(op, index))), index)(input)
}

func unaryAtom(input c.RollbackLexer) ([]c.Node, error) {
	op := c.OneOf(acceptToken("-"), acceptToken("#"))
	return c.OneOf(c.Fmap(mkUnaryOp, (c.And(op, atom))), atom)(input)
}

func divmul(input c.RollbackLexer) ([]c.Node, error) {
	op := c.OneOf(acceptToken("*"), acceptToken("/"))
	chain := c.Any(c.Conditional{Gate: op, OnSuccess: unary})
	return c.Fmap(mkLeftChain, c.And(unary, chain))(input)
}

func addsub(input c.RollbackLexer) ([]c.Node, error) {
	op := c.OneOf(acceptToken("+"), acceptToken("-"))
	chain := c.Any(c.Conditional{Gate: op, OnSuccess: divmul})
	return c.Fmap(mkLeftChain, c.And(divmul, chain))(input)
}

func logic(input c.RollbackLexer) ([]c.Node, error) {
	op := c.OneOf(acceptToken("&"), acceptToken("|"))
	chain := c.Any(c.Conditional{Gate: op, OnSuccess: addsub})
	return c.Fmap(mkLeftChain, c.And(addsub, chain))(input)
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
	chain := c.Any(c.Conditional{Gate: relOp, OnSuccess: logic})
	return c.Fmap(mkLeftChain, c.And(logic, chain))(input)
}

func expression(input c.RollbackLexer) ([]c.Node, error) {
	return relational(input)
}

func assignment(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkLeftChain, c.Seq(varName, acceptToken("="), expression))(input)
}

func statement(input c.RollbackLexer) ([]c.Node, error) {
	return c.Choose(
		c.Conditional{Gate: c.Assert(acceptToken("if")), OnSuccess: conditional},
		c.Conditional{Gate: c.Assert(acceptToken("while")), OnSuccess: loop},
		c.Conditional{Gate: c.Assert(acceptToken("return")), OnSuccess: returning},
		c.Conditional{Gate: c.Assert(c.And(varName, acceptToken("="))), OnSuccess: assignment},
		c.Conditional{Gate: c.Ok(), OnSuccess: expression})(input)
}

func conditional(input c.RollbackLexer) ([]c.Node, error) {
	return c.Fmap(mkIf,
		c.Seq(
			acceptToken("if"),
			expression,
			block,
			c.Choose(
				c.Conditional{Gate: acceptToken("else"), OnSuccess: block},
				c.Conditional{Gate: c.Ok(), OnSuccess: c.Ok()}),
		))(input)
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

var eol = c.Fmap(func(n []c.Node) []c.Node { return []c.Node{} }, acceptTerm(token.EOL, "end of line"))
var eof = c.Fmap(func(n []c.Node) []c.Node { return []c.Node{} }, acceptTerm(token.EOF, "end of file"))

func statements(input c.RollbackLexer) ([]c.Node, error) {
	pred := c.Assert(c.And(eol, c.Not(acceptToken("}"))))
	return c.And(statement, c.Any(c.Conditional{Gate: pred, OnSuccess: c.And(eol, statement)}))(input)
}

func block(input c.RollbackLexer) ([]c.Node, error) {
	return c.Choose(
		c.Conditional{Gate: c.Assert(acceptToken("{")),
			OnSuccess: c.Fmap(mkBlock,
				c.SurroundedBy(
					c.And(acceptToken("{"), eol),
					statements,
					c.And(eol, acceptToken("}"))))},
		c.Conditional{Gate: c.Ok(), OnSuccess: statement})(input)
}

var program = c.Seq(
	block,
	c.Any(c.Conditional{Gate: c.Assert(c.And(eol, c.Not(eof))), OnSuccess: c.And(eol, block)}),
	eof)

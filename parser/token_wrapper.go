package parser

import (
	"log"

	"github.com/phaul/calc/combinator"
	"github.com/phaul/calc/types/node"
	"github.com/phaul/calc/types/token"
)

// tokenWrapper wraps a single lexeme/token in AST node. We basically translate
// a linear structure of stream of tokens into a tree structure, although at
// this point we don't have a tree structure we are just creating leafs.

type tokenWrapper struct{}

func (_ tokenWrapper) Wrap(t combinator.Token) combinator.Node {
	realT := t.(token.Type)

	switch realT.Type {
	case token.IntLit:
		return node.Int(realT.Value)

	case token.FloatLit:
		return node.Float(realT.Value)

	case token.Sticky:
		return node.BinOp{Op: realT.Value}

	case token.Name:
		return node.Name(realT.Value)

	case token.NotSticky, token.EOF, token.EOL:
		return node.Invalid{}

	default:
		log.Panicf("unexpected token type to wrap %v", realT)
	}
	panic("unreachable code")
}

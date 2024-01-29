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
	var kind node.NodeType

	switch realT.Type {
	case token.IntLit:
		kind = node.Int

	case token.FloatLit:
		kind = node.Float

	case token.Sticky:
		switch realT.Value {
		case "+", "-", "*", "/", "|", "&", "<", ">", "<=", ">=", "==", "!=", "->", "=":
			kind = node.Op
		default:
			log.Panicf("unexpected sticky token to wrap %s", realT.Value)
		}

	case token.NotSticky:
		kind = node.Invalid

	case token.Name:
		kind = node.Name

	case token.EOF, token.EOL:
		kind = node.Invalid

	default:
		log.Panicf("unexpected token type to wrap %v", realT)
	}
	return node.Type{Token: realT.Value, Kind: kind}
}

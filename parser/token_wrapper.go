package parser

import (
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/paulsonkoly/calc/combinator"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/token"
)

// tokenWrapper wraps a single lexeme/token in AST node. We basically translate
// a linear structure of stream of tokens into a tree structure, although at
// this point we don't have a tree structure we are just creating leafs.

type tokenWrapper struct{}

var ops = [...]string{"+", "-", "*", "/", "<", ">", "<=", ">=", "==", "!=", "&&", "||", "&", "|", "%", "#", "~", ":", "!"}

func (tokenWrapper) Wrap(t combinator.Token) combinator.Node {
	realT := t.(token.Type)

	switch realT.Type {
	case token.IntLit:
		x, err := strconv.Atoi(realT.Value)
		if err != nil {
			panic(err)
		}
		return node.Int(x)

	case token.FloatLit:
		x, err := strconv.ParseFloat(realT.Value, 64)
		if err != nil {
			panic(err)
		}
		return node.Float(x)

	case token.StringLit:
		s := realT.Value

		// remove the first and last quotes, and replace escaped quotes with quotes
		s = strings.ReplaceAll(s, "\\\"", "\"")
		s = s[1 : len(s)-1]
		return node.String(s)

	case token.Sticky, token.NotSticky:
		if slices.Contains(ops[:], realT.Value) {
			return node.BinOp{Op: realT.Value}
		}

		return node.Invalid{}

	case token.Name:
		switch realT.Value {
		case "true":
			return node.Bool(true)
		case "false":
			return node.Bool(false)
		default:
			return node.Name(realT.Value)
		}

	case token.EOF, token.EOL:
		return node.Invalid{}

	default:
		log.Panicf("unexpected token type to wrap %v", realT)
	}
	panic("unreachable code")
}

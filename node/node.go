package node

import "github.com/phaul/calc/lexer"

type Node struct {
	Token    lexer.Token
	Children []Node
}

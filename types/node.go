package types

import (
	"fmt"
	"strings"
)

type Node struct {
	Token    Token
	Children []Node
}

func (n Node) PrettyPrint() { recurse(0, n) }

func recurse(depth int, n Node) {
	fmt.Printf("%s %v", strings.Repeat(" ", 3*depth), n.Token)
	for _, c := range n.Children {
		recurse(depth+1, c)
	}
}

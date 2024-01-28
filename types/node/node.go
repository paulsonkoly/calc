package node

import (
	"fmt"
	"strings"

	"github.com/phaul/calc/types/token"
)

type Type struct {
	Token    token.Type
	Children []Type
}

func (n Type) PrettyPrint() { recurse(0, n) }

func recurse(depth int, n Type) {
	fmt.Printf("%s %v\n", strings.Repeat(" ", 3*depth), n.Token)
	for _, c := range n.Children {
		recurse(depth+1, c)
	}
}

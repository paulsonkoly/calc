// node is an abstract syntax tree (AST) node
package node

import (
	"fmt"
	"strings"
)

// NodeType represents the different types of AST nodes
type NodeType int

const (
	Invalid = NodeType(iota)
	Call    // Call is function call
	Int     // Int is integer literal
	Float   // Float is float literal
	Op      // Op is an operator of any kind, anything from "=", "->",  etc.
	If      // If is a conditional construct
	While   // While is a loop construct
	Return  // Return is a return statement
	Name    // Variable name (also "true", "false" etc.)
	Block   // Block is a code block / sequence that was in '{', '}'
)

// Type is AST node type
type Type struct {
	Token    string   // Token is the literal token that the node was derived from
	Kind     NodeType // Kind is the kind of node
	Children []Type   // Children are the child nodes
}

func (n Type) PrettyPrint() { recurse(0, n) }

func recurse(depth int, n Type) {
	fmt.Printf("%s %v\n", strings.Repeat(" ", 3*depth), n)
	for _, c := range n.Children {
		recurse(depth+1, c)
	}
}

func (t Type) String() string {
	return fmt.Sprintf("[\"%s\" %v]", t.Token, t.Kind)
}

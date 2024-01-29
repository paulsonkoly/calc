package parser

import (
	"log"

	c "github.com/phaul/calc/combinator"
	"github.com/phaul/calc/types/node"
)

// Node transformations. We receive a parsed linear sequence of nodes, arrange it in sub-trees.

// mkUnaryOp is used for unary operators
//
// It rewrites the pair of nodes putting the second under the first.
func mkUnaryOp(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		log.Panicf("incorrect number of sub nodes for unary operator (%d)", len(nodes))
	}

	r := nodes[0].(node.Type)
	r.Children = []node.Type{nodes[1].(node.Type)}
	return []c.Node{r}
}

// mkReturn is for return statements
func mkReturn(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		log.Panicf("incorrect number of sub nodes for return (%d)", len(nodes))
	}

	r := node.Type{Kind: node.Return, Children: []node.Type{nodes[1].(node.Type)}}
	return []c.Node{r}
}

// mkLeftChain rewrites a sequence of binary operators applied on operands in a
// left assictive structure
//
// In effect it arranges a+b+c sequence in:
//
//	+
//	|-+
//	| |-a
//	| `-b
//	`-c
func mkLeftChain(nodes []c.Node) []c.Node {
	if len(nodes) < 3 || len(nodes)%2 == 0 {
		log.Panicf("incorrect number of sub nodes for left chain (%d)", len(nodes))
	}
	r := nodes[0]
	for i := 1; i+1 < len(nodes); i += 2 {
		n := nodes[i].(node.Type)
		n.Children = []node.Type{r.(node.Type), nodes[i+1].(node.Type)}
		r = n
	}
	return []c.Node{r}
}

func wrap(nodes []c.Node) []c.Node {
	r := node.Type{Children: make([]node.Type, 0)}
	for _, n := range nodes {
		r.Children = append(r.Children, n.(node.Type))
	}
	return []c.Node{r}
}

// mkBlock wraps a sequence of nodes in a single block node
func mkBlock(nodes []c.Node) []c.Node {
	r := node.Type{Kind: node.Block, Children: make([]node.Type, 0)}
	for _, n := range nodes {
		r.Children = append(r.Children, n.(node.Type))
	}
	return []c.Node{r}
}

// mkFCall creates a function call node
func mkFCall(nodes []c.Node) []c.Node {
	r := node.Type{Kind: node.Call, Children: []node.Type{nodes[0].(node.Type), nodes[1].(node.Type)}}
	return []c.Node{r}
}

// mkIf creates a conditional structure
func mkIf(nodes []c.Node) []c.Node {
	if len(nodes) != 3 && len(nodes) != 5 {
		log.Panicf("incorrect number of sub nodes for if (%d)", len(nodes))
	}
	r := node.Type{Kind: node.If}
	r.Children = []node.Type{nodes[1].(node.Type), nodes[2].(node.Type)}
	if len(nodes) == 5 {
		r.Children = append(r.Children, nodes[4].(node.Type))
	}
	return []c.Node{r}
}

// mkWhile creates a while loop structure
func mkWhile(nodes []c.Node) []c.Node {
	if len(nodes) != 3 {
		log.Panicf("incorrect number of sub nodes for while (%d)", len(nodes))
	}
	r := node.Type{Kind: node.While}
	r.Children = []node.Type{nodes[1].(node.Type), nodes[2].(node.Type)}
	return []c.Node{r}
}

package parser

import (
	"log"

	c "github.com/paulsonkoly/calc/combinator"
	"github.com/paulsonkoly/calc/types/node"
)

// Node transformations. We receive a parsed linear sequence of nodes, arrange it in sub-trees.

// mkUnaryOp is used for unary operators
//
// It rewrites the pair of nodes putting the second under the first.
func mkUnaryOp(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		log.Panicf("incorrect number of sub nodes for unary operator (%d)", len(nodes))
	}
	n := node.UnOp{Op: nodes[0].(node.BinOp).Op, Target: nodes[1].(node.Type)}

	return []c.Node{n}
}

// mkReturn is for return statements
func mkReturn(nodes []c.Node) []c.Node {
	if len(nodes) != 2 {
		log.Panicf("incorrect number of sub nodes for return (%d)", len(nodes))
	}
	n := node.Return{Target: nodes[1].(node.Type)}
	return []c.Node{n}
}

// mkYield is for yield statements
func mkYield(nodes []c.Node) []c.Node {
	if len(nodes)!= 2 {
		log.Panicf("incorrect number of sub nodes for yield (%d)", len(nodes))
	}
	n := node.Yield{Target: nodes[1].(node.Type)}
	return []c.Node{n}
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
	if len(nodes)%2 == 0 {
		log.Panicf("incorrect number of sub nodes for left chain (%d)", len(nodes))
	}
	r := nodes[0]
	for i := 1; i+1 < len(nodes); i += 2 {
		n := nodes[i].(node.BinOp)
		n.Left = r.(node.Type)
		n.Right = nodes[i+1].(node.Type)
		r = n
	}
	return []c.Node{r}
}

// mkAssign creates an assigment to a variable
func mkAssign(nodes []c.Node) []c.Node {
	if len(nodes) != 3 {
		log.Panicf("incorrect number of sub nodes for assignment (%d)", len(nodes))
	}
	assign := node.Assign{VarRef: nodes[0].(node.Type), Value: nodes[2].(node.Type)}
	return []c.Node{assign}
}

// mkIndex rewrites a sequence describing an array indexing into an Index node
func mkIndex(nodes []c.Node) []c.Node {
	switch len(nodes) {

	case 3: // ary[from:to]
		return []c.Node{node.IndexFromTo{Ary: nodes[0].(node.Type), From: nodes[1].(node.Type), To: nodes[2].(node.Type)}}

	case 2: // ary[at]
		return []c.Node{node.IndexAt{Ary: nodes[0].(node.Type), At: nodes[1].(node.Type)}}

  case 1:
    return nodes

	default:
		log.Panicf("incorrect number of sub nodes for indexing (%d)", len(nodes))
	}
	panic("unreachable code")
}

func mkList(nodes []c.Node) []c.Node {
	r := node.List{Elems: make([]node.Type, 0)}
	for _, n := range nodes {
		r.Elems = append(r.Elems, n.(node.Type))
	}
	return []c.Node{r}
}

// mkBlock wraps a sequence of nodes in a single block node
func mkBlock(nodes []c.Node) []c.Node {
	if len(nodes) <= 1 {
		return nodes
	}
	r := node.Block{Body: make([]node.Type, 0)}
	for _, n := range nodes {
		r.Body = append(r.Body, n.(node.Type))
	}
	return []c.Node{r}
}

// mkFCall creates a function call node
func mkFCall(nodes []c.Node) []c.Node {
	r := node.Call{Name: nodes[0].(node.Type), Arguments: nodes[1].(node.List)}
	return []c.Node{r}
}

func mkFunction(nodes []c.Node) []c.Node {
	if len(nodes) != 3 {
		log.Panicf("incorrect number of sub nodes for function (%d)", len(nodes))
	}

	r := node.Function{Parameters: nodes[0].(node.List), Body: nodes[2].(node.Type)}
	return []c.Node{r}
}

// mkIf creates a conditional structure
func mkIf(nodes []c.Node) []c.Node {
	var n node.Type
	switch len(nodes) {
	case 3:
		n = node.If{Condition: nodes[1].(node.Type), TrueCase: nodes[2].(node.Type)}
	case 5:
		n = node.IfElse{Condition: nodes[1].(node.Type), TrueCase: nodes[2].(node.Type), FalseCase: nodes[4].(node.Type)}
	default:
		log.Panicf("incorrect number of sub nodes for if (%d)", len(nodes))
	}
	return []c.Node{n}
}

// mkWhile creates a while loop structure
func mkWhile(nodes []c.Node) []c.Node {
	if len(nodes) != 3 {
		log.Panicf("incorrect number of sub nodes for while (%d)", len(nodes))
	}
	n := node.While{Condition: nodes[1].(node.Type), Body: nodes[2].(node.Type)}
	return []c.Node{n}
}

// mkFor creates a for loop structure
func mkFor(nodes []c.Node) []c.Node {
    if len(nodes) !=5 {
        log.Panicf("incorrect number of sub nodes for for (%d)", len(nodes))
    }
    n := node.For{VarRef: nodes[1].(node.Type), Iterator: nodes[3].(node.Type), Body: nodes[4].(node.Type)}
    return []c.Node{n}
}

// Package builtin contains the builtin functions.
package builtin

import (
	"github.com/paulsonkoly/calc/types/compresult"
	"github.com/paulsonkoly/calc/types/node"
)

// Load compiles the built in functions and adds them to cr.
func Load(cr compresult.Type) {
	for _, fun := range all {
		fNode := fun.STRewrite(node.SymTbl{})
		node.ByteCodeNoStck(fNode, cr)
	}
}

var all = [...]node.Assign{
	readF,
	writeF,
	atonF,
	toaF,
	exitF,
	fromToF,
	indicesF,
	elemsF,
}

var readF = node.Assign{VarRef: node.Name("read"), Value: node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Read{}}}

var writeF = node.Assign{VarRef: node.Name("write"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Write{Value: v}}}

var atonF = node.Assign{VarRef: node.Name("aton"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Aton{Value: v}}}

var toaF = node.Assign{VarRef: node.Name("toa"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Toa{Value: v}}}

var exitF = node.Assign{VarRef: node.Name("exit"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Exit{Value: v}}}

var fromToF = node.Assign{
	VarRef: node.Name("fromto"),
	Value: node.Function{
		Parameters: node.List{Elems: []node.Type{a, b}},
		Body: node.While{
			Condition: node.BinOp{Op: "<", Left: a, Right: b},
			Body: node.Block{
				Body: []node.Type{
					node.Yield{Target: a},
					node.Assign{VarRef: node.Name("a"), Value: node.BinOp{Op: "+", Left: a, Right: node.Int(1)}},
				},
			},
		},
	},
}

var indicesF = node.Assign{
	VarRef: node.Name("indices"),
	Value: node.Function{
		Parameters: node.List{Elems: []node.Type{a}},
		Body: node.Block{
			Body: []node.Type{
				node.Assign{VarRef: node.Name("i"), Value: node.Int(0)},
				node.While{
					Condition: node.BinOp{Op: "<", Left: node.Name("i"), Right: node.UnOp{Op: "#", Target: a}},
					Body: node.Block{
						Body: []node.Type{
							node.Yield{Target: node.Name("i")},
							node.Assign{VarRef: node.Name("i"), Value: node.BinOp{Op: "+", Left: node.Name("i"), Right: node.Int(1)}},
						},
					},
				},
			},
		},
	},
}

var elemsF = node.Assign{
	VarRef: node.Name("elems"),
	Value: node.Function{
		Parameters: node.List{Elems: []node.Type{a}},
		Body: node.Block{
			Body: []node.Type{
				node.Assign{VarRef: node.Name("i"), Value: node.Int(0)},
				node.While{
					Condition: node.BinOp{Op: "<", Left: node.Name("i"), Right: node.UnOp{Op: "#", Target: a}},
					Body: node.Block{
						Body: []node.Type{
							node.Yield{Target: node.IndexAt{Ary: a, At: node.Name("i")}},
							node.Assign{VarRef: node.Name("i"), Value: node.BinOp{Op: "+", Left: node.Name("i"), Right: node.Int(1)}},
						},
					},
				},
			},
		},
	},
}

var v = node.Name("v")
var a = node.Name("a")
var b = node.Name("b")

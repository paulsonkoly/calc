package builtin

import (
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
)

func Load(cs *[]bytecode.Type, ds *[]value.Type) {
	for _, fun := range all {
		fNode := fun.STRewrite(node.SymTbl{})
		node.ByteCodeNoStck(fNode, cs, ds)
	}
}

var all = [...]node.Assign{
	readF,
	writeF,
	atonF,
	toaF,
	errorF,
	exitF,
	fromToF,
	indicesF,
	elemsF,
}

var readF = node.Assign{VarRef: node.Name("read"), Value: node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Read{}}}

var writeF = node.Assign{VarRef: node.Name("write"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Write{Value: v}}}

var atonF = node.Assign{VarRef: node.Name("aton"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Aton{Value: v}}}

var toaF = node.Assign{VarRef: node.Name("toa"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Toa{Value: v}}}

var errorF = node.Assign{VarRef: node.Name("error"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Error{Value: v}}}

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

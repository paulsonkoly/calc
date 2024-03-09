package builtin

import (
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/types/node"
)

func Load(m *memory.Type) {
	for name, fun := range all {
		fNode := fun.STRewrite(node.SymTbl{})
    // TODO, should we just use value instead of node?
		fVal, _ := fNode.Evaluate(m, nil, nil)
		m.SetGlobal(name, fVal)
	}
}

var all = map[string]node.Function{
	"read":  readF,
	"write": writeF,
	"aton":  atonF,
	"error": errorF,
	"toa":   toaF,
	"exit":  exitF,
}

var readF = node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Read{}}

var writeF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Write{Value: v}}

var atonF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Aton{Value: v}}

var toaF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Toa{Value: v}}

var errorF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Error{Value: v}}

var exitF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Exit{Value: v}}

var v = node.Name("v")


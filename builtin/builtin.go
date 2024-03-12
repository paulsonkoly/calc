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
	// "read":  readF,
	writeF,
	atonF,
	toaF,
	// "error": errorF,
	// "exit":  exitF,
}

var readF = node.Assign{VarRef: node.Name("read"), Value: node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Read{}}}

var writeF = node.Assign{VarRef: node.Name("write"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Write{Value: v}}}

var atonF = node.Assign{ VarRef: node.Name("aton"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Aton{Value: v}}}

var toaF = node.Assign{VarRef: node.Name("toa"), Value: node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Toa{Value: v}}}

var errorF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Error{Value: v}}

var exitF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Exit{Value: v}}

var v = node.Name("v")

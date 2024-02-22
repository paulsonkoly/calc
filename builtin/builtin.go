package builtin

import (
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/node"
)

func Load(m *memory.Type) {
	for name, fun := range all {
		fNode := fun.STRewrite(node.SymTbl{})
		fVal, _ := fNode.Evaluate(m)
		m.SetGlobal(name, fVal)
  }
}

var all = map[string]node.Function{"read": readF, "write": writeF, "aton": atonF, "repl": replF, "error": errorF, "toa": toaF}

var readF = node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Read{}}

var writeF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Write{Value: v}}

var atonF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Aton{Value: v}}

var toaF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Toa{Value: v}}

var parseInstance = parser.Type{}
var replF = node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Repl{Parser: parseInstance}}

var errorF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Error{Value: v}}

var v = node.Name("v")

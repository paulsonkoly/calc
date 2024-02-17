package builtin

import (
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
)

var All = map[string]value.Type{"read": Read, "write": Write}

var readF = node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Read{}}
var Read = value.Function{Node: &readF, Frame: nil}

var writeF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Write{Value: v}}
var Write = value.Function{Node: &writeF, Frame: nil}

var v = node.Name("v")

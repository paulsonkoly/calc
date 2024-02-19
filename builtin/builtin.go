package builtin

import (
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
)

type Type struct {}

func (_ Type)All() map[string]value.Type { return all }

var all = map[string]value.Type{"read": Read, "write": Write, "aton": Aton, "repl": Repl, "error": Error}

var readF = node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Read{}}
var Read = value.Function{Node: &readF, Frame: nil}

var writeF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Write{Value: v}}
var Write = value.Function{Node: &writeF, Frame: nil}

var atonF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Aton{Value: v}}
var Aton = value.Function{Node: &atonF, Frame: nil}

var parseInstance = parser.Type{}
var replF = node.Function{Parameters: node.List{Elems: []node.Type{}}, Body: node.Repl{ Parser: parseInstance}}
var Repl = value.Function{Node: &replF, Frame: nil}

var errorF = node.Function{Parameters: node.List{Elems: []node.Type{v}}, Body: node.Error{Value: v}}
var Error = value.Function{Node: &errorF, Frame: nil}

var v = node.Name("v")

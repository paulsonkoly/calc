package evaluator

import (
	"fmt"
	"strconv"

	"github.com/phaul/calc/types"
)

type Variables map[string]Value

func Evaluate(vars Variables, n types.Node) Value {
	switch n.Token.Type {

	case types.IntLit:
		i, err := strconv.Atoi(n.Token.Value)
		if err != nil {
			panic(err)
		}
		return ValueInt(i)

	case types.FloatLit:
		f, err := strconv.ParseFloat(n.Token.Value, 64)
		if err != nil {
			panic(err)
		}
		return ValueFloat(f)

	case types.Op:
		switch n.Token.Value {

		case "+", "-", "*", "/":
			// special case unary -
			if len(n.Children) == 1 {
				r := Evaluate(vars, n.Children[0])
				r = r.Op("*", ValueInt(-1))
				return r
			}
			return Evaluate(vars, n.Children[0]).Op(n.Token.Value, Evaluate(vars, n.Children[1]))

		case "=":
			v := Evaluate(vars, n.Children[1])
			vars[n.Children[0].Token.Value] = v
			return v
		}

	case types.VarName:
		var r Value
		if v, ok := vars[n.Token.Value]; ok {
			return v
		}
		_ = fmt.Errorf("variable %s not defined\n", n.Token.Value)
		return r
	}

	panic("unsupported node tpye")
}

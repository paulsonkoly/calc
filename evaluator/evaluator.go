package evaluator

import (
	"strconv"

	"github.com/phaul/calc/types"
)

func Evaluate(n types.Node) Value {
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

  case types.InvalidToken:
    switch (n.Children[1].Token.Value) {
    case "+":
      a := Evaluate(n.Children[0])
      b := Evaluate(n.Children[2])

      return a.Plus(b)
    }

	}
	panic("unsupported node tpye")
}

package evaluator

import (
	"fmt"
	"log"
	"strconv"

	"github.com/phaul/calc/types"
)

type Variables map[string]Value

var NoResultError = ValueError("no result")

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

	case types.Sticky:
		switch n.Token.Value {

		case "+", "-", "*", "/":
			// special case unary -
			if len(n.Children) == 1 {
				r := Evaluate(vars, n.Children[0])
				r = r.Arith("*", ValueInt(-1))
				return r
			}
			return Evaluate(vars, n.Children[0]).Arith(n.Token.Value, Evaluate(vars, n.Children[1]))

		case "<", "<=", ">", ">=", "==", "!=":
			return Evaluate(vars, n.Children[0]).Relational(n.Token.Value, Evaluate(vars, n.Children[1]))

		case "=":
			v := Evaluate(vars, n.Children[1])
			vars[n.Children[0].Token.Value] = v
			return v

		default:
			log.Panicf("unexpected single character in evaluator: %s", n.Token.Value)
		}

	case types.Name:
		switch n.Token.Value {

		case "true":
			return ValueBool(true)

		case "false":
			return ValueBool(false)

		case "if":
			c:= Evaluate(vars, n.Children[0])
			if cc, ok := c.(ValueBool); ok {
				if cc {
					return Evaluate(vars, n.Children[1])
				} else if len(n.Children) > 2 {
					return Evaluate(vars, n.Children[2])
				} else {
					return NoResultError
				}
			} else {
				return TypeError
			}

		case "while":
			r := Value(NoResultError)
			for {
				cond := Evaluate(vars, n.Children[0])
				if ccond, ok := cond.(ValueBool); ok {
					if ! ccond {
						return r
					}
					r = Evaluate(vars, n.Children[1])
				} else {
					return TypeError
				}

			}

		default:
			if v, ok := vars[n.Token.Value]; ok {
				return v
			}
			return ValueError(fmt.Sprintf("variable %s not defined", n.Token.Value))
		}
	default:
		if len(n.Children) < 1 {
			log.Panic("empty block")
		}
		r := Evaluate(vars, n.Children[0])
		for _, e := range n.Children[1:] {
			r = Evaluate(vars, e)
		}
		return r
	}

	panic("unsupported node tpye")
}

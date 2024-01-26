package evaluator

import (
	"log"
	"strconv"

	"github.com/phaul/calc/stack"
	"github.com/phaul/calc/types"
)

func Evaluate(s stack.Stack, n types.Node) types.Value {
	switch n.Token.Type {

	case types.IntLit:
		i, err := strconv.Atoi(n.Token.Value)
		if err != nil {
			panic(err)
		}
		return types.ValueInt(i)

	case types.FloatLit:
		f, err := strconv.ParseFloat(n.Token.Value, 64)
		if err != nil {
			panic(err)
		}
		return types.ValueFloat(f)

	case types.Call:
		fName := n.Children[0].Token.Value
		if fl, ok := s.LookUp(fName); ok {
			s.Push()
			if f, ok := fl.(types.ValueFunction); ok {
				args := n.Children[1].Children
				params := f.Children[0].Children
				if len(args) == len(params) {
					s.Push()
					for i := 0; i < len(args); i++ {
						s.Set(params[i].Token.Value, Evaluate(s, args[i]))
					}
					r := Evaluate(s, types.Node(f.Children[1]))
					s.Pop()
					return r
				} else {
					return types.ArgumentError
				}
			} else {
				return types.TypeError
			}
		} else {
			return fl
		}

	case types.Sticky:
		switch n.Token.Value {

		case "+", "-", "*", "/":
			// special case unary -
			if len(n.Children) == 1 {
				r := Evaluate(s, n.Children[0])
				r = r.Arith("*", types.ValueInt(-1))
				return r
			}
			return Evaluate(s, n.Children[0]).Arith(n.Token.Value, Evaluate(s, n.Children[1]))

		case "&", "|":
			return Evaluate(s, n.Children[0]).Logic(n.Token.Value, Evaluate(s, n.Children[1]))

		case "<", "<=", ">", ">=", "==", "!=":
			return Evaluate(s, n.Children[0]).Relational(n.Token.Value, Evaluate(s, n.Children[1]))

		case "->":
			return types.ValueFunction(n)

		case "=":
			v := Evaluate(s, n.Children[1])
			s.Set(n.Children[0].Token.Value, v)
			return v

		default:
			log.Panicf("unexpected single character in evaluator: %s", n.Token.Value)
		}

	case types.Name:
		switch n.Token.Value {

		case "true":
			return types.ValueBool(true)

		case "false":
			return types.ValueBool(false)

		case "if":
			c := Evaluate(s, n.Children[0])
			if cc, ok := c.(types.ValueBool); ok {
				if cc {
					return Evaluate(s, n.Children[1])
				} else if len(n.Children) > 2 {
					return Evaluate(s, n.Children[2])
				} else {
					return types.NoResultError
				}
			} else {
				return types.TypeError
			}

		case "while":
			r := types.Value(types.NoResultError)
			for {
				cond := Evaluate(s, n.Children[0])
				if ccond, ok := cond.(types.ValueBool); ok {
					if !ccond {
						return r
					}
					r = Evaluate(s, n.Children[1])
				} else {
					return types.TypeError
				}

			}

		default:
			v, _ := s.LookUp(n.Token.Value)
			return v
		}
	default:
		if len(n.Children) < 1 {
			log.Panic("empty block")
		}
		r := Evaluate(s, n.Children[0])
		for _, e := range n.Children[1:] {
			r = Evaluate(s, e)
		}
		return r
	}

	panic("unsupported node tpye")
}

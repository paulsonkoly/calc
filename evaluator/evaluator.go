package evaluator

import (
	"log"
	"strconv"

	"github.com/phaul/calc/stack"
	"github.com/phaul/calc/types"
)

func Evaluate(s stack.Stack, n types.Node) types.Value {
	r, _ := evaluate(s, n)
	return r
}

func evaluate(s stack.Stack, n types.Node) (types.Value, bool) {
	switch n.Token.Type {

	case types.IntLit:
		i, err := strconv.Atoi(n.Token.Value)
		if err != nil {
			panic(err)
		}
		return types.ValueInt(i), false

	case types.FloatLit:
		f, err := strconv.ParseFloat(n.Token.Value, 64)
		if err != nil {
			panic(err)
		}
		return types.ValueFloat(f), false

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
					return r, false
				} else {
					return types.ArgumentError, false
				}
			} else {
				return types.TypeError, false
			}
		} else {
			return fl, false
		}

	case types.Sticky:
		switch n.Token.Value {

		case "+", "-", "*", "/":
			// special case unary -
			if len(n.Children) == 1 {
				r := Evaluate(s, n.Children[0])
				r = r.Arith("*", types.ValueInt(-1))
				return r, false
			}
			return Evaluate(s, n.Children[0]).Arith(n.Token.Value, Evaluate(s, n.Children[1])), false

		case "&", "|":
			return Evaluate(s, n.Children[0]).Logic(n.Token.Value, Evaluate(s, n.Children[1])), false

		case "<", "<=", ">", ">=", "==", "!=":
			return Evaluate(s, n.Children[0]).Relational(n.Token.Value, Evaluate(s, n.Children[1])), false

		case "->":
			return types.ValueFunction(n), false

		case "=":
			v := Evaluate(s, n.Children[1])
			s.Set(n.Children[0].Token.Value, v)
			return v, false

		default:
			log.Panicf("unexpected single character in evaluator: %s", n.Token.Value)
		}

	case types.Name:
		switch n.Token.Value {

		case "true":
			return types.ValueBool(true), false

		case "false":
			return types.ValueBool(false), false

		case "if":
			c := Evaluate(s, n.Children[0])
			if cc, ok := c.(types.ValueBool); ok {
				if cc {
					return evaluate(s, n.Children[1])
				} else if len(n.Children) > 2 {
					return evaluate(s, n.Children[2])
				} else {
					return types.NoResultError, false
				}
			} else {
				return types.TypeError, false
			}

		case "while":
			r := types.Value(types.NoResultError)
			returning := false
			for {
				cond := Evaluate(s, n.Children[0])
				if ccond, ok := cond.(types.ValueBool); ok {
					if !bool(ccond) || returning {
						return r, returning
					}
					r, returning = evaluate(s, n.Children[1])
				} else {
					return types.TypeError, false
				}
			}

		case "return":
			return Evaluate(s, n.Children[0]), true

		default:
			v, _ := s.LookUp(n.Token.Value)
			return v, false
		}
	default:
		if len(n.Children) < 1 {
			log.Panic("empty block")
		}
		r, returning := evaluate(s, n.Children[0])
		for i := 1; i < len(n.Children) && !returning; i++ {
			r, returning = evaluate(s, n.Children[i])
		}
		return r, returning
	}

	panic("unsupported node tpye")
}

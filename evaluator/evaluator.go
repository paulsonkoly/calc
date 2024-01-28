package evaluator

import (
	"log"
	"strconv"

	"github.com/phaul/calc/stack"
	"github.com/phaul/calc/types"
	"github.com/phaul/calc/types/value"
)

func Evaluate(s stack.Stack, n types.Node) value.Type {
	r, _ := evaluate(s, n)
	return r
}

func evaluate(s stack.Stack, n types.Node) (value.Type, bool) {
	switch n.Token.Type {

	case types.IntLit:
		i, err := strconv.Atoi(n.Token.Value)
		if err != nil {
			panic(err)
		}
		return value.Int(i), false

	case types.FloatLit:
		f, err := strconv.ParseFloat(n.Token.Value, 64)
		if err != nil {
			panic(err)
		}
		return value.Float(f), false

	case types.Call:
		fName := n.Children[0].Token.Value
		if fl, ok := s.LookUp(fName); ok {
			if f, ok := fl.(value.Function); ok {
				args := n.Children[1].Children
				params := f.Node.Children[0].Children
				if len(args) == len(params) {
					// push 2 frames, one is the closure environment, the other is the frame for arguments
					// the arguments have to be evaluated before we push anything on the stack because what we push
					// ie the closure frame migth contain variables that affect the argument evaluation
					frm := make(value.Frame)
					for i := 0; i < len(args); i++ {
						frm[params[i].Token.Value] = Evaluate(s, args[i])
					}
					if f.Frame != nil {
						s.Push(f.Frame)
					}
					s.Push(&frm)
					r := Evaluate(s, types.Node(f.Node.Children[1]))
					if f.Frame != nil {
						s.Pop()
					}
					s.Pop()
					return r, false
				} else {
					return value.ArgumentError, false
				}
			} else {
				return value.TypeError, false
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
				r = r.Arith("*", value.Int(-1))
				return r, false
			}
			return Evaluate(s, n.Children[0]).Arith(n.Token.Value, Evaluate(s, n.Children[1])), false

		case "&", "|":
			return Evaluate(s, n.Children[0]).Logic(n.Token.Value, Evaluate(s, n.Children[1])), false

		case "<", "<=", ">", ">=", "==", "!=":
			return Evaluate(s, n.Children[0]).Relational(n.Token.Value, Evaluate(s, n.Children[1])), false

		case "->":
			if len(s) == 1 {
				return value.Function{Node: n}, false
			} else {
				return value.Function{Node: n, Frame: s.Top()}, false
			}

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
			return value.Bool(true), false

		case "false":
			return value.Bool(false), false

		case "if":
			c := Evaluate(s, n.Children[0])
			if cc, ok := c.(value.Bool); ok {
				if cc {
					return evaluate(s, n.Children[1])
				} else if len(n.Children) > 2 {
					return evaluate(s, n.Children[2])
				} else {
					return value.NoResultError, false
				}
			} else {
				return value.TypeError, false
			}

		case "while":
			r := value.Type(value.NoResultError)
			returning := false
			for {
				cond := Evaluate(s, n.Children[0])
				if ccond, ok := cond.(value.Bool); ok {
					if !bool(ccond) || returning {
						return r, returning
					}
					r, returning = evaluate(s, n.Children[1])
				} else {
					return value.TypeError, false
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

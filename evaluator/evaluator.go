package evaluator

import (
	"log"
	"strconv"

	"github.com/phaul/calc/stack"
	"github.com/phaul/calc/types/node"
	"github.com/phaul/calc/types/value"
)

func Evaluate(s stack.Stack, n node.Type) value.Type {
	r, _ := evaluate(s, n)
	return r
}

func evaluate(s stack.Stack, n node.Type) (value.Type, bool) {
	switch realN := n.(type) {

	case node.Int:
		i, err := strconv.Atoi(n.Token())
		if err != nil {
			panic(err)
		}
		return value.Int(i), false

	case node.Float:
		f, err := strconv.ParseFloat(n.Token(), 64)
		if err != nil {
			panic(err)
		}
		return value.Float(f), false

	case node.Call:
		fName := realN.Name
		if fl, ok := s.LookUp(fName); ok {
			if f, ok := fl.(value.Function); ok {
				args := realN.Arguments.Elems
				fNode := (*f.Node)
				params := fNode.Parameters.Elems
				if len(args) == len(params) {
					// push 2 frames, one is the closure environment, the other is the frame for arguments
					// the arguments have to be evaluated before we push anything on the stack because what we push
					// ie the closure frame migth contain variables that affect the argument evaluation
					frm := make(value.Frame)
					for i := 0; i < len(args); i++ {
						frm[params[i].Token()] = Evaluate(s, args[i])
					}
					if f.Frame != nil {
						s.Push(f.Frame)
					}
					s.Push(&frm)
					r := Evaluate(s, node.Type(fNode.Body))
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

	case node.UnOp:
		if n.Token() == "-" {
			r := Evaluate(s, realN.Target)
			r = r.Arith("*", value.Int(-1))
			return r, false
		} else {
			log.Panicf("unexpected unary op: %s\n", n.Token())
		}

	case node.BinOp:
		switch n.Token() {

		case "+", "-", "*", "/":
			return Evaluate(s, realN.Left).Arith(n.Token(), Evaluate(s, realN.Right)), false

		case "&", "|":
			return Evaluate(s, realN.Left).Logic(n.Token(), Evaluate(s, realN.Right)), false

		case "<", "<=", ">", ">=", "==", "!=":
			return Evaluate(s, realN.Left).Relational(n.Token(), Evaluate(s, realN.Right)), false

		case "=":
			v := Evaluate(s, realN.Right)
			s.Set(realN.Left.Token(), v)
			return v, false

		default:
			log.Panicf("unexpected single character in evaluator: %s", n.Token())
		}

	case node.Function:
		if len(s) == 1 {
			return value.Function{Node: &realN}, false
		} else {
			return value.Function{Node: &realN, Frame: s.Top()}, false
		}

	case node.Name:
		switch n.Token() {

		case "true":
			return value.Bool(true), false

		case "false":
			return value.Bool(false), false

		default:
			v, _ := s.LookUp(n.Token())
			return v, false
		}

	case node.If:
		c := Evaluate(s, realN.Condition)
		if cc, ok := c.(value.Bool); ok {
			if cc {
				return evaluate(s, realN.TrueCase)
			} else {
				return value.NoResultError, false
			}
		} else {
			return value.TypeError, false
		}

	case node.IfElse:
		c := Evaluate(s, realN.Condition)
		if cc, ok := c.(value.Bool); ok {
			if cc {
				return evaluate(s, realN.TrueCase)
			} else {
				return evaluate(s, realN.FalseCase)
			}
		} else {
			return value.TypeError, false
		}

	case node.While:
		r := value.Type(value.NoResultError)
		returning := false
		for {
			cond := Evaluate(s, realN.Condition)
			if ccond, ok := cond.(value.Bool); ok {
				if !bool(ccond) || returning {
					return r, returning
				}
				r, returning = evaluate(s, realN.Body)
			} else {
				return value.TypeError, false
			}
		}

	case node.Return:
		return Evaluate(s, realN.Target), true

	case node.Block:
		if len(realN.Body) < 1 {
			panic("empty block")
		}
		r, returning := evaluate(s, realN.Body[0])
		for i := 1; i < len(realN.Body) && !returning; i++ {
			r, returning = evaluate(s, realN.Body[i])
		}
		return r, returning
	}

	n.PrettyPrint(0)
	log.Panicf("unsupported node type %T", n)
	panic("unreachable code")
}

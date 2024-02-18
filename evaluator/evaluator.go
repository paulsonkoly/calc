package evaluator

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/paulsonkoly/calc/stack"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
)

// Evaluate evaluates the given AST node producing a Value
func Evaluate(s stack.Stack, n node.Type) value.Type {
	r, _ := wrap(n).Evaluate(s)
	return r
}

type Evaluator interface {
	Evaluate(s stack.Stack) (value.Type, bool)
}

type Int node.Int
type Float node.Float
type String node.String
type Array node.List
type Call node.Call
type UnOp node.UnOp
type BinOp node.BinOp
type IndexAt node.IndexAt
type IndexFromTo node.IndexFromTo
type Function node.Function
type Name node.Name
type If node.If
type IfElse node.IfElse
type While node.While
type Return node.Return
type Read node.Read
type Write node.Write
type Aton node.Aton
type Repl node.Repl
type Error node.Error
type Block node.Block

func wrap(n node.Type) Evaluator {
	switch n := n.(type) {
	case node.Int:
		return Int(n)
	case node.Float:
		return Float(n)
	case node.String:
		return String(n)
	case node.List:
		return Array(n)
	case node.Call:
		return Call(n)
	case node.UnOp:
		return UnOp(n)
	case node.BinOp:
		return BinOp(n)
	case node.IndexAt:
		return IndexAt(n)
	case node.IndexFromTo:
		return IndexFromTo(n)
	case node.Function:
		return Function(n)
	case node.Name:
		return Name(n)
	case node.If:
		return If(n)
	case node.IfElse:
		return IfElse(n)
	case node.While:
		return While(n)
	case node.Return:
		return Return(n)
	case node.Read:
		return Read(n)
	case node.Write:
		return Write(n)
	case node.Aton:
		return Aton(n)
	case node.Repl:
		return Repl(n)
	case node.Error:
		return Error(n)
	case node.Block:
		return Block(n)
	}

	n.PrettyPrint(0)
	log.Panicf("unsupported node type %T", n)
	panic("unreachable code")
}

func (i Int) Evaluate(s stack.Stack) (value.Type, bool) {
	x, err := strconv.Atoi(node.Int(i).Token())
	if err != nil {
		panic(err)
	}
	return value.Int(x), false
}

func (f Float) Evaluate(s stack.Stack) (value.Type, bool) {
	x, err := strconv.ParseFloat(node.Float(f).Token(), 64)
	if err != nil {
		panic(err)
	}
	return value.Float(x), false
}

func (s String) Evaluate(_ stack.Stack) (value.Type, bool) {
	tok := node.String(s).Token()
	// remove the first and last quotes, and replace escaped quotes with quotes
	tok = strings.ReplaceAll(tok, "\\\"", "\"")
	tok = tok[1 : len(tok)-1]
	return value.String(tok), false
}

func (a Array) Evaluate(s stack.Stack) (value.Type, bool) {
	elems := node.List(a).Elems
	evalElems := make(value.Array, 0)

	for _, e := range elems {
		evalElems = append(evalElems, Evaluate(s, e))
	}

	return evalElems, false
}

func (c Call) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.Call(c)
	fName := n.Name
	if fl, ok := s.LookUp(fName); ok {
		if f, ok := fl.(value.Function); ok {
			args := n.Arguments.Elems
			params := f.Node.Parameters.Elems
			if len(args) == len(params) {
				// push 2 frames, one is the closure environment, the other is the frame for arguments
				// the arguments have to be evaluated before we push anything on the stack because what we push
				// ie the closure frame might contain variables that affect the argument evaluation
				frm := make(value.Frame)
				for i := 0; i < len(args); i++ {
					frm[params[i].Token()] = Evaluate(s, args[i])
				}
				if f.Frame != nil {
					s.Push(f.Frame)
				}
				s.Push(&frm)
				r := Evaluate(s, node.Type(f.Node.Body))
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
}

func (u UnOp) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.UnOp(u)
	switch n.Token() {

	case "-":
		r := Evaluate(s, n.Target)
		r = r.Arith("*", value.Int(-1))
		return r, false

	case "#":
		r := Evaluate(s, n.Target)
		return r.Len(), false

	default:
		log.Panicf("unexpected unary op: %s\n", n.Token())
	}
	panic("unreachable code")
}

func (b BinOp) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.BinOp(b)
	switch n.Token() {

	case "+", "-", "*", "/":
		return Evaluate(s, n.Left).Arith(n.Token(), Evaluate(s, n.Right)), false

	case "&", "|":
		return Evaluate(s, n.Left).Logic(n.Token(), Evaluate(s, n.Right)), false

	case "==":
		return Evaluate(s, n.Left).Eq(Evaluate(s, n.Right)), false

	case "!=":
		eq := Evaluate(s, n.Left).Eq(Evaluate(s, n.Right))
		if eq, ok := eq.(value.Bool); ok {
			return value.Bool(!eq), false
		}
		return eq, false

	case "<", "<=", ">", ">=":
		return Evaluate(s, n.Left).Relational(n.Token(), Evaluate(s, n.Right)), false

	case "=":
		v := Evaluate(s, n.Right)
		s.Set(n.Left.Token(), v)
		return v, false

	default:
		log.Panicf("unexpected single character in evaluator: %s", n.Token())
	}
	panic("unreachable code")
}

func (i IndexAt) Evaluate(s stack.Stack) (value.Type, bool) {
	ary := Evaluate(s, i.Ary)
	at := Evaluate(s, i.At)
	return ary.Index(at), false
}

func (i IndexFromTo) Evaluate(s stack.Stack) (value.Type, bool) {
	ary := Evaluate(s, i.Ary)
	from := Evaluate(s, i.From)
	to := Evaluate(s, i.To)
	return ary.Index(from, to), false
}

func (f Function) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.Function(f)
	if len(s) == 1 {
		return value.Function{Node: &n}, false
	} else {
		return value.Function{Node: &n, Frame: s.Top()}, false
	}
}

func (n Name) Evaluate(s stack.Stack) (value.Type, bool) {
	nd := node.Name(n)
	switch nd.Token() {

	case "true":
		return value.Bool(true), false

	case "false":
		return value.Bool(false), false

	default:
		v, _ := s.LookUp(nd.Token())
		return v, false
	}
}

func (i If) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.If(i)
	c := Evaluate(s, n.Condition)
	if cc, ok := c.(value.Bool); ok {
		if cc {
			return wrap(n.TrueCase).Evaluate(s)
		} else {
			return value.NoResultError, false
		}
	} else {
		return value.TypeError, false
	}
}

func (i IfElse) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.IfElse(i)
	c := Evaluate(s, n.Condition)
	if cc, ok := c.(value.Bool); ok {
		if cc {
			return wrap(n.TrueCase).Evaluate(s)
		} else {
			return wrap(n.FalseCase).Evaluate(s)
		}
	} else {
		return value.TypeError, false
	}
}

func (w While) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.While(w)
	r := value.Type(value.NoResultError)
	returning := false
	for {
		cond := Evaluate(s, n.Condition)
		if ccond, ok := cond.(value.Bool); ok {
			if !bool(ccond) || returning {
				return r, returning
			}
			r, returning = wrap(n.Body).Evaluate(s)
		} else {
			return value.TypeError, false
		}
	}
}

func (r Return) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.Return(r)
	return Evaluate(s, n.Target), true
}

func (r Read) Evaluate(s stack.Stack) (value.Type, bool) {
	b := bufio.NewReader(os.Stdin)
	line, err := b.ReadString('\n')
	if err != nil {
		return value.Error(fmt.Sprintf("read error %s", err)), false
	}
	return value.String(line), false
}

func (w Write) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.Write(w)
	v := Evaluate(s, n.Value)
	fmt.Println(v)
	return value.NoResultError, false
}

func (r Repl) Evaluate(s stack.Stack) (value.Type, bool) {
	rl := NewRLReader()
	defer rl.Close()
	Loop(rl, s, true)

	return value.NoResultError, false
}

func (a Aton) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.Aton(a)
	sv, ok := Evaluate(s, n.Value).(value.String)
	if !ok {
		return value.TypeError, false
	}

	if v, err := strconv.Atoi(string(sv)); err == nil {
		return value.Int(v), false
	}

	if v, err := strconv.ParseFloat(string(sv), 64); err == nil {
		return value.Float(v), false
	}

	return value.ConversionError, false
}

func (e Error) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.Error(e)
	v := Evaluate(s, n.Value)
	if s, ok := v.(value.String); ok {
		return value.Error(string(s)), false
	}
	return value.TypeError, false
}

func (b Block) Evaluate(s stack.Stack) (value.Type, bool) {
	n := node.Block(b)
	if len(n.Body) < 1 {
		panic("empty block")
	}
	r, returning := wrap(n.Body[0]).Evaluate(s)
	for i := 1; i < len(n.Body) && !returning; i++ {
		r, returning = wrap(n.Body[i]).Evaluate(s)
	}
	return r, returning
}

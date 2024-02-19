package node

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/paulsonkoly/calc/stack"
	"github.com/paulsonkoly/calc/types/value"
)

type Evaluator interface {
	Evaluate(s stack.Stack) (value.Type, bool)
}

// Evaluate evaluates the given AST node producing a Value
func Evaluate(s stack.Stack, e Evaluator) value.Type {
	r, _ := e.Evaluate(s)
	return r
}

func (i Int) Evaluate(s stack.Stack) (value.Type, bool) {
	x, err := strconv.Atoi(i.Token())
	if err != nil {
		panic(err)
	}
	return value.Int(x), false
}

func (f Float) Evaluate(s stack.Stack) (value.Type, bool) {
	x, err := strconv.ParseFloat(f.Token(), 64)
	if err != nil {
		panic(err)
	}
	return value.Float(x), false
}

func (s String) Evaluate(_ stack.Stack) (value.Type, bool) {
	tok := s.Token()
	// remove the first and last quotes, and replace escaped quotes with quotes
	tok = strings.ReplaceAll(tok, "\\\"", "\"")
	tok = tok[1 : len(tok)-1]
	return value.String(tok), false
}

func (a List) Evaluate(s stack.Stack) (value.Type, bool) {
	elems := a.Elems
	evalElems := make(value.Array, 0)

	for _, e := range elems {
		evalElems = append(evalElems, Evaluate(s, e))
	}

	return evalElems, false
}

func (c Call) Evaluate(s stack.Stack) (value.Type, bool) {
	fName := c.Name
	fl, ok := s.LookUp(fName)
  if ! ok {
    return fl, false
  }

  f, ok := fl.(value.Function)
  if !ok {
    return value.TypeError, false
  }

  fNode := f.Node.(*Function) // let panic if fails
  args := c.Arguments.Elems
  params := fNode.Parameters.Elems

  if len(args) != len(params) {
    return value.ArgumentError, false
  }

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
  r := Evaluate(s, fNode.Body)
  if f.Frame != nil {
    s.Pop()
  }
  s.Pop()
  return r, false
}

func (u UnOp) Evaluate(s stack.Stack) (value.Type, bool) {
	switch u.Token() {

	case "-":
		r := Evaluate(s, u.Target)
		r = r.Arith("*", value.Int(-1))
		return r, false

	case "#":
		r := Evaluate(s, u.Target)
		return r.Len(), false

	default:
		log.Panicf("unexpected unary op: %s\n", u.Token())
	}
	panic("unreachable code")
}

func (b BinOp) Evaluate(s stack.Stack) (value.Type, bool) {
	switch b.Token() {

	case "+", "-", "*", "/":
		return Evaluate(s, b.Left).Arith(b.Token(), Evaluate(s, b.Right)), false

	case "&", "|":
		return Evaluate(s, b.Left).Logic(b.Token(), Evaluate(s, b.Right)), false

	case "==":
		return Evaluate(s, b.Left).Eq(Evaluate(s, b.Right)), false

	case "!=":
		eq := Evaluate(s, b.Left).Eq(Evaluate(s, b.Right))
		if eq, ok := eq.(value.Bool); ok {
			return value.Bool(!eq), false
		}
		return eq, false

	case "<", "<=", ">", ">=":
		return Evaluate(s, b.Left).Relational(b.Token(), Evaluate(s, b.Right)), false

	case "=":
		v := Evaluate(s, b.Right)
		s.Set(b.Left.Token(), v)
		return v, false

	default:
		log.Panicf("unexpected single character in evaluator: %s", b.Token())
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
	if len(s) == 1 {
		return value.Function{Node: &f}, false
	} else {
		return value.Function{Node: &f, Frame: s.Top()}, false
	}
}

func (n Name) Evaluate(s stack.Stack) (value.Type, bool) {
	switch n.Token() {

	case "true":
		return value.Bool(true), false

	case "false":
		return value.Bool(false), false

	default:
		v, _ := s.LookUp(n.Token())
		return v, false
	}
}

func (i If) Evaluate(s stack.Stack) (value.Type, bool) {
	c := Evaluate(s, i.Condition)
	if cc, ok := c.(value.Bool); ok {
		if cc {
			return i.TrueCase.Evaluate(s)
		} else {
			return value.NoResultError, false
		}
	} else {
		return value.TypeError, false
	}
}

func (i IfElse) Evaluate(s stack.Stack) (value.Type, bool) {
	c := Evaluate(s, i.Condition)
	if cc, ok := c.(value.Bool); ok {
		if cc {
			return i.TrueCase.Evaluate(s)
		} else {
			return i.FalseCase.Evaluate(s)
		}
	} else {
		return value.TypeError, false
	}
}

func (w While) Evaluate(s stack.Stack) (value.Type, bool) {
	r := value.Type(value.NoResultError)
	returning := false
	for {
		cond := Evaluate(s, w.Condition)
		if ccond, ok := cond.(value.Bool); ok {
			if !bool(ccond) || returning {
				return r, returning
			}
			r, returning = w.Body.Evaluate(s)
		} else {
			return value.TypeError, false
		}
	}
}

func (r Return) Evaluate(s stack.Stack) (value.Type, bool) {
	return Evaluate(s, r.Target), true
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
	v := Evaluate(s, w.Value)
	fmt.Println(v)
	return value.NoResultError, false
}

func (r Repl) Evaluate(s stack.Stack) (value.Type, bool) {
	rl := NewRLReader()
	defer rl.Close()
	Loop(rl, r.Parser, s, true)

	return value.NoResultError, false
}

func (a Aton) Evaluate(s stack.Stack) (value.Type, bool) {
	sv, ok := Evaluate(s, a.Value).(value.String)
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
	v := Evaluate(s, e.Value)
	if s, ok := v.(value.String); ok {
		return value.Error(string(s)), false
	}
	return value.TypeError, false
}

func (b Block) Evaluate(s stack.Stack) (value.Type, bool) {
	if len(b.Body) < 1 {
		panic("empty block")
	}
	r, returning := b.Body[0].Evaluate(s)
	for i := 1; i < len(b.Body) && !returning; i++ {
		r, returning = b.Body[i].Evaluate(s)
	}
	return r, returning
}

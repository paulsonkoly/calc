package node

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/types/value"
)

type Evaluator interface {
	Evaluate(m *memory.Type) (value.Type, bool)
}

// Evaluate evaluates the given AST node producing a Value
func Evaluate(m *memory.Type, e Evaluator) value.Type {
	r, _ := e.Evaluate(m)
	return r
}

func (i Int) Evaluate(m *memory.Type) (value.Type, bool) {
	x, err := strconv.Atoi(i.Token())
	if err != nil {
		panic(err)
	}
	return value.Int(x), false
}

func (f Float) Evaluate(m *memory.Type) (value.Type, bool) {
	x, err := strconv.ParseFloat(f.Token(), 64)
	if err != nil {
		panic(err)
	}
	return value.Float(x), false
}

func (b Bool) Evaluate(_ *memory.Type) (value.Type, bool) { return value.Bool(b), false }

func (s String) Evaluate(_ *memory.Type) (value.Type, bool) {
	tok := s.Token()
	// remove the first and last quotes, and replace escaped quotes with quotes
	tok = strings.ReplaceAll(tok, "\\\"", "\"")
	tok = tok[1 : len(tok)-1]
	return value.String(tok), false
}

func (a List) Evaluate(m *memory.Type) (value.Type, bool) {
	elems := a.Elems
	evalElems := make(value.Array, 0)

	for _, e := range elems {
		evalElems = append(evalElems, Evaluate(m, e))
	}

	return evalElems, false
}

func (c Call) Evaluate(m *memory.Type) (value.Type, bool) {
	f := Evaluate(m, c.Name)

	fVal, ok := f.(value.Function)
	if !ok {
		return value.TypeError, false
	}

	fNode := fVal.Node.(*Function) // let panic if fails
	args := c.Arguments.Elems
	params := fNode.Parameters.Elems

	if len(args) != len(params) {
		return value.ArgumentError, false
	}

	// push 2 frames, one is the closure environment, the other is the frame for arguments
	// the arguments have to be evaluated before we push anything on the stack because what we push
	// ie the closure frame might contain variables that affect the argument evaluation
	frm := memory.NewFrame(fNode.LocalCnt)
	for i, a := range args {
		frm.Set(i, Evaluate(m, a))
	}
	if fVal.Frame != nil {
		m.PushFrame(fVal.Frame.(memory.Frame))
	}
	m.PushFrame(frm)
	r := Evaluate(m, fNode.Body)
	if fVal.Frame != nil {
		m.PopFrame()
	}
	m.PopFrame()
	return r, false
}

func (u UnOp) Evaluate(m *memory.Type) (value.Type, bool) {
	switch u.Token() {

	case "-":
		r := Evaluate(m, u.Target)
		r = r.Arith("*", value.Int(-1))
		return r, false

	case "#":
		r := Evaluate(m, u.Target)
		return r.Len(), false

	default:
		log.Panicf("unexpected unary op: %s\n", u.Token())
	}
	panic("unreachable code")
}

func (b BinOp) Evaluate(m *memory.Type) (value.Type, bool) {
	switch b.Token() {

	case "+", "-", "*", "/":
		return Evaluate(m, b.Left).Arith(b.Token(), Evaluate(m, b.Right)), false

	case "&", "|":
		return Evaluate(m, b.Left).Logic(b.Token(), Evaluate(m, b.Right)), false

	case "==", "!=":
		return Evaluate(m, b.Left).Eq(b.Token(), Evaluate(m, b.Right)), false

	case "<", "<=", ">", ">=":
		return Evaluate(m, b.Left).Relational(b.Token(), Evaluate(m, b.Right)), false

	default:
		log.Panicf("unexpected single character in evaluator: %s", b.Token())
	}
	panic("unreachable code")
}

func (a Assign) Evaluate(m *memory.Type) (value.Type, bool) {
	v := Evaluate(m, a.Value)

	switch vr := a.VarRef.(type) {
	case Local:
		m.Set(int(vr), v)
	case Name:
		m.SetGlobal(string(vr), v)
	default:
		panic("assignment lhs is neither local or global variable")
	}
	return v, false
}

func (i IndexAt) Evaluate(m *memory.Type) (value.Type, bool) {
	ary := Evaluate(m, i.Ary)
	at := Evaluate(m, i.At)
	return ary.Index(at), false
}

func (i IndexFromTo) Evaluate(m *memory.Type) (value.Type, bool) {
	ary := Evaluate(m, i.Ary)
	from := Evaluate(m, i.From)
	to := Evaluate(m, i.To)
	return ary.Index(from, to), false
}

func (f Function) Evaluate(m *memory.Type) (value.Type, bool) {
	return value.Function{Node: &f, Frame: m.Top()}, false
}

func (n Name) Evaluate(m *memory.Type) (value.Type, bool) {
	return m.LookUpGlobal(string(n)), false
}

func (l Local) Evaluate(m *memory.Type) (value.Type, bool) {
	return m.LookUpLocal(int(l)), false
}

func (c Closure) Evaluate(m *memory.Type) (value.Type, bool) {
	return m.LookUpClosure(int(c)), false
}

func (i If) Evaluate(m *memory.Type) (value.Type, bool) {
	c := Evaluate(m, i.Condition)
	if cc, ok := c.(value.Bool); ok {
		if cc {
			return i.TrueCase.Evaluate(m)
		} else {
			return value.NoResultError, false
		}
	} else {
		return value.TypeError, false
	}
}

func (i IfElse) Evaluate(m *memory.Type) (value.Type, bool) {
	c := Evaluate(m, i.Condition)
	if cc, ok := c.(value.Bool); ok {
		if cc {
			return i.TrueCase.Evaluate(m)
		} else {
			return i.FalseCase.Evaluate(m)
		}
	} else {
		return value.TypeError, false
	}
}

func (w While) Evaluate(m *memory.Type) (value.Type, bool) {
	r := value.Type(value.NoResultError)
	returning := false
	for {
		cond := Evaluate(m, w.Condition)
		if ccond, ok := cond.(value.Bool); ok {
			if !bool(ccond) || returning {
				return r, returning
			}
			r, returning = w.Body.Evaluate(m)
		} else {
			return value.TypeError, false
		}
	}
}

func (r Return) Evaluate(m *memory.Type) (value.Type, bool) {
	return Evaluate(m, r.Target), true
}

func (r Read) Evaluate(m *memory.Type) (value.Type, bool) {
	b := bufio.NewReader(os.Stdin)
	line, err := b.ReadString('\n')
	if err != nil {
		return value.Error(fmt.Sprintf("read error %s", err)), false
	}
	return value.String(line), false
}

func (w Write) Evaluate(m *memory.Type) (value.Type, bool) {
	v := Evaluate(m, w.Value)
	fmt.Println(v)
	return value.NoResultError, false
}

func (r Repl) Evaluate(m *memory.Type) (value.Type, bool) {
	rl := NewRLReader()
	defer rl.Close()
	Loop(rl, r.Parser, m, true, false)

	return value.NoResultError, false
}

func (a Aton) Evaluate(m *memory.Type) (value.Type, bool) {
	sv, ok := Evaluate(m, a.Value).(value.String)
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

func (t Toa) Evaluate(m *memory.Type) (value.Type, bool) {
	v := Evaluate(m, t.Value)
	return value.String(fmt.Sprint(v)), false
}

func (e Error) Evaluate(m *memory.Type) (value.Type, bool) {
	v := Evaluate(m, e.Value)
	if s, ok := v.(value.String); ok {
		return value.Error(string(s)), false
	}
	return value.TypeError, false
}

func (b Block) Evaluate(m *memory.Type) (value.Type, bool) {
	if len(b.Body) < 1 {
		panic("empty block")
	}
	r, returning := b.Body[0].Evaluate(m)
	for i := 1; i < len(b.Body) && !returning; i++ {
		r, returning = b.Body[i].Evaluate(m)
	}
	return r, returning
}

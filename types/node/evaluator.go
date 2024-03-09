package node

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/types/value"
)

type Evaluator interface {
	Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool)
}

// Evaluate evaluates the given AST node producing a Value
func Evaluate(m *memory.Type, e Evaluator, yield chan<- value.Type, control <-chan bool) value.Type {
	r, _ := e.Evaluate(m, yield, control)
	return r
}

func (i Int) Evaluate(_ *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return value.NewInt(int(i)), false
}
func (f Float) Evaluate(_ *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return value.NewFloat(float64(f)), false
}
func (b Bool) Evaluate(_ *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return value.NewBool(bool(b)), false
}
func (s String) Evaluate(_ *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return value.NewString(string(s)), false
}

func (a List) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	elems := a.Elems
	evalElems := make([]value.Type, len(elems))

	for i, e := range elems {
		evalElems[i] = Evaluate(m, e, yield, control)
	}

	return value.NewArray(evalElems), false
}

func (c Call) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	f := Evaluate(m, c.Name, yield, control)

	fVal, ok := f.ToFunction()
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
		frm.Set(i, Evaluate(m, a, yield, control))
	}
	if fVal.Frame != nil {
		m.PushFrame(fVal.Frame.(memory.Frame))
	}
	m.PushFrame(frm)
	r := Evaluate(m, fNode.Body, yield, control)
	if fVal.Frame != nil {
		m.PopFrame()
	}
	m.PopFrame()
	return r, false
}

func (u UnOp) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	switch u.Op {

	case "-":
		r := Evaluate(m, u.Target, yield, control)
		r = r.Arith("*", value.NewInt(-1))
		return r, false

	case "#":
		r := Evaluate(m, u.Target, yield, control)
		return r.Len(), false

	case "!":
		r := Evaluate(m, u.Target, yield, control)
		r = r.Not()
		return r, false

	default:
		log.Panicf("unexpected unary op: %s\n", u.Op)
	}
	panic("unreachable code")
}

func (b BinOp) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	switch b.Op {

	case "+", "-", "*", "/":
		return Evaluate(m, b.Left, yield, control).Arith(b.Op, Evaluate(m, b.Right, yield, control)), false

	case "%":
		return Evaluate(m, b.Left, yield, control).Mod(Evaluate(m, b.Right, yield, control)), false

	case "&", "|":
		return Evaluate(m, b.Left, yield, control).Logic(b.Op, Evaluate(m, b.Right, yield, control)), false

	case "==", "!=":
		return Evaluate(m, b.Left, yield, control).Eq(b.Op, Evaluate(m, b.Right, yield, control)), false

	case "<", "<=", ">", ">=":
		return Evaluate(m, b.Left, yield, control).Relational(b.Op, Evaluate(m, b.Right, yield, control)), false

	default:
		log.Panicf("unexpected single character in evaluator: %s", b.Op)
	}
	panic("unreachable code")
}

func (a Assign) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	v := Evaluate(m, a.Value, yield, control)

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

func (i IndexAt) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	ary := Evaluate(m, i.Ary, yield, control)
	at := Evaluate(m, i.At, yield, control)
	return ary.Index(at), false
}

func (i IndexFromTo) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	ary := Evaluate(m, i.Ary, yield, control)
	from := Evaluate(m, i.From, yield, control)
	to := Evaluate(m, i.To, yield, control)
	return ary.Index(from, to), false
}

func (f Function) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return value.NewFunction(&f, m.Top()), false
}

func (n Name) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return m.LookUpGlobal(string(n)), false
}

func (l Local) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return m.LookUpLocal(int(l)), false
}

func (c Closure) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return m.LookUpClosure(int(c)), false
}

func (i If) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	c := Evaluate(m, i.Condition, yield, control)
	if cc, ok := c.ToBool(); ok {
		if cc {
			return i.TrueCase.Evaluate(m, yield, control)
		} else {
			return value.NoResultError, false
		}
	} else {
		return value.TypeError, false
	}
}

func (i IfElse) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	c := Evaluate(m, i.Condition, yield, control)
	if cc, ok := c.ToBool(); ok {
		if cc {
			return i.TrueCase.Evaluate(m, yield, control)
		} else {
			return i.FalseCase.Evaluate(m, yield, control)
		}
	} else {
		return value.TypeError, false
	}
}

func (w While) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	r := value.Type(value.NoResultError)
	returning := false
	for {
		if returning {
			return r, returning
		}
		cond := Evaluate(m, w.Condition, yield, control)
		if ccond, ok := cond.ToBool(); ok {
			if !bool(ccond) {
				return r, returning
			}
			r, returning = w.Body.Evaluate(m, yield, control)
		} else {
			return value.TypeError, false
		}
	}
}

func (f For) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	myYield := make(chan value.Type)
	myControl := make(chan bool)

	// start a go routine as a co-routine to kick off the iterator
	go func() {
    // the iterator can't modify our stack, if it was pushing frames, our view
    // of the stack would be corrupted
    myM := m.Clone()
		Evaluate(myM, f.Iterator, myYield, myControl)
		close(myYield)
	}()

	r := value.NoResultError
	returning := false
	for {
		v, ok := <-myYield

		if !ok {
      close(myControl)
			return r, returning
		}

		switch vr := f.VarRef.(type) {
		case Local:
			m.Set(int(vr), v)
    case Name:
      m.SetGlobal(string(vr), v)
    default:
      panic("assignment lhs is neither local or global variable")
		}

    r, returning = f.Body.Evaluate(m, yield, control)

    // we need to stop and wait for the iterator
    // myControl <- !returning
	}
}

func (y Yield) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	v := Evaluate(m, y.Target, yield, control)
	yield <- v
	// c := <-control
  c := true
	return value.NoResultError, !c
}

func (r Return) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	return Evaluate(m, r.Target, yield, control), true
}

func (b Block) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	if len(b.Body) < 1 {
		panic("empty block")
	}
	r, returning := b.Body[0].Evaluate(m, yield, control)
	for i := 1; i < len(b.Body) && !returning; i++ {
		r, returning = b.Body[i].Evaluate(m, yield, control)
	}
	return r, returning
}

func (r Read) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	b := bufio.NewReader(os.Stdin)
	line, err := b.ReadString('\n')
	if err != nil {
		msg := fmt.Sprintf("read error %s", err)
		return value.NewError(&msg), false
	}
	return value.NewString(line), false
}

func (w Write) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	v := Evaluate(m, w.Value, yield, control)
	fmt.Println(v)
	return value.NoResultError, false
}

func (a Aton) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	sv, ok := Evaluate(m, a.Value, yield, control).ToString()
	if !ok {
		return value.TypeError, false
	}

	if v, err := strconv.Atoi(string(sv)); err == nil {
		return value.NewInt(v), false
	}

	if v, err := strconv.ParseFloat(string(sv), 64); err == nil {
		return value.NewFloat(v), false
	}

	return value.ConversionError, false
}

func (t Toa) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	v := Evaluate(m, t.Value, yield, control)
	return value.NewString(fmt.Sprint(v)), false
}

func (e Error) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	v := Evaluate(m, e.Value, yield, control)
	if msg, ok := v.ToString(); ok {
		return value.NewError(&msg), false
	}
	return value.TypeError, false
}

func (e Exit) Evaluate(m *memory.Type, yield chan<- value.Type, control <-chan bool) (value.Type, bool) {
	v := Evaluate(m, e.Value, yield, control)
	if i, ok := v.ToInt(); ok {
		os.Exit(i)
	}
	return value.TypeError, false
}

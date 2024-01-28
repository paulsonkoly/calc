package value

import "github.com/phaul/calc/types/node"

// type Type represents the evaluation result value
type Type interface {
	// Arith is the operator token string, other is the RHS of the operation,
	// possibly different type in which case type coercion happens
	Arith(op string, other Type) Type
	Relational(op string, other Type) Type
	Logic(op string, other Type) Type
}

type Int int
type Float float64
type Error string
type Bool bool
type Function struct {
	Node  node.Type
	Frame *Frame
}

// errors
var ZeroDivError = Error("division by zero")
var TypeError = Error("type error")
var InvalidOpError = Error("invalid operator")
var NoResultError = Error("no result")
var ArgumentError = Error("argument error")

func (i Int) Arith(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		if op == "/" && int(o) == 0 {
			return ZeroDivError
		}
		return Int(builtinArith[int](op, int(i), int(o)))

	case Float:
		return Float(builtinArith[float64](op, float64(i), float64(o)))

	case Bool, Function:
		return TypeError

	case Error:
		return o

	}
	panic("no type conversion")
}

func (i Int) Relational(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		return Bool(builtinRelational[int](op, int(i), int(o)))

	case Float:
		return Bool(builtinRelational[float64](op, float64(i), float64(o)))

	case Bool, Function:
		return TypeError

	case Error:
		return o

	}
	panic("no type conversion")
}

func (i Int) Logic(_ string, _ Type) Type { return TypeError }

func (f Float) Arith(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		return Float(builtinArith[float64](op, float64(f), float64(o)))

	case Float:
		return Float(builtinArith[float64](op, float64(f), float64(o)))

	case Bool, Function:
		return TypeError

	case Error:
		return o
	}
	panic("no type conversion")
}

func (f Float) Relational(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		return Bool(builtinRelational[float64](op, float64(f), float64(o)))

	case Float:
		return Bool(builtinRelational[float64](op, float64(f), float64(o)))

	case Bool, Function:
		return TypeError

	case Error:
		return o

	}
	panic("no type conversion")
}

func (f Float) Logic(_ string, _ Type) Type { return TypeError }

func (b Bool) Arith(_ string, _ Type) Type { return TypeError }

func (b Bool) Relational(op string, other Type) Type {
	switch o := other.(type) {

	case Int, Float, Function:
		return TypeError

	case Bool:
		switch op {
		case "==":
			return Bool(b == o)
		case "!=":
			return Bool(b != o)
		default:
			return InvalidOpError
		}

	case Error:
		return o

	}
	panic("no type conversion")
}

func (b Bool) Logic(op string, other Type) Type {
	switch o := other.(type) {

	case Int, Float, Function:
		return TypeError

	case Bool:
		switch op {
		case "&":
			return Bool(bool(b) && bool(o))
		case "|":
			return Bool(bool(b) || bool(o))
		default:
			return InvalidOpError
		}

	case Error:
		return o

	}
	panic("no type conversion")
}

func (e Error) Arith(_ string, _ Type) Type      { return e }
func (e Error) Relational(_ string, _ Type) Type { return e }
func (e Error) Logic(_ string, _ Type) Type      { return e }

func (f Function) Arith(_ string, _ Type) Type      { return TypeError }
func (f Function) Relational(_ string, _ Type) Type { return TypeError }
func (f Function) Logic(_ string, _ Type) Type      { return TypeError }
func (f Function) String() string                   { return "function" }

func builtinArith[t int | float64](op string, a, b t) t {
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		return a / b

	}
	panic("unknown operator")
}

func builtinRelational[t int | float64](op string, a, b t) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case "<":
		return a < b
	case ">":
		return a > b
	case "<=":
		return a <= b
	case ">=":
		return a >= b
	}
	panic("unknown operator")
}

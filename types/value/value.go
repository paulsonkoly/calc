package value

import (
	"fmt"
	"slices"

	"github.com/paulsonkoly/calc/types/node"
)

// Type represents the evaluation result value
//
// other is the RHS of the operation in which case type coersions happen
type Type interface {
	Arith(op string, other Type) Type
	Relational(op string, other Type) Type
	Logic(op string, other Type) Type
	Index(index ...Type) Type
	Eq(other Type) Type
	Len() Type
}

type Int int
type Float float64
type String string
type Array []Type
type Error string
type Bool bool
type Function struct {
	Node  *node.Function
	Frame *Frame
}

// errors
var ZeroDivError = Error("division by zero")
var TypeError = Error("type error")
var InvalidOpError = Error("invalid operator")
var NoResultError = Error("no result")
var ArgumentError = Error("argument error")
var IndexError = Error("index error")
var ConversionError = Error("conversion error")

func (i Int) Arith(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		if op == "/" && int(o) == 0 {
			return ZeroDivError
		}
		return Int(builtinArith[int](op, int(i), int(o)))

	case Float:
		return Float(builtinArith[float64](op, float64(i), float64(o)))

	case Bool, Function, String, Array:
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

	case Bool, Function, String, Array:
		return TypeError

	case Error:
		return o

	}
	panic("no type conversion")
}

func (i Int) Logic(_ string, _ Type) Type { return TypeError }
func (i Int) Index(_ ...Type) Type        { return TypeError }
func (i Int) Len() Type                   { return TypeError }

func (i Int) Eq(other Type) Type {
  switch other := other.(type) {
  case Int:
		return Bool(i == other)
  case Float:
		return Bool(i == Int(other))
  default:
	return TypeError
  }
}

func (f Float) Arith(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		return Float(builtinArith[float64](op, float64(f), float64(o)))

	case Float:
		return Float(builtinArith[float64](op, float64(f), float64(o)))

	case Bool, Function, String, Array:
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

	case Bool, Function, String, Array:
		return TypeError

	case Error:
		return o

	}
	panic("no type conversion")
}

func (f Float) Logic(_ string, _ Type) Type { return TypeError }

func (f Float) Index(_ ...Type) Type { return TypeError }

func (f Float) Len() Type { return TypeError }

func (f Float) Eq(other Type) Type {
  switch other := other.(type) {
  case Int:
		return Bool(f == Float(other))
  case Float:
		return Bool(f == other)
  default:
	return TypeError
  }
}

func (s String) Arith(op string, other Type) Type {
	switch other := other.(type) {
	case String:
		if op == "+" {
			return String(string(s) + string(other))
		} else {
			return InvalidOpError
		}

	case Error:
		return other

	default:
		return TypeError
	}
}

func (s String) Relational(op string, other Type) Type {
	switch other := other.(type) {
	case String:
		switch op {
		case "==":
			return Bool(s == other)

		case "!=":
			return Bool(s != other)

		default:
			return InvalidOpError
		}
	case Error:
		return other

	default:
		return TypeError
	}
}

func (s String) Logic(_ string, _ Type) Type { return TypeError }

func (s String) Len() Type { return Int(len(s)) }

func (s String) Index(index ...Type) Type {
	iix := []int{}

	for _, t := range index {
		if conv, ok := t.(Int); ok {
			iix = append(iix, int(conv))
			continue
		}
		return TypeError
	}

	switch len(iix) {
	case 2:
		if iix[0] < 0 || iix[0] >= len(s) {
			return IndexError
		}

		if iix[1] < iix[0] || iix[1] > len(s) {
			return IndexError
		}

		return String(string(s)[iix[0]:iix[1]])

	case 1:
		if iix[0] < 0 || iix[0] >= len(s) {
			return IndexError
		}

		return String(string(s)[iix[0]])

	default:
		panic("evaluator called index incorrectly")
	}
}

func (s String) Eq(other Type) Type {
	if other, ok := other.(String); ok {
		return Bool(s == other)
	}
	return TypeError
}

func (s String) String() string {
	return fmt.Sprintf("\"%s\"", string(s))
}

func (a Array) Arith(op string, other Type) Type {
	switch other := other.(type) {
	case Array:
		if op == "+" {
			return Array(append([]Type(a), []Type(other)...))
		} else {
			return InvalidOpError
		}

	case Error:
		return other

	default:
		return TypeError
	}
}

func (a Array) Relational(op string, other Type) Type {
	switch other := other.(type) {
	case Array:
		switch op {
		case "==":
			return Bool(slices.Equal(a, other))

		case "!=":
			return Bool(!slices.Equal(a, other))

		default:
			return InvalidOpError
		}
	case Error:
		return other

	default:
		return TypeError
	}
}

func (a Array) Logic(_ string, _ Type) Type { return TypeError }

func (a Array) Len() Type { return Int(len(a)) }

func (a Array) Index(index ...Type) Type {
	iix := []int{}

	for _, t := range index {
		if conv, ok := t.(Int); ok {
			iix = append(iix, int(conv))
			continue
		}
		return TypeError
	}

	switch len(iix) {
	case 2:
		if iix[0] < 0 || iix[0] >= len(a) {
			return IndexError
		}

		if iix[1] < iix[0] || iix[1] > len(a) {
			return IndexError
		}

		return Array(a[iix[0]:iix[1]])

	case 1:
		if iix[0] < 0 || iix[0] >= len(a) {
			return IndexError
		}

		return a[iix[0]]

	default:
		panic("evaluator called index incorrectly")
	}
}

func (a Array) Eq(other Type) Type {
	if other, ok := other.(Array); ok {
		if len(a) != len(other) {
			return Bool(false)
		}
		for i, v := range a {
			r := v.Eq(other[i])
			switch r {
			case Bool(false):
				return r
			case Bool(true):
				continue

			default:
				return r
			}
		}
		return Bool(true)
	}
	return TypeError
}

func (a Array) String() string {
	r := ""
	if len(a) > 0 {
		r += fmt.Sprintf("%v", a[0])

		for _, v := range a[1:] {
			r += ", " + fmt.Sprintf("%v", v)
		}
	}
	return "[" + r + "]"
}

func (b Bool) Arith(_ string, _ Type) Type { return TypeError }

func (b Bool) Relational(op string, other Type) Type {
	switch o := other.(type) {

	case Int, Float, Function, String, Array:
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

	case Int, Float, Function, String, Array:
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

func (b Bool) Index(_ ...Type) Type { return TypeError }
func (b Bool) Len() Type            { return TypeError }

func (b Bool) Eq(other Type) Type {
	if other, ok := other.(Bool); ok {
		return Bool(b == other)
	}
	return TypeError
}

func (e Error) Arith(_ string, _ Type) Type      { return e }
func (e Error) Relational(_ string, _ Type) Type { return e }
func (e Error) Logic(_ string, _ Type) Type      { return e }
func (e Error) Len() Type                        { return TypeError }
func (e Error) Index(_ ...Type) Type             { return e }
func (e Error) Eq(_ Type) Type                   { return e }

func (f Function) Arith(_ string, _ Type) Type      { return TypeError }
func (f Function) Relational(_ string, _ Type) Type { return TypeError }
func (f Function) Logic(_ string, _ Type) Type      { return TypeError }
func (f Function) Len() Type                        { return TypeError }
func (f Function) Index(_ ...Type) Type             { return TypeError }
func (f Function) Eq(_ Type) Type                   { return TypeError }
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

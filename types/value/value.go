package value

import "fmt"

// Type represents the evaluation result value
//
// other is the RHS of the operation in which case type coersions happen
type Type interface {
	Arith(op string, other Type) Type
	Relational(op string, other Type) Type
	Logic(op string, other Type) Type
	Index(index ...Type) Type
	Eq(op string, other Type) Type
	Len() Type
}

type Int int
type Float float64
type String string
type Array []Type
type Error struct{ Message *string }
type Bool bool
type Function struct {
	Node  any // Node is function AST
	Frame any // Frame is closure frame pointer
}

// errors
var (
	zeroDivStr    = "division by zero"
	typeStr       = "type error"
	invalidOpStr  = "invalid operator"
	noResultStr   = "no result"
	argumentStr   = "argument error"
	indexStr      = "index error"
	conversionStr = "conversion error"

	ZeroDivError    = Error{Message: &zeroDivStr}
	TypeError       = Error{Message: &typeStr}
	InvalidOpError  = Error{Message: &invalidOpStr}
	NoResultError   = Error{Message: &noResultStr}
	ArgumentError   = Error{Message: &argumentStr}
	IndexError      = Error{Message: &indexStr}
	ConversionError = Error{Message: &conversionStr}
)

func (i Int) Arith(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		if op == "/" && int(o) == 0 {
			return ZeroDivError
		}
		return builtinArith(op, i, o)

	case Float:
		return builtinArith(op, Float(i), o)

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
		return builtinRelational(op, i, o)

	case Float:
		return builtinRelational(op, Float(i), o)

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

func (i Int) Eq(op string, other Type) Type {
	switch other := other.(type) {
	case Int:
		return builtinEq(op, i, other)
	case Float:
		return builtinEq(op, Float(i), other)
	default:
		return TypeError
	}
}

func (f Float) Arith(op string, other Type) Type {
	switch o := other.(type) {

	case Int:
		return builtinArith(op, f, Float(o))

	case Float:
		return builtinArith(op, f, o)

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
		return builtinRelational(op, f, Float(o))

	case Float:
		return builtinRelational(op, f, o)

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

func (f Float) Eq(op string, other Type) Type {
	switch other := other.(type) {
	case Int:
		return builtinEq(op, f, Float(other))
	case Float:
		return builtinEq(op, f, other)
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
		return InvalidOpError

	case Error:
		return other

	default:
		return TypeError
	}
}

func (s String) Logic(_ string, _ Type) Type { return TypeError }

func (s String) Len() Type { return Int(len(s)) }

func (s String) Index(index ...Type) Type {
	iix := [2]int{}

	for i, t := range index {
		if conv, ok := t.(Int); ok {
			iix[i] = int(conv)
			continue
		}
		return TypeError
	}

	switch len(index) {
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
	}
	panic("unreachable code")
}

func (s String) Eq(op string, other Type) Type {
	if other, ok := other.(String); ok {
		return builtinEq(op, s, other)
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
		return InvalidOpError

	case Error:
		return other

	default:
		return TypeError
	}
}

func (a Array) Logic(_ string, _ Type) Type { return TypeError }

func (a Array) Len() Type { return Int(len(a)) }

func (a Array) Index(index ...Type) Type {
	iix := [2]int{}

	for i, t := range index {
		if conv, ok := t.(Int); ok {
			iix[i] = int(conv)
			continue
		}
		return TypeError
	}

	switch len(index) {
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
	}
	panic("unreachable code")
}

func (a Array) Eq(op string, other Type) Type {
	aother, ok := other.(Array)
	if !ok {
		return TypeError
	}

	found := true

	if op == "==" {
		found = false
		if len(a) != len(aother) {
			return Bool(false)
		}
	}

	for i, v := range a {
		r := v.Eq("==", aother[i])
		switch r {
		case Bool(false):
			return Bool(found)
		case Bool(true):
			continue

		default:
			return r
		}
	}
	return Bool(!found)
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
		return InvalidOpError

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

func (b Bool) Eq(op string, other Type) Type {
	if other, ok := other.(Bool); ok {
		return builtinEq(op, b, other)
	}
	return TypeError
}

func (e Error) Arith(_ string, _ Type) Type      { return e }
func (e Error) Relational(_ string, _ Type) Type { return e }
func (e Error) Logic(_ string, _ Type) Type      { return e }
func (e Error) Len() Type                        { return TypeError }
func (e Error) Index(_ ...Type) Type             { return e }
func (e Error) Eq(_ string, _ Type) Type         { return e }
func (e Error) String() string                   { return *e.Message }

func (f Function) Arith(_ string, _ Type) Type      { return TypeError }
func (f Function) Relational(_ string, _ Type) Type { return TypeError }
func (f Function) Logic(_ string, _ Type) Type      { return TypeError }
func (f Function) Len() Type                        { return TypeError }
func (f Function) Index(_ ...Type) Type             { return TypeError }
func (f Function) Eq(_ string, _ Type) Type         { return TypeError }
func (f Function) String() string                   { return "function" }

func builtinArith[t ~int | ~float64](op string, a, b t) t {
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

func builtinRelational[t ~int | ~float64](op string, a, b t) Bool {
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

func builtinEq[t ~int | ~float64 | ~bool | ~string](op string, a, b t) Bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	}
	panic("unknown operator")
}

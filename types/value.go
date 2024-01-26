package types

// type Value represents the evaluation result value
type Value interface {
	// Arith is the operator token string, other is the RHS of the operation,
	// possibly different type in which case type coercion happens
	Arith(op string, other Value) Value
	Relational(op string, other Value) Value
	Logic(op string, other Value) Value
}

type ValueInt int
type ValueFloat float64
type ValueError string
type ValueBool bool
type ValueFunction Node

// errors
var ZeroDivError = ValueError("division by zero")
var TypeError = ValueError("type error")
var InvalidOpError = ValueError("invalid operator")
var NoResultError = ValueError("no result")
var ArgumentError = ValueError("argument error")

func (i ValueInt) Arith(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		if op == "/" && int(o) == 0 {
			return ZeroDivError
		}
		return ValueInt(builtinArith[int](op, int(i), int(o)))

	case ValueFloat:
		return ValueFloat(builtinArith[float64](op, float64(i), float64(o)))

	case ValueBool, ValueFunction:
		return TypeError

	case ValueError:
		return o

	}
	panic("no type conversion")
}

func (i ValueInt) Relational(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		return ValueBool(builtinRelational[int](op, int(i), int(o)))

	case ValueFloat:
		return ValueBool(builtinRelational[float64](op, float64(i), float64(o)))

	case ValueBool, ValueFunction:
		return TypeError

	case ValueError:
		return o

	}
	panic("no type conversion")
}

func (i ValueInt) Logic(_ string, _ Value) Value { return TypeError }

func (f ValueFloat) Arith(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		return ValueFloat(builtinArith[float64](op, float64(f), float64(o)))

	case ValueFloat:
		return ValueFloat(builtinArith[float64](op, float64(f), float64(o)))

	case ValueBool, ValueFunction:
		return TypeError

	case ValueError:
		return o
	}
	panic("no type conversion")
}

func (f ValueFloat) Relational(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		return ValueBool(builtinRelational[float64](op, float64(f), float64(o)))

	case ValueFloat:
		return ValueBool(builtinRelational[float64](op, float64(f), float64(o)))

	case ValueBool, ValueFunction:
		return TypeError

	case ValueError:
		return o

	}
	panic("no type conversion")
}

func (f ValueFloat) Logic(_ string, _ Value) Value { return TypeError }

func (b ValueBool) Arith(_ string, _ Value) Value { return TypeError }

func (b ValueBool) Relational(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt, ValueFloat, ValueFunction:
		return TypeError

	case ValueBool:
		switch op {
		case "==":
			return ValueBool(b == o)
		case "!=":
			return ValueBool(b != o)
		default:
			return InvalidOpError
		}

	case ValueError:
		return o

	}
	panic("no type conversion")
}

func (b ValueBool) Logic(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt, ValueFloat, ValueFunction:
		return TypeError

	case ValueBool:
		switch op {
    case "&":
			return ValueBool(bool(b) && bool(o))
		case "|":
			return ValueBool(bool(b) || bool(o))
		default:
			return InvalidOpError
		}

	case ValueError:
		return o

	}
	panic("no type conversion")
}


func (e ValueError) Arith(_ string, _ Value) Value      { return e }
func (e ValueError) Relational(_ string, _ Value) Value { return e }
func (e ValueError) Logic(_ string, _ Value) Value      { return e }

func (f ValueFunction) Arith(_ string, _ Value) Value      { return TypeError }
func (f ValueFunction) Relational(_ string, _ Value) Value { return TypeError }
func (f ValueFunction) Logic(_ string, _ Value) Value      { return TypeError }
func (f ValueFunction) String() string                     { return "function" }

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

package evaluator

// type Value represents the evaluation result value
type Value interface {
	// Op is the operator token string, other is the RHS of the operation,
	// possibly different type in which case type coercion happens
	Op(op string, other Value) Value
}

type ValueInt int
type ValueFloat float64
type ValueError string

func (i ValueInt) Op(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		if op == "/" && int(o) == 0 {
			return ValueError("division by zero")
		}
		return ValueInt(doOp[int](op, int(i), int(o)))

	case ValueFloat:
		return ValueFloat(doOp[float64](op, float64(i), float64(o)))

	case ValueError:
		return o

	}
	panic("no type conversion")
}

func (f ValueFloat) Op(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		return ValueFloat(doOp[float64](op, float64(f), float64(o)))

	case ValueFloat:
		return ValueFloat(doOp[float64](op, float64(f), float64(o)))

	case ValueError:
		return o
	}
	panic("no type conversion")
}

func (e ValueError) Op(op string, other Value) Value { return e }

func doOp[t int | float64](op string, a, b t) t {
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
	panic("unkown operator")
}

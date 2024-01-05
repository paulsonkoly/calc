package evaluator

// type Value represents the evaluation result value
type Value interface {
	// Op is the operator token string, other is the RHS of the operation,
	// possibly different type in which case type coersion happens
	Op(op string, other Value) Value
}

type ValueInt int
type ValueFloat float64

func (i ValueInt) Op(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		return ValueInt(doOp[int](op, int(i), int(o)))

	case ValueFloat:
		return ValueFloat(doOp[float64](op, float64(i), float64(o)))

	}
	panic("no type conversion")
}

func (f ValueFloat) Op(op string, other Value) Value {
	switch o := other.(type) {

	case ValueInt:
		return ValueFloat(doOp[float64](op, float64(f), float64(o)))

	case ValueFloat:
		return ValueFloat(doOp[float64](op, float64(f), float64(o)))

	}
	panic("no type conversion")
}

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

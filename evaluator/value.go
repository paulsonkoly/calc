package evaluator

// The calc value type, supporting integers and floats with very simple
// coersion rules: float combined with anything results in a float, int
// combined with int is int
type Value interface {
	Plus(other Value) Value
	Minus(other Value) Value
	Mul(other Value) Value
	Div(other Value) Value
}

type ValueInt int
type ValueFloat float64

func (i ValueInt) Plus(other Value) Value {
	switch other.(type) {
	case ValueInt:
		{
			return ValueInt(int(i) + int(other.(ValueInt)))
		}
	case ValueFloat:
		{
			return ValueFloat(float64(i) + float64(other.(ValueFloat)))
		}
	}
	panic("no type conversion")
}

func (i ValueInt) Minus(other Value) Value {
	switch other.(type) {
	case ValueInt:
		{
			return ValueInt(int(i) - int(other.(ValueInt)))
		}
	case ValueFloat:
		{
			return ValueFloat(float64(i) - float64(other.(ValueFloat)))
		}
	}
	panic("no type conversion")
}

func (i ValueInt) Mul(other Value) Value {
	switch other.(type) {
	case ValueInt:
		{
			return ValueInt(int(i) * int(other.(ValueInt)))
		}
	case ValueFloat:
		{
			return ValueFloat(float64(i) * float64(other.(ValueFloat)))
		}
	}
	panic("no type conversion")
}

func (i ValueInt) Div(other Value) Value {
	switch other.(type) {
	case ValueInt:
		{
			return ValueInt(int(i) / int(other.(ValueInt)))
		}
	case ValueFloat:
		{
			return ValueFloat(float64(i) / float64(other.(ValueFloat)))
		}
	}
	panic("no type conversion")
}

func (i ValueFloat) Plus(other Value) Value {
	return ValueFloat(float64(i) + float64(other.(ValueFloat)))
}

func (i ValueFloat) Minus(other Value) Value {
	return ValueFloat(float64(i) - float64(other.(ValueFloat)))
}

func (i ValueFloat) Mul(other Value) Value {
	return ValueFloat(float64(i) * float64(other.(ValueFloat)))
}

func (i ValueFloat) Div(other Value) Value {
	return ValueFloat(float64(i) / float64(other.(ValueFloat)))
}

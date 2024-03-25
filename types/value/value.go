// value defines the type that represents the evaluation result.
package value

import (
	"fmt"
	"slices"
	"strconv"
	"unsafe"
)

type kind int

const (
	invalidT = kind(iota)
	intT
	floatT
	stringT
	arrayT
	errorT
	boolT
	functionT
)

// Type is evaluation result value.
//
// It is a uniform structure, not an interface, to keep evaluation on the stack
// as much as possible.
type Type struct {
	typ   kind // Kind of the the value
	morph uint64
	ptr   unsafe.Pointer
}

// A structure that represents a function value.
type FunctionData struct {
	Node     int // Pointer to the code of the function - the AST node that holds the function
	Frame    any // Pointer to the closure stack frame
	ParamCnt int // ParamCnt is the number of parameters of the function
	LocalCnt int // LocalCnt is the number of local variables of the function including ParamCnt
}

// unsafe (no type check) accessors.
func (t Type) i() int     { return *(*int)(unsafe.Pointer(&t.morph)) }
func (t Type) f() float64 { return *(*float64)(unsafe.Pointer(&t.morph)) }
func (t Type) b() bool    { return t.morph != 0 }
func (t Type) s() string  { return *(*string)(t.ptr) }
func (t Type) a() []Type  { return *(*[]Type)(t.ptr) }

// NewInt allocates a new int value.
func NewInt(i int) Type { return Type{typ: intT, morph: *(*uint64)(unsafe.Pointer(&i))} }

// NewFloat allocates a new float value.
func NewFloat(f float64) Type { return Type{typ: floatT, morph: *(*uint64)(unsafe.Pointer(&f))} }

// NewBool allocates a new bool value.
func NewBool(b bool) Type {
	t := Type{typ: boolT}

	if b {
		t.morph = 1
	} else {
		t.morph = 0
	}
	return t
}

// NewArray allocates a new array value.
func NewArray(a []Type) Type { return Type{typ: arrayT, ptr: unsafe.Pointer(&a)} }

// NewString allocates a new string value.
func NewString(s string) Type { return Type{typ: stringT, ptr: unsafe.Pointer(&s)} }

// NewError allocates a new error value.
func NewError(m *string) Type { return Type{typ: errorT, ptr: unsafe.Pointer(m)} }

// NewFunction allocates a new function value.
func NewFunction(node int, frame any, paramCnt int, localCnt int) Type {
	d := FunctionData{Node: node, Frame: frame, ParamCnt: paramCnt, LocalCnt: localCnt}
	return Type{typ: functionT, ptr: unsafe.Pointer(&d)}
}

// ToFunction converts a value to FunctionData.
//
// It returns ok false if not a function.
func (t Type) ToFunction() (FunctionData, bool) {
	if t.typ != functionT {
		return FunctionData{}, false
	}

	return *(*FunctionData)(unsafe.Pointer(t.ptr)), true
}

// ToInt converts a value to int.
//
// It returns ok false if not an int.
func (t Type) ToInt() (int, bool) {
	if t.typ != intT {
		return 0, false
	}
	return *(*int)(unsafe.Pointer(&t.morph)), true
}

// ToBool converts a value to bool.
//
// It returns ok false if not an bool.
func (t Type) ToBool() (bool, bool) {
	if t.typ != boolT {
		return false, false
	}
	return t.morph == 1, true
}

// ToString converts a value to string.
//
// It returns ok false if not a string.
func (t Type) ToString() (string, bool) {
	if t.typ != stringT {
		return "", false
	}
	return *(*string)(unsafe.Pointer(t.ptr)), true
}

func (t Type) ToArray() ([]Type, bool) {
	if t.typ != arrayT {
		return nil, false
	}
	return *(*[]Type)(unsafe.Pointer(t.ptr)), true
}

// String converts any value.Type to string.
func (t Type) String() string {
	switch t.typ {
	case intT:
		return strconv.Itoa(*(*int)(unsafe.Pointer(&t.morph)))
	case floatT:
		return fmt.Sprint(*(*float64)(unsafe.Pointer(&t.morph)))
	case boolT:
		return strconv.FormatBool(t.morph == 1)
	case stringT:
		return *(*string)(t.ptr)
	case errorT:
		return *(*string)(t.ptr)
	case functionT:
		return "function"
	case arrayT:
		a := *(*[]Type)(t.ptr)

		r := ""
		if len(a) > 0 {
			r += fmt.Sprintf("%v", a[0])

			for _, v := range a[1:] {
				r += ", " + fmt.Sprintf("%v", v)
			}
		}
		return "[" + r + "]"
	}
	panic("type not handled in String")
}

// Display converts a value to a string for calc result printing.
//
// Adds extra quotes around string type.
func (t Type) Display() string {
	if t.typ == stringT {
		return fmt.Sprintf("\"%v\"", *(*string)(t.ptr))
	}
	return t.String()
}

// Predefined errors.
var (
	typeErrorStr       = "type error"
	invalidOpErrorStr  = "invalid operation"
	zeroDivErrorStr    = "division by zero"
	indexErrorStr      = "index error"
	noResultErrorStr   = "no result"
	conversionErrorStr = "conversion error"
	argumentErrorStr   = "argument error"

	TypeError       = NewError(&typeErrorStr)
	InvalidOpError  = NewError(&invalidOpErrorStr)
	ZeroDivError    = NewError(&zeroDivErrorStr)
	IndexError      = NewError(&indexErrorStr)
	NoResultError   = NewError(&noResultErrorStr)
	ConversionError = NewError(&conversionErrorStr)
	ArgumentError   = NewError(&argumentErrorStr)
)

// Arith is value arithmetics, +, -, * /.
func (t Type) Arith(op string, b Type) Type {

	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | intT:
		aVal := t.i()
		bVal := b.i()

		if op == "/" && bVal == 0 {
			return ZeroDivError
		}

		return NewInt(builtinArith(op, aVal, bVal))

	case (intT << 4) | floatT:
		aVal := t.i()
		bVal := b.f()
		return NewFloat(builtinArith(op, float64(aVal), bVal))

	case (floatT << 4) | intT:
		aVal := t.f()
		bVal := b.i()
		return NewFloat(builtinArith(op, aVal, float64(bVal)))

	case (floatT << 4) | floatT:
		aVal := t.f()
		bVal := b.f()
		return NewFloat(builtinArith(op, aVal, bVal))

	case (intT << 4) | errorT, (floatT << 4) | errorT, (stringT << 4) | errorT,
		(arrayT << 4) | errorT, (boolT << 4) | errorT, (functionT << 4) | errorT:
		return b

	case (errorT << 4) | intT, (errorT << 4) | floatT, (errorT << 4) | stringT,
		(errorT << 4) | arrayT, (errorT << 4) | errorT, (errorT << 4) | boolT,
		(errorT << 4) | functionT:
		return t

	case (stringT << 4) | stringT:
		if op != "+" {
			return InvalidOpError
		}

		aVal := t.s()
		bVal := b.s()
		return NewString(aVal + bVal)

	case (arrayT << 4) | arrayT:
		if op != "+" {
			return InvalidOpError
		}

		aVal := t.a()
		bVal := b.a()
		return NewArray(append(slices.Clone(aVal), bVal...))

	default:
		if t.typ == b.typ {
			return InvalidOpError
		} else {
			return TypeError
		}
	}
}

func (t Type) Mod(b Type) Type {
	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | intT:
		aVal := t.i()
		bVal := b.i()

		return NewInt(aVal % bVal)

	case (intT << 4) | errorT, (floatT << 4) | errorT, (stringT << 4) | errorT,
		(arrayT << 4) | errorT, (boolT << 4) | errorT, (functionT << 4) | errorT:
		return b

	case (errorT << 4) | intT, (errorT << 4) | floatT, (errorT << 4) | stringT,
		(errorT << 4) | arrayT, (errorT << 4) | errorT, (errorT << 4) | boolT,
		(errorT << 4) | functionT:
		return t

	default:
		if t.typ == b.typ {
			return InvalidOpError
		} else {
			return TypeError
		}
	}
}

// Relational is value relational <, >, <= ...
func (t Type) Relational(op string, b Type) Type {
	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | intT:
		aVal := t.i()
		bVal := b.i()

		return NewBool(builtinRelational(op, aVal, bVal))

	case (intT << 4) | floatT:
		aVal := t.i()
		bVal := b.f()
		return NewBool(builtinRelational(op, float64(aVal), bVal))

	case (floatT << 4) | intT:
		aVal := t.f()
		bVal := b.i()
		return NewBool(builtinRelational(op, aVal, float64(bVal)))

	case (floatT << 4) | floatT:
		aVal := t.f()
		bVal := b.f()
		return NewBool(builtinRelational(op, aVal, bVal))

	case (intT << 4) | errorT, (floatT << 4) | errorT, (stringT << 4) | errorT,
		(arrayT << 4) | errorT, (boolT << 4) | errorT, (functionT << 4) | errorT:
		return b

	case (errorT << 4) | intT, (errorT << 4) | floatT, (errorT << 4) | stringT,
		(errorT << 4) | arrayT, (errorT << 4) | errorT, (errorT << 4) | boolT,
		(errorT << 4) | functionT:
		return t

	default:
		if t.typ == b.typ {
			return InvalidOpError
		} else {
			return TypeError
		}
	}
}

// Logic is value logic ops &, |.
func (t Type) Logic(op string, b Type) Type {

	switch (t.typ)<<4 | b.typ {
	case (intT << 4) | intT:
		aVal := t.morph
		bVal := b.morph

		if op == "&" {
			aVal &= bVal
		} else {
			aVal |= bVal
		}
		return NewInt(int(aVal))

	case (boolT << 4) | boolT:
		aVal := t.morph
		bVal := b.morph

		if op == "&" {
			aVal &= bVal
		} else {
			aVal |= bVal
		}
		return NewBool(aVal == 1)

	case (intT << 4) | errorT, (floatT << 4) | errorT, (stringT << 4) | errorT,
		(arrayT << 4) | errorT, (boolT << 4) | errorT, (functionT << 4) | errorT:
		return b

	case (errorT << 4) | intT, (errorT << 4) | floatT, (errorT << 4) | stringT,
		(errorT << 4) | arrayT, (errorT << 4) | errorT, (errorT << 4) | boolT,
		(errorT << 4) | functionT:
		return t

	default:
		if t.typ == b.typ {
			return InvalidOpError
		} else {
			return TypeError
		}
	}
}

// Shift is bit shift ops <<, >>.
func (t Type) Shift(op string, b Type) Type {

	switch (t.typ)<<4 | b.typ {
	case (intT << 4) | intT:
		aVal := t.morph
		bVal := b.morph

		if op == "<<" {
			aVal <<= bVal
		} else {
			aVal >>= bVal
		}
		return NewInt(int(aVal))

	case (intT << 4) | errorT, (floatT << 4) | errorT, (stringT << 4) | errorT,
		(arrayT << 4) | errorT, (boolT << 4) | errorT, (functionT << 4) | errorT:
		return b

	case (errorT << 4) | intT, (errorT << 4) | floatT, (errorT << 4) | stringT,
		(errorT << 4) | arrayT, (errorT << 4) | errorT, (errorT << 4) | boolT,
		(errorT << 4) | functionT:
		return t

	default:
		if t.typ == b.typ {
			return InvalidOpError
		} else {
			return TypeError
		}
	}
}

// Flip is integer bit flip operator.
func (t Type) Flip() Type {
	switch t.typ {
	case intT:
		return NewInt(int(^t.morph))
	case errorT:
		return t
	default:
		return TypeError
	}
}

// Not is boolean not operator.
func (t Type) Not() Type {
	switch t.typ {
	case boolT:
		return NewBool(t.morph != 1)
	case errorT:
		return t
	default:
		return TypeError
	}
}

// Index is value indexing, [] and [:].
func (t Type) Index(b ...Type) Type {

	if len(b) < 1 || len(b) > 2 {
		panic("Index incorrectly called")
	}

	iix := [2]int{}
	for i, t := range b {
		if t.typ == intT {
			iix[i] = t.i()
			continue
		}
		return TypeError
	}

	switch t.typ {
	case stringT:
		s := *(*string)(t.ptr)

		switch len(b) {
		case 2:
			if iix[0] < 0 || iix[0] > len(s) || iix[1] < iix[0] || iix[1] > len(s) {

				return IndexError
			}

			s = s[iix[0]:iix[1]]
			return NewString(s)
		case 1:

			if iix[0] < 0 || iix[0] >= len(s) {

				return IndexError
			}

			s = string(s[iix[0]])
			return NewString(s)
		}
	case arrayT:

		ary := *(*[]Type)(t.ptr)

		switch len(b) {
		case 2:
			if iix[0] < 0 || iix[0] > len(ary) || iix[1] < iix[0] || iix[1] > len(ary) {

				return IndexError
			}

			ary = ary[iix[0]:iix[1]]
			return NewArray(ary)
		case 1:

			if iix[0] < 0 || iix[0] >= len(ary) {

				return IndexError
			}

			return ary[iix[0]]
		}

	case (errorT << 4) | intT, (errorT << 4) | floatT, (errorT << 4) | stringT,
		(errorT << 4) | arrayT, (errorT << 4) | errorT, (errorT << 4) | boolT,
		(errorT << 4) | functionT:
		return t

	default:
		// There is no invalid op here, and indexing with error also throws away
		// the error unlike other operators. Incorrect indexing is always type
		// error except the case above, that keeps error
		return TypeError
	}
	panic("unreachable code")
}

// Len is value length.
func (t Type) Len() Type {

	switch t.typ {

	case stringT:
		s := *(*string)(t.ptr)
		i := len(s)
		return NewInt(i)

	case arrayT:
		s := *(*[]Type)(t.ptr)
		i := len(s)
		return NewInt(i)

	case errorT:
		return t

	default:
		return TypeError
	}
}

// StrictEq decides whether a and b are exactly the same value.
//
//	1 == 1.0 -> false
//
// in strict equality all functions are equal, this is counter intuitive, but
// for testing it makes sense.
func (t *Type) StrictEq(b Type) bool {
	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | intT:
		aVal := t.i()
		bVal := b.i()

		return aVal == bVal

	case (floatT << 4) | floatT:
		aVal := t.f()
		bVal := b.f()

		return aVal == bVal

	case (boolT << 4) | boolT:
		aVal := t.b()
		bVal := b.b()

		return aVal == bVal

	case (stringT << 4) | stringT:
		aVal := *(*string)(t.ptr)
		bVal := *(*string)(b.ptr)

		return aVal == bVal

	case (errorT << 4) | errorT:
		aVal := *(*string)(t.ptr)
		bVal := *(*string)(b.ptr)

		return aVal == bVal

	case (arrayT << 4) | arrayT:
		aVal := *(*[]Type)(t.ptr)
		bVal := *(*[]Type)(b.ptr)

		if len(aVal) != len(bVal) {
			return false
		}

		for i, t := range aVal {
			if !t.StrictEq(bVal[i]) {
				return false
			}
		}
		return true

	case (functionT << 4) | functionT:
		return true

	default:
		return false
	}
}

// WeakEq decides whether a and b are the same value as per language rules.
//
//	1 == 1.0 -> true
//
// All functions are un-equal.
func (t *Type) WeakEq(b Type) bool {

	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | floatT:
		aVal := t.i()
		bVal := b.f()

		return float64(aVal) == bVal

	case (floatT << 4) | intT:
		aVal := t.f()
		bVal := b.i()

		return aVal == float64(bVal)

	case (arrayT << 4) | arrayT:
		aVal := *(*[]Type)(t.ptr)
		bVal := *(*[]Type)(b.ptr)

		if len(aVal) != len(bVal) {
			return false
		}

		for i, t := range aVal {
			if !t.WeakEq(bVal[i]) {
				return false
			}
		}
		return true

	case (functionT << 4) | functionT:
		return false

	default:
		return t.StrictEq(b)
	}
}

// Equality check, ==, !=.
func (t Type) Eq(op string, b Type) Type {
	r := t.WeakEq(b)
	if op == "!=" {
		r = !r
	}
	return NewBool(r)
}

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

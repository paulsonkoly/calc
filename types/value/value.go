// Package value defines the type that represents the evaluation result.
package value

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"unsafe"

	"github.com/paulsonkoly/calc/types/bytecode"
)

type kind int

const (
	nilT = kind(iota)
	intT
	floatT
	stringT
	arrayT
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
	Node     int     // Pointer to the code of the function - the AST node that holds the function
	Frame    *[]Type // Pointer to the closure stack frame
	ParamCnt int     // ParamCnt is the number of parameters of the function
	LocalCnt int     // LocalCnt is the number of local variables of the function including ParamCnt
}

// unsafe (no type check) accessors.
func (t Type) i() int     { return *(*int)(unsafe.Pointer(&t.morph)) }
func (t Type) f() float64 { return *(*float64)(unsafe.Pointer(&t.morph)) }
func (t Type) b() bool    { return t.morph != 0 }
func (t Type) s() string  { return *(*string)(t.ptr) }
func (t Type) a() []Type  { return *(*[]Type)(t.ptr) }

// Nil is the nil value.
var Nil = Type{typ: nilT}

// IsNil determines whether a value is nil.
func (t Type) IsNil() bool { return t.typ == nilT }

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

// Function binary layout.
const (
	paramsCntHi = 63
	paramsCntLo = 48
	localCntHi  = 47
	localCntLo  = 32
	ipHi        = 31
	ipLo        = 0
)

// NewFunction allocates a new function value.
func NewFunction(node int, frame *[]Type, paramCnt int, localCnt int) Type {
	nd := ((uint64)(node)) & ((1 << (ipHi - ipLo + 1)) - 1)
	pc := ((uint64)(paramCnt)) & ((1 << (paramsCntHi - paramsCntLo + 1)) - 1)
	lc := ((uint64)(localCnt)) & ((1 << (localCntHi - localCntLo + 1)) - 1)
	morp := (pc << paramsCntLo) | (lc << localCntLo) | nd<<ipLo
	return Type{typ: functionT, morph: morp, ptr: unsafe.Pointer(frame)}
}

func (t *Type) SetFrame(frame *[]Type) {
	if t.typ != functionT {
		panic("type is not a function")
	}
	t.ptr = unsafe.Pointer(frame)
}

// ToFunction converts a value to FunctionData.
//
// It returns ok false if not a function.
func (t Type) ToFunction() (FunctionData, bool) {
	if t.typ != functionT {
		return FunctionData{}, false
	}

	nd := int((t.morph)>>ipLo) & ((1 << (ipHi - ipLo + 1)) - 1)
	pc := int((t.morph)>>paramsCntLo) & ((1 << (paramsCntHi - paramsCntLo + 1)) - 1)
	lc := int((t.morph)>>localCntLo) & ((1 << (localCntHi - localCntLo + 1)) - 1)

	return FunctionData{ParamCnt: pc, Frame: (*[]Type)(t.ptr), LocalCnt: lc, Node: nd}, true
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
	case nilT:
		return "nil"
	case intT:
		return strconv.Itoa(*(*int)(unsafe.Pointer(&t.morph)))
	case floatT:
		return fmt.Sprint(*(*float64)(unsafe.Pointer(&t.morph)))
	case boolT:
		return strconv.FormatBool(t.morph == 1)
	case stringT:
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

// Abbrev abbreviates string of t to 20 characters with ellipses.
func (t Type) Abbrev() string {
	s := t.String()
	if len(s) > 20 {
		return s[:17] + "..."
	}
	return s
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
	ErrNil     = errors.New("nil error")
	ErrType    = errors.New("type error")
	ErrZeroDiv = errors.New("division by zero")
	ErrIndex   = errors.New("index error")
)

// Arith is value arithmetics, +, -, * /.
func (t Type) Arith(op bytecode.OpCode, b Type) (Type, error) {

	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | intT:
		aVal := t.i()
		bVal := b.i()

		if op == bytecode.DIV && bVal == 0 {
			return Nil, ErrZeroDiv
		}

		return NewInt(builtinArith(op, aVal, bVal)), nil

	case (intT << 4) | floatT:
		aVal := t.i()
		bVal := b.f()
		return NewFloat(builtinArith(op, float64(aVal), bVal)), nil

	case (floatT << 4) | intT:
		aVal := t.f()
		bVal := b.i()
		return NewFloat(builtinArith(op, aVal, float64(bVal))), nil

	case (floatT << 4) | floatT:
		aVal := t.f()
		bVal := b.f()
		return NewFloat(builtinArith(op, aVal, bVal)), nil

	case (stringT << 4) | stringT:
		if op != bytecode.ADD {
			return Nil, ErrType
		}

		aVal := t.s()
		bVal := b.s()
		return NewString(aVal + bVal), nil

	case (arrayT << 4) | arrayT:
		if op != bytecode.ADD {
			return Nil, ErrType
		}

		aVal := t.a()
		bVal := b.a()
		return NewArray(append(slices.Clone(aVal), bVal...)), nil

	default:
		if t.typ == nilT || b.typ == nilT {
			return Nil, ErrNil
		} else {
			return Nil, ErrType
		}
	}
}

func (t Type) Mod(b Type) (Type, error) {
	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | intT:
		aVal := t.i()
		bVal := b.i()

		return NewInt(aVal % bVal), nil

	default:
		if t.typ == nilT || b.typ == nilT {
			return Nil, ErrNil
		} else {
			return Nil, ErrType
		}
	}
}

// Relational is value relational <, >, <= ...
func (t Type) Relational(op bytecode.OpCode, b Type) (Type, error) {
	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | intT:
		aVal := t.i()
		bVal := b.i()

		return NewBool(builtinRelational(op, aVal, bVal)), nil

	case (intT << 4) | floatT:
		aVal := t.i()
		bVal := b.f()
		return NewBool(builtinRelational(op, float64(aVal), bVal)), nil

	case (floatT << 4) | intT:
		aVal := t.f()
		bVal := b.i()
		return NewBool(builtinRelational(op, aVal, float64(bVal))), nil

	case (floatT << 4) | floatT:
		aVal := t.f()
		bVal := b.f()
		return NewBool(builtinRelational(op, aVal, bVal)), nil

	default:
		if t.typ == nilT || b.typ == nilT {
			return Nil, ErrNil
		} else {
			return Nil, ErrType
		}
	}
}

// Logic is value logic ops &, |.
func (t Type) Logic(op bytecode.OpCode, b Type) (Type, error) {

	switch (t.typ)<<4 | b.typ {
	case (intT << 4) | intT:
		aVal := t.morph
		bVal := b.morph

		if op == bytecode.AND {
			aVal &= bVal
		} else {
			aVal |= bVal
		}
		return NewInt(int(aVal)), nil

	case (boolT << 4) | boolT:
		aVal := t.morph
		bVal := b.morph

		if op == bytecode.AND {
			aVal &= bVal
		} else {
			aVal |= bVal
		}
		return NewBool(aVal == 1), nil

	default:
		if t.typ == nilT || b.typ == nilT {
			return Nil, ErrNil
		} else {
			return Nil, ErrType
		}
	}
}

// Shift is bit shift ops <<, >>.
func (t Type) Shift(op bytecode.OpCode, b Type) (Type, error) {

	switch (t.typ)<<4 | b.typ {
	case (intT << 4) | intT:
		aVal := t.morph
		bVal := b.morph

		if op == bytecode.LSH {
			aVal <<= bVal
		} else {
			aVal >>= bVal
		}
		return NewInt(int(aVal)), nil

	default:
		if t.typ == nilT || b.typ == nilT {
			return Nil, ErrNil
		} else {
			return Nil, ErrType
		}
	}
}

// Flip is integer bit flip operator.
func (t Type) Flip() (Type, error) {
	switch t.typ {
	case intT:
		return NewInt(int(^t.morph)), nil
	case nilT:
		return Nil, ErrNil
	default:
		return Nil, ErrType
	}
}

// Not is boolean not operator.
func (t Type) Not() (Type, error) {
	switch t.typ {
	case boolT:
		return NewBool(t.morph != 1), nil
	case nilT:
		return Nil, ErrNil
	default:
		return Nil, ErrType
	}
}

// Index is value indexing, [] and [:].
func (t Type) Index(b ...Type) (Type, error) {

	if len(b) < 1 || len(b) > 2 {
		panic("Index incorrectly called")
	}

	iix := [2]int{}
	for i, t := range b {
		switch t.typ {
		case intT:
			iix[i] = t.i()
		case nilT:
			return Nil, ErrNil
		default:
			return Nil, ErrType
		}
	}

	switch t.typ {
	case stringT:
		s := *(*string)(t.ptr)

		switch len(b) {
		case 2:
			if iix[0] < 0 || iix[0] > len(s) || iix[1] < iix[0] || iix[1] > len(s) {

				return Nil, ErrIndex
			}

			s = s[iix[0]:iix[1]]
			return NewString(s), nil
		case 1:

			if iix[0] < 0 || iix[0] >= len(s) {

				return Nil, ErrIndex
			}

			s = string(s[iix[0]])
			return NewString(s), nil
		}
	case arrayT:

		ary := *(*[]Type)(t.ptr)

		switch len(b) {
		case 2:
			if iix[0] < 0 || iix[0] > len(ary) || iix[1] < iix[0] || iix[1] > len(ary) {

				return Nil, ErrIndex
			}

			ary = ary[iix[0]:iix[1]]
			return NewArray(ary), nil
		case 1:

			if iix[0] < 0 || iix[0] >= len(ary) {

				return Nil, ErrIndex
			}

			return ary[iix[0]], nil
		}

	default:
		// There is no invalid op here, and indexing with error also throws away
		// the error unlike other operators. Incorrect indexing is always type
		// error except the case above, that keeps error
		return Nil, ErrType
	}
	panic("unreachable code")
}

// Len is value length.
func (t Type) Len() (Type, error) {

	switch t.typ {

	case stringT:
		s := *(*string)(t.ptr)
		i := len(s)
		return NewInt(i), nil

	case arrayT:
		s := *(*[]Type)(t.ptr)
		i := len(s)
		return NewInt(i), nil

	case nilT:
		return Nil, ErrNil

	default:
		return Nil, ErrType
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

	case (nilT << 4) | nilT, (functionT << 4) | functionT:
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
func (t *Type) WeakEq(b Type) (bool, error) {

	switch (t.typ)<<4 | b.typ {

	case (intT << 4) | floatT:
		aVal := t.i()
		bVal := b.f()

		return float64(aVal) == bVal, nil

	case (floatT << 4) | intT:
		aVal := t.f()
		bVal := b.i()

		return aVal == float64(bVal), nil

	case (arrayT << 4) | arrayT:
		aVal := *(*[]Type)(t.ptr)
		bVal := *(*[]Type)(b.ptr)

		if len(aVal) != len(bVal) {
			return false, nil
		}

		for i, t := range aVal {
			r, err := t.WeakEq(bVal[i])
			if !r {
				return false, err
			}
		}
		return true, nil

	case (functionT << 4) | functionT:
		return false, nil

	default:
		if t.typ == nilT || b.typ == nilT {
			return false, ErrNil
		}
		return t.StrictEq(b), nil
	}
}

// Equality check, ==, !=.
func (t Type) Eq(op bytecode.OpCode, b Type) (Type, error) {
	r, err := t.WeakEq(b)
	if err != nil {
		return Nil, err
	}
	if op == bytecode.NE {
		r = !r
	}
	return NewBool(r), nil
}

func builtinArith[t int | float64](op bytecode.OpCode, a, b t) t {
	switch op {
	case bytecode.ADD:
		return a + b
	case bytecode.SUB:
		return a - b
	case bytecode.MUL:
		return a * b
	case bytecode.DIV:
		return a / b

	}
	panic("unknown operator")
}

func builtinRelational[t int | float64](op bytecode.OpCode, a, b t) bool {
	switch op {
	case bytecode.LT:
		return a < b
	case bytecode.GT:
		return a > b
	case bytecode.LE:
		return a <= b
	case bytecode.GE:
		return a >= b
	}
	panic("unknown operator")
}

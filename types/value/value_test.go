package value_test

import (
	"math"
	"testing"

	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/value"
)

type TestDatum struct {
	Name     string
	Func     func() (value.Type, error)
	Expected value.Type
	Error    error
}

var emptyFunc = value.NewFunction(0, nil, 0, 0)

var testData = []TestDatum{
	{"Arithmetics int + int", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.ADD, value.NewInt(2)) }, value.NewInt(3), nil},
	{"Arithmetics int - int", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.SUB, value.NewInt(2)) }, value.NewInt(-1), nil},
	{"Arithmetics int * int", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.MUL, value.NewInt(2)) }, value.NewInt(2), nil},
	{"Arithmetics int / int", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.DIV, value.NewInt(2)) }, value.NewInt(0), nil},
	{"Arithmetics int / 0", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.DIV, value.NewInt(0)) }, value.Nil, value.ErrZeroDiv},

	{"Arithmetics float + float", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.ADD, value.NewFloat(2)) }, value.NewFloat(3), nil},
	{"Arithmetics float - float", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.SUB, value.NewFloat(2)) }, value.NewFloat(-1), nil},
	{"Arithmetics float * float", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.MUL, value.NewFloat(2)) }, value.NewFloat(2), nil},
	{"Arithmetics float / float", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.DIV, value.NewFloat(2)) }, value.NewFloat(0.5), nil},
	{"Arithmetics float / 0", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.DIV, value.NewFloat(0)) }, value.NewFloat(math.Inf(1)), nil},

	{"Arithmetics int + float", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.ADD, value.NewFloat(2)) }, value.NewFloat(3), nil},
	{"Arithmetics float + int", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.ADD, value.NewInt(2)) }, value.NewFloat(3), nil},
	{"Arithmetics int - float", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.SUB, value.NewFloat(2)) }, value.NewFloat(-1), nil},
	{"Arithmetics float - int", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.SUB, value.NewInt(2)) }, value.NewFloat(-1), nil},
	{"Arithmetics int * float", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.MUL, value.NewFloat(2)) }, value.NewFloat(2), nil},
	{"Arithmetics float * int", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.MUL, value.NewInt(2)) }, value.NewFloat(2), nil},
	{"Arithmetics int / float", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.DIV, value.NewFloat(2)) }, value.NewFloat(0.5), nil},
	{"Arithmetics float / int", func() (value.Type, error) { return value.NewFloat(1).Arith(bytecode.DIV, value.NewInt(2)) }, value.NewFloat(0.5), nil},

	{"Arithmetics string + string", func() (value.Type, error) { return value.NewString("a").Arith(bytecode.ADD, value.NewString("b")) }, value.NewString("ab"), nil},
	{"Arithmetics string - string", func() (value.Type, error) { return value.NewString("a").Arith(bytecode.SUB, value.NewString("b")) }, value.Nil, value.ErrType},
	{"Arithmetics array + array",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)})
			b := value.NewArray([]value.Type{value.NewInt(3), value.NewInt(4)})
			return a.Arith(bytecode.ADD, b)
		},
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3), value.NewInt(4)}), nil,
	},
	{"Arithmetics array - array",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)})
			b := value.NewArray([]value.Type{value.NewInt(3), value.NewInt(4)})
			return a.Arith(bytecode.SUB, b)
		},
		value.Nil,
		value.ErrType,
	},

	{"Arithmetics bool + bool", func() (value.Type, error) { return value.NewBool(true).Arith(bytecode.ADD, value.NewBool(true)) }, value.Nil, value.ErrType},
	{"Arithmetics function + function", func() (value.Type, error) { return emptyFunc.Arith(bytecode.ADD, emptyFunc) }, value.Nil, value.ErrType},
	{"Arithmetics int + nil", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.ADD, value.Nil) }, value.Nil, value.ErrNil},
	{"Arithmetics nil + int", func() (value.Type, error) { return value.Nil.Arith(bytecode.ADD, value.NewInt(1)) }, value.Nil, value.ErrNil},
	{"Arithmetics nil + nil", func() (value.Type, error) { return value.Nil.Arith(bytecode.ADD, value.Nil) }, value.Nil, value.ErrNil},
	{"Arithmetics int + function", func() (value.Type, error) { return value.NewInt(1).Arith(bytecode.ADD, emptyFunc) }, value.Nil, value.ErrType},
	{"Arithmetics function + int", func() (value.Type, error) { return emptyFunc.Arith(bytecode.ADD, value.NewInt(1)) }, value.Nil, value.ErrType},

	{"Modulo int % int", func() (value.Type, error) { return value.NewInt(5).Mod(value.NewInt(3)) }, value.NewInt(2), nil},
	{"Modulo int % float", func() (value.Type, error) { return value.NewInt(5).Mod(value.NewFloat(3)) }, value.Nil, value.ErrType},
	{"Modulo float % int", func() (value.Type, error) { return value.NewFloat(5).Mod(value.NewInt(3)) }, value.Nil, value.ErrType},
	{"Modulo float % float", func() (value.Type, error) { return value.NewFloat(5).Mod(value.NewFloat(3)) }, value.Nil, value.ErrType},
	{"Module int % nil", func() (value.Type, error) { return value.NewInt(5).Mod(value.Nil) }, value.Nil, value.ErrNil},
	{"Module nil % int", func() (value.Type, error) { return value.Nil.Mod(value.NewInt(5)) }, value.Nil, value.ErrNil},

	{"Logic bool & bool", func() (value.Type, error) { return value.NewBool(true).Logic(bytecode.AND, value.NewBool(false)) }, value.NewBool(false), nil},
	{"Logic bool | bool", func() (value.Type, error) { return value.NewBool(true).Logic(bytecode.OR, value.NewBool(false)) }, value.NewBool(true), nil},

	{"Logic int & int", func() (value.Type, error) { return value.NewInt(3).Logic(bytecode.AND, value.NewInt(6)) }, value.NewInt(2), nil},
	{"Logic int | int", func() (value.Type, error) { return value.NewInt(3).Logic(bytecode.OR, value.NewInt(6)) }, value.NewInt(7), nil},

	{"Logic float & float", func() (value.Type, error) { return value.NewFloat(1.0).Logic(bytecode.AND, value.NewFloat(2.0)) }, value.Nil, value.ErrType},
	{"Logic string & string", func() (value.Type, error) { return value.NewString("1").Logic(bytecode.AND, value.NewString("2")) }, value.Nil, value.ErrType},
	{"Logic array & array",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)})
			b := value.NewArray([]value.Type{value.NewInt(3), value.NewInt(4)})
			return a.Logic(bytecode.AND, b)
		},
		value.Nil,
		value.ErrType,
	},
	{"Logic function & function",
		func() (value.Type, error) { return emptyFunc.Logic(bytecode.AND, emptyFunc) },
		value.Nil,
		value.ErrType,
	},

	{"Logic int & float", func() (value.Type, error) { return value.NewInt(1).Logic(bytecode.AND, value.NewFloat(2)) }, value.Nil, value.ErrType},

	{"Logic nil & int", func() (value.Type, error) { return value.Nil.Logic(bytecode.AND, value.NewInt(1)) }, value.Nil, value.ErrNil},
	{"Logic int & nil", func() (value.Type, error) { return value.NewInt(1).Logic(bytecode.AND, value.Nil) }, value.Nil, value.ErrNil},
	{"Logic nil & nil", func() (value.Type, error) { return value.Nil.Logic(bytecode.AND, value.Nil) }, value.Nil, value.ErrNil},

	{"Shift int << int", func() (value.Type, error) { return value.NewInt(1).Shift(bytecode.LSH, value.NewInt(2)) }, value.NewInt(4), nil},
	{"Shift int >> int", func() (value.Type, error) { return value.NewInt(10).Shift(bytecode.RSH, value.NewInt(1)) }, value.NewInt(5), nil},
	{"Shift int << float", func() (value.Type, error) { return value.NewInt(1).Shift(bytecode.LSH, value.NewFloat(2)) }, value.Nil, value.ErrType},
	{"Shift int << bool", func() (value.Type, error) { return value.NewInt(1).Shift(bytecode.LSH, value.NewBool(true)) }, value.Nil, value.ErrType},
	{"Shift bool << bool", func() (value.Type, error) { return value.NewBool(true).Shift(bytecode.LSH, value.NewBool(true)) }, value.Nil, value.ErrType},
	{"Shift nil << nil", func() (value.Type, error) { return value.Nil.Shift(bytecode.LSH, value.Nil) }, value.Nil, value.ErrNil},

	{"Bit flip ~true", func() (value.Type, error) { return value.NewBool(true).Flip() }, value.Nil, value.ErrType},
	{"Bit flip ~false", func() (value.Type, error) { return value.NewBool(false).Flip() }, value.Nil, value.ErrType},
	{"Bit flip ~int", func() (value.Type, error) { return value.NewInt(1).Flip() }, value.NewInt(-2), nil},
	{"Bit flip ~float", func() (value.Type, error) { return value.NewFloat(1).Flip() }, value.Nil, value.ErrType},
	{"Bit flip ~string", func() (value.Type, error) { return value.NewString("1").Flip() }, value.Nil, value.ErrType},
	{"Bit flip ~array", func() (value.Type, error) { return value.NewArray([]value.Type{value.NewInt(1)}).Flip() }, value.Nil, value.ErrType},
	{"Bit flip ~function", emptyFunc.Flip, value.Nil, value.ErrType},
	{"Bit flip ~nil", func() (value.Type, error) { return value.Nil.Flip() }, value.Nil, value.ErrNil},

	{"Not !true", func() (value.Type, error) { return value.NewBool(true).Not() }, value.NewBool(false), nil},
	{"Not !false", func() (value.Type, error) { return value.NewBool(false).Not() }, value.NewBool(true), nil},
	{"Not !int", func() (value.Type, error) { return value.NewInt(1).Not() }, value.Nil, value.ErrType},
	{"Not !float", func() (value.Type, error) { return value.NewFloat(1).Not() }, value.Nil, value.ErrType},
	{"Not !string", func() (value.Type, error) { return value.NewString("1").Not() }, value.Nil, value.ErrType},
	{"Not !array", func() (value.Type, error) { return value.NewArray([]value.Type{value.NewInt(1)}).Not() }, value.Nil, value.ErrType},
	{"Not !function", emptyFunc.Not, value.Nil, value.ErrType},
	{"Not !nil", func() (value.Type, error) { return value.Nil.Not() }, value.Nil, value.ErrNil},

	{"Relational int < int", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.LT, value.NewInt(2)) }, value.NewBool(true), nil},
	{"Relational int <= int", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.LE, value.NewInt(2)) }, value.NewBool(true), nil},
	{"Relational int > int", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.GT, value.NewInt(2)) }, value.NewBool(false), nil},
	{"Relational int >= int", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.GE, value.NewInt(2)) }, value.NewBool(false), nil},

	{"Relational float < float", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.LT, value.NewFloat(2.0)) }, value.NewBool(true), nil},
	{"Relational float <= float", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.LE, value.NewFloat(2.0)) }, value.NewBool(true), nil},
	{"Relational float > float", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.GT, value.NewFloat(2.0)) }, value.NewBool(false), nil},
	{"Relational float >= float", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.GE, value.NewFloat(2.0)) }, value.NewBool(false), nil},

	{"Relational float < int", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.LT, value.NewInt(2)) }, value.NewBool(true), nil},
	{"Relational float <= int", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.LE, value.NewInt(2)) }, value.NewBool(true), nil},
	{"Relational float > int", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.GT, value.NewInt(2)) }, value.NewBool(false), nil},
	{"Relational float >= int", func() (value.Type, error) { return value.NewFloat(1.0).Relational(bytecode.GE, value.NewInt(2)) }, value.NewBool(false), nil},

	{"Relational int < float", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.LT, value.NewFloat(2)) }, value.NewBool(true), nil},
	{"Relational int <= float", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.LE, value.NewFloat(2)) }, value.NewBool(true), nil},
	{"Relational int > float", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.GT, value.NewFloat(2)) }, value.NewBool(false), nil},
	{"Relational int >= float", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.GE, value.NewFloat(2)) }, value.NewBool(false), nil},

	{"Relational string < string", func() (value.Type, error) { return value.NewString("a").Relational(bytecode.LT, value.NewString("b")) }, value.Nil, value.ErrType},
	{"Relational array < array",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewString("a"), value.NewString("b")})
			b := value.NewArray([]value.Type{value.NewString("c"), value.NewString("d")})
			return a.Relational(bytecode.LT, b)
		},
		value.Nil,
		value.ErrType,
	},
	{"Relational function < function",
		func() (value.Type, error) { return emptyFunc.Relational(bytecode.LT, emptyFunc) },
		value.Nil,
		value.ErrType,
	},

	{"Relational int < string", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.LT, value.NewString("b")) }, value.Nil, value.ErrType},

	{"Relational bool < bool", func() (value.Type, error) { return value.NewBool(true).Relational(bytecode.LT, value.NewBool(false)) }, value.Nil, value.ErrType},
	{"Relational int < nil", func() (value.Type, error) { return value.NewInt(1).Relational(bytecode.LT, value.Nil) }, value.Nil, value.ErrNil},
	{"Relational nil < int", func() (value.Type, error) { return value.Nil.Relational(bytecode.LT, value.NewInt(1)) }, value.Nil, value.ErrNil},
	{"Relational nil < nil", func() (value.Type, error) { return value.Nil.Relational(bytecode.LT, value.Nil) }, value.Nil, value.ErrNil},

	{"Index array[int]",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)})
			b := value.NewInt(1)
			return a.Index(b)
		}, value.NewInt(2), nil},

	{"Index string[int]", func() (value.Type, error) { return value.NewString("ab").Index(value.NewInt(1)) }, value.NewString("b"), nil},
	{"Index array[int:int]",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)})
			return a.Index(value.NewInt(1), value.NewInt(2))
		},
		value.NewArray([]value.Type{value.NewInt(2)}),
		nil,
	},

	{"Index string[int] outside", func() (value.Type, error) { return value.NewString("ab").Index(value.NewInt(5)) }, value.Nil, value.ErrIndex},
	{"Index array[int] outside",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)})
			return a.Index(value.NewInt(5))
		},
		value.Nil,
		value.ErrIndex,
	},
	{"Index array[int:int] indices equal",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)})
			return a.Index(value.NewInt(1), value.NewInt(1))
		},
		value.NewArray([]value.Type{}),
		nil,
	},
	{"Index array[int:int] indices backwards",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)})
			return a.Index(value.NewInt(2), value.NewInt(1))
		},
		value.Nil,
		value.ErrIndex,
	},
	{"Index array[int:int] indices past end",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)})
			return a.Index(value.NewInt(3), value.NewInt(3))
		},
		value.NewArray([]value.Type{}),
		nil,
	},
	{"Index string[bool]", func() (value.Type, error) { return value.NewString("ab").Index(value.NewBool(true)) }, value.Nil, value.ErrType},

	{"Len string", func() (value.Type, error) { return value.NewString("a").Len() }, value.NewInt(1), nil},
	{"Len array",
		func() (value.Type, error) { return value.NewArray([]value.Type{value.NewInt(1)}).Len() },
		value.NewInt(1),
		nil,
	},

	{"Len int", func() (value.Type, error) { return value.NewInt(1).Len() }, value.Nil, value.ErrType},
	{"Len float", func() (value.Type, error) { return value.NewFloat(1.0).Len() }, value.Nil, value.ErrType},
	{"Len bool", func() (value.Type, error) { return value.NewBool(true).Len() }, value.Nil, value.ErrType},
	{"Len function", emptyFunc.Len, value.Nil, value.ErrType},
	{"Len nil", func() (value.Type, error) { return value.Nil.Len() }, value.Nil, value.ErrNil},

	{"Equality int == int", func() (value.Type, error) { return value.NewInt(1).Eq(bytecode.EQ, value.NewInt(1)) }, value.NewBool(true), nil},
	{"Equality int != int", func() (value.Type, error) { return value.NewInt(1).Eq(bytecode.NE, value.NewInt(2)) }, value.NewBool(true), nil},
	{"Equality float == float", func() (value.Type, error) { return value.NewFloat(1.0).Eq(bytecode.EQ, value.NewFloat(1.0)) }, value.NewBool(true), nil},
	{"Equality float!= float", func() (value.Type, error) { return value.NewFloat(1.0).Eq(bytecode.NE, value.NewFloat(2.0)) }, value.NewBool(true), nil},

	{"Equality int == float", func() (value.Type, error) { return value.NewInt(1).Eq(bytecode.EQ, value.NewFloat(1.0)) }, value.NewBool(true), nil},
	{"Equality int!= float", func() (value.Type, error) { return value.NewInt(1).Eq(bytecode.NE, value.NewFloat(2.0)) }, value.NewBool(true), nil},
	{"Equality float == int", func() (value.Type, error) { return value.NewFloat(1.0).Eq(bytecode.EQ, value.NewInt(1)) }, value.NewBool(true), nil},
	{"Equality float!= int", func() (value.Type, error) { return value.NewFloat(1.0).Eq(bytecode.NE, value.NewInt(2)) }, value.NewBool(true), nil},

	{"Equality string == string", func() (value.Type, error) { return value.NewString("a").Eq(bytecode.EQ, value.NewString("a")) }, value.NewBool(true), nil},
	{"Equality string != string", func() (value.Type, error) { return value.NewString("a").Eq(bytecode.NE, value.NewString("b")) }, value.NewBool(true), nil},

	{"Equality array == array",
		func() (value.Type, error) {
			a := value.NewArray([]value.Type{value.NewInt(1)})
			b := value.NewArray([]value.Type{value.NewFloat(1.0)})
			return a.Eq(bytecode.EQ, b)
		},
		value.NewBool(true),
		nil,
	},

	{"Equality bool == bool", func() (value.Type, error) { return value.NewBool(true).Eq(bytecode.EQ, value.NewBool(true)) }, value.NewBool(true), nil},
	{"Equality bool != bool", func() (value.Type, error) { return value.NewBool(true).Eq(bytecode.NE, value.NewBool(false)) }, value.NewBool(true), nil},

	{"Equality nil == nil", func() (value.Type, error) { return value.Nil.Eq(bytecode.EQ, value.Nil) }, value.Nil, value.ErrNil},
	{"Equality nil != nil", func() (value.Type, error) { return value.Nil.Eq(bytecode.NE, value.Nil) }, value.Nil, value.ErrNil},

	{"Equality function == function", func() (value.Type, error) { return emptyFunc.Eq(bytecode.EQ, emptyFunc) }, value.NewBool(false), nil},

	{"Equality int == array", func() (value.Type, error) {
		return value.NewInt(1).Eq(bytecode.EQ, value.NewArray([]value.Type{value.NewInt(1)}))
	}, value.NewBool(false), nil},
	{"Equality int!= array", func() (value.Type, error) {
		return value.NewInt(1).Eq(bytecode.NE, value.NewArray([]value.Type{value.NewInt(2)}))
	}, value.NewBool(true), nil},

	{"Equality nil == int", func() (value.Type, error) { return value.Nil.Eq(bytecode.EQ, value.NewInt(1)) }, value.Nil, value.ErrNil},
}

func TestValue(t *testing.T) {
	for _, d := range testData {
		t.Run(d.Name, func(t *testing.T) {
			if v, err := d.Func(); !v.StrictEq(d.Expected) || err != d.Error {
				t.Errorf("Expected (%v, %v), got (%v, %v)", d.Expected, d.Error, v, err)
			}
		})
	}
}

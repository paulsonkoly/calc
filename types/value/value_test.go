package value_test

import (
	"math"
	"testing"

	"github.com/paulsonkoly/calc/types/value"
)

type TestDatum struct {
	Name     string
	A, B     value.Type
	Func     func(a, b value.Type) value.Type
	Expected value.Type
}

var emptyFunc = value.NewFunction(0, nil, 0, 0)

var testData = []TestDatum{
	{"Arithmetics int + int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.NewInt(3)},
	{"Arithmetics int - int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("-", b) }, value.NewInt(-1)},
	{"Arithmetics int * int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("*", b) }, value.NewInt(2)},
	{"Arithmetics int / int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("/", b) }, value.NewInt(0)},
	{"Arithmetics int / 0", value.NewInt(1), value.NewInt(0), func(a, b value.Type) value.Type { return a.Arith("/", b) }, value.ZeroDivError},

	{"Arithmetics float + float", value.NewFloat(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.NewFloat(3)},
	{"Arithmetics float - float", value.NewFloat(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("-", b) }, value.NewFloat(-1)},
	{"Arithmetics float * float", value.NewFloat(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("*", b) }, value.NewFloat(2)},
	{"Arithmetics float / float", value.NewFloat(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("/", b) }, value.NewFloat(0.5)},
	{"Arithmetics float / 0", value.NewFloat(1), value.NewFloat(0), func(a, b value.Type) value.Type { return a.Arith("/", b) }, value.NewFloat(math.Inf(1))},

	{"Arithmetics int + float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.NewFloat(3)},
	{"Arithmetics float + int", value.NewFloat(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.NewFloat(3)},
	{"Arithmetics int - float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("-", b) }, value.NewFloat(-1)},
	{"Arithmetics float - int", value.NewFloat(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("-", b) }, value.NewFloat(-1)},
	{"Arithmetics int * float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("*", b) }, value.NewFloat(2)},
	{"Arithmetics float * int", value.NewFloat(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("*", b) }, value.NewFloat(2)},
	{"Arithmetics int / float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Arith("/", b) }, value.NewFloat(0.5)},
	{"Arithmetics float / int", value.NewFloat(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Arith("/", b) }, value.NewFloat(0.5)},

	{"Arithmetics string + string", value.NewString("a"), value.NewString("b"), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.NewString("ab")},
	{"Arithmetics string - string", value.NewString("a"), value.NewString("b"), func(a, b value.Type) value.Type { return a.Arith("-", b) }, value.InvalidOpError},
	{"Arithmetics array + array",
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)}),
		value.NewArray([]value.Type{value.NewInt(3), value.NewInt(4)}),
		func(a, b value.Type) value.Type { return a.Arith("+", b) },
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3), value.NewInt(4)}),
	},
	{"Arithmetics array - array",
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)}),
		value.NewArray([]value.Type{value.NewInt(3), value.NewInt(4)}),
		func(a, b value.Type) value.Type { return a.Arith("-", b) },
		value.InvalidOpError,
	},

	{"Arithmetics bool + bool", value.NewBool(true), value.NewBool(true), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.InvalidOpError},
	{"Arithmetics function + function", emptyFunc, emptyFunc, func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.InvalidOpError},
	{"Arithmetics int + error", value.NewInt(1), value.TypeError, func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.TypeError},
	{"Arithmetics error + int", value.TypeError, value.NewInt(1), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.TypeError},
	{"Arithmetics error + error", value.ZeroDivError, value.IndexError, func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.ZeroDivError},
	{"Arithmetics int + function", value.NewInt(1), emptyFunc, func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.TypeError},
	{"Arithmetics function + int", emptyFunc, value.NewInt(1), func(a, b value.Type) value.Type { return a.Arith("+", b) }, value.TypeError},

	{"Modulo int % int", value.NewInt(5), value.NewInt(3), func(a, b value.Type) value.Type { return a.Mod(b) }, value.NewInt(2)},
	{"Modulo int % float", value.NewInt(5), value.NewFloat(3), func(a, b value.Type) value.Type { return a.Mod(b) }, value.TypeError},
	{"Modulo float % int", value.NewFloat(5), value.NewInt(3), func(a, b value.Type) value.Type { return a.Mod(b) }, value.TypeError},
	{"Modulo float % float", value.NewFloat(5), value.NewFloat(3), func(a, b value.Type) value.Type { return a.Mod(b) }, value.InvalidOpError},
	{"Module int % error", value.NewInt(5), value.IndexError, func(a, b value.Type) value.Type { return a.Mod(b) }, value.IndexError},
	{"Module error % int", value.NewFloat(5), value.IndexError, func(a, b value.Type) value.Type { return a.Mod(b) }, value.IndexError},

	{"Logic bool & bool", value.NewBool(true), value.NewBool(false), func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.NewBool(false)},
	{"Logic bool | bool", value.NewBool(true), value.NewBool(false), func(a, b value.Type) value.Type { return a.Logic("|", b) }, value.NewBool(true)},

	{"Logic int & int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.InvalidOpError},
	{"Logic float & float", value.NewFloat(1.0), value.NewFloat(2.0), func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.InvalidOpError},
	{"Logic string & string", value.NewString("1"), value.NewString("2"), func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.InvalidOpError},
	{"Logic array & array",
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)}),
		value.NewArray([]value.Type{value.NewInt(3), value.NewInt(4)}),
		func(a, b value.Type) value.Type { return a.Logic("&", b) },
		value.InvalidOpError},
	{"Logic function & function",
		emptyFunc,
		emptyFunc,
		func(a, b value.Type) value.Type { return a.Logic("&", b) },
		value.InvalidOpError},

	{"Logic int & float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.TypeError},

	{"Logic error & int", value.ZeroDivError, value.NewInt(1), func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.ZeroDivError},
	{"Logic int & error", value.NewInt(1), value.ZeroDivError, func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.ZeroDivError},
	{"Logic error & error", value.ZeroDivError, value.IndexError, func(a, b value.Type) value.Type { return a.Logic("&", b) }, value.ZeroDivError},

	{"Not !true", value.NewBool(true), value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.NewBool(false)},
	{"Not !false", value.NewBool(false), value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.NewBool(true)},
	{"Not !int", value.NewInt(1), value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.TypeError},
	{"Not !float", value.NewFloat(1), value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.TypeError},
	{"Not !string", value.NewString("1"), value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.TypeError},
	{"Not !array", value.NewArray([]value.Type{value.NewInt(1)}), value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.TypeError},
	{"Not !function", emptyFunc, value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.TypeError},
	{"Not !error", value.ZeroDivError, value.NewBool(false), func(a, b value.Type) value.Type { return a.Not() }, value.ZeroDivError},

	{"Relational int < int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.NewBool(true)},
	{"Relational int <= int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational("<=", b) }, value.NewBool(true)},
	{"Relational int > int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational(">", b) }, value.NewBool(false)},
	{"Relational int >= int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational(">=", b) }, value.NewBool(false)},

	{"Relational float < float", value.NewFloat(1.0), value.NewFloat(2.0), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.NewBool(true)},
	{"Relational float <= float", value.NewFloat(1.0), value.NewFloat(2.0), func(a, b value.Type) value.Type { return a.Relational("<=", b) }, value.NewBool(true)},
	{"Relational float > float", value.NewFloat(1.0), value.NewFloat(2.0), func(a, b value.Type) value.Type { return a.Relational(">", b) }, value.NewBool(false)},
	{"Relational float >= float", value.NewFloat(1.0), value.NewFloat(2.0), func(a, b value.Type) value.Type { return a.Relational(">=", b) }, value.NewBool(false)},

	{"Relational float < int", value.NewFloat(1.0), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.NewBool(true)},
	{"Relational float <= int", value.NewFloat(1.0), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational("<=", b) }, value.NewBool(true)},
	{"Relational float > int", value.NewFloat(1.0), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational(">", b) }, value.NewBool(false)},
	{"Relational float >= int", value.NewFloat(1.0), value.NewInt(2), func(a, b value.Type) value.Type { return a.Relational(">=", b) }, value.NewBool(false)},

	{"Relational int < float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.NewBool(true)},
	{"Relational int <= float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Relational("<=", b) }, value.NewBool(true)},
	{"Relational int > float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Relational(">", b) }, value.NewBool(false)},
	{"Relational int >= float", value.NewInt(1), value.NewFloat(2), func(a, b value.Type) value.Type { return a.Relational(">=", b) }, value.NewBool(false)},

	{"Relational string < string", value.NewString("a"), value.NewString("b"), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.InvalidOpError},
	{"Relational array < array",
		value.NewArray([]value.Type{value.NewString("a"), value.NewString("b")}),
		value.NewArray([]value.Type{value.NewString("c"), value.NewString("d")}),
		func(a, b value.Type) value.Type { return a.Relational("<", b) },
		value.InvalidOpError,
	},
	{"Relational function < function",
		emptyFunc,
		emptyFunc,
		func(a, b value.Type) value.Type { return a.Relational("<", b) },
		value.InvalidOpError,
	},

	{"Relational int < string", value.NewInt(1), value.NewString("b"), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.TypeError},

	{"Relational bool < bool", value.NewBool(true), value.NewBool(false), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.InvalidOpError},
	{"Relational int < error", value.NewInt(1), value.ZeroDivError, func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.ZeroDivError},
	{"Relational error < int", value.ZeroDivError, value.NewInt(1), func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.ZeroDivError},
	{"Relational error < error", value.ZeroDivError, value.IndexError, func(a, b value.Type) value.Type { return a.Relational("<", b) }, value.ZeroDivError},

	{"Index array[int]", value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)}), value.NewInt(1), func(a, b value.Type) value.Type { return a.Index(b) }, value.NewInt(2)},

	{"Index string[int]", value.NewString("ab"), value.NewInt(1), func(a, b value.Type) value.Type { return a.Index(b) }, value.NewString("b")},
	{"Index array[int:int]",
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)}),
		value.NewInt(1),
		func(a, b value.Type) value.Type { return a.Index(value.NewInt(1), value.NewInt(2)) },
		value.NewArray([]value.Type{value.NewInt(2)}),
	},

	{"Index string[int] outside", value.NewString("ab"), value.NewInt(5), func(a, b value.Type) value.Type { return a.Index(b) }, value.IndexError},
	{"Index array[int] outside",
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2)}),
		value.NewInt(5), func(a, b value.Type) value.Type { return a.Index(b) },
		value.IndexError,
	},
	{"Index array[int:int] indices equal",
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)}),
		value.NewInt(5),
		func(a, b value.Type) value.Type { return a.Index(value.NewInt(1), value.NewInt(1)) },
		value.NewArray([]value.Type{}),
	},
	{"Index array[int:int] indices backwards",
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)}),
		value.NewInt(5),
		func(a, b value.Type) value.Type { return a.Index(value.NewInt(2), value.NewInt(1)) },
		value.IndexError,
	},
	{"Index string[bool]", value.NewString("ab"), value.NewBool(true), func(a, b value.Type) value.Type { return a.Index(b) }, value.TypeError},

	{"Len string", value.NewString("a"), value.NewInt(1), func(a, b value.Type) value.Type { return a.Len() }, value.NewInt(1)},
	{"Len array",
		value.NewArray([]value.Type{value.NewInt(1)}),
		value.NewInt(1),
		func(a, b value.Type) value.Type { return a.Len() },
		value.NewInt(1),
	},

	{"Len int", value.NewInt(1), value.NewInt(1), func(a, b value.Type) value.Type { return a.Len() }, value.TypeError},
	{"Len float", value.NewFloat(1.0), value.NewInt(1), func(a, b value.Type) value.Type { return a.Len() }, value.TypeError},
	{"Len bool", value.NewBool(true), value.NewInt(1), func(a, b value.Type) value.Type { return a.Len() }, value.TypeError},
	{"Len function", emptyFunc, value.NewInt(1), func(a, b value.Type) value.Type { return a.Len() }, value.TypeError},
	{"Len error", value.ZeroDivError, value.NewInt(1), func(a, b value.Type) value.Type { return a.Len() }, value.ZeroDivError},

	{"Equality int == int", value.NewInt(1), value.NewInt(1), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(true)},
	{"Equality int != int", value.NewInt(1), value.NewInt(2), func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},
	{"Equality float == float", value.NewFloat(1.0), value.NewFloat(1.0), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(true)},
	{"Equality float!= float", value.NewFloat(1.0), value.NewFloat(2.0), func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},

	{"Equality int == float", value.NewInt(1), value.NewFloat(1.0), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(true)},
	{"Equality int!= float", value.NewInt(1), value.NewFloat(2.0), func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},
	{"Equality float == int", value.NewFloat(1.0), value.NewInt(1), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(true)},
	{"Equality float!= int", value.NewFloat(1.0), value.NewInt(2), func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},

	{"Equality string == string", value.NewString("a"), value.NewString("a"), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(true)},
	{"Equality string != string", value.NewString("a"), value.NewString("b"), func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},

	{"Equality array == array",
		value.NewArray([]value.Type{value.NewInt(1)}),
		value.NewArray([]value.Type{value.NewFloat(1.0)}),
		func(a, b value.Type) value.Type { return a.Eq("==", b) },
		value.NewBool(true),
	},

	{"Equality bool == bool", value.NewBool(true), value.NewBool(true), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(true)},
	{"Equality bool != bool", value.NewBool(true), value.NewBool(false), func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},

	{"Equality error == error", value.IndexError, value.IndexError, func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(true)},
	{"Equality error != error", value.IndexError, value.ZeroDivError, func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},

	{"Equality function == function", emptyFunc, emptyFunc, func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(false)},

	{"Equality int == array", value.NewInt(1), value.NewArray([]value.Type{value.NewInt(1)}), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(false)},
	{"Equality int!= array", value.NewInt(1), value.NewArray([]value.Type{value.NewInt(2)}), func(a, b value.Type) value.Type { return a.Eq("!=", b) }, value.NewBool(true)},

	{"Equality error== int", value.IndexError, value.NewInt(1), func(a, b value.Type) value.Type { return a.Eq("==", b) }, value.NewBool(false)},
}

func TestValue(t *testing.T) {
	for _, d := range testData {
		t.Run(d.Name, func(t *testing.T) {
			if v := d.Func(d.A, d.B); !v.StrictEq(d.Expected) {
				t.Errorf("Expected %v, got %v", d.Expected, v)
			}
		})
	}
}

package main_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
	"github.com/stretchr/testify/assert"
)

type TestDatum struct {
	name       string
	input      string
	parseError error
	value      value.Type
}

var messages = []string{"a not defined", "hi"}

var testData = [...]TestDatum{
	{"simple literal/integer", "1", nil, value.NewInt(1)},
	{"simple literal/float", "3.14", nil, value.NewFloat(3.14)},
	{"simple literal/bool", "false", nil, value.NewBool(false)},
	{"simple literal/string", "\"abc\"", nil, value.NewString("abc")},
	{"simple literal/array empty", "[]", nil, value.NewArray([]value.Type{})},
	{"simple literal/array", "[1, false]", nil, value.NewArray([]value.Type{value.NewInt(1), value.NewBool(false)})},

	{"simple arithmetic/addition", "1+2", nil, value.NewInt(3)},
	{"simple arithmetic/coercion", "1+2.0", nil, value.NewFloat(3)},
	{"simple arithmetic/coercion", "1.0+2", nil, value.NewFloat(3)},

	{"string indexing/simple", "\"apple\" @ 1", nil, value.NewString("p")},
	{"string indexing/complex empty", "\"apple\" @ 1 : 1", nil, value.NewString("")},
	{"string indexing/complex", "\"apple\" @ 1 : 3", nil, value.NewString("pp")},
	{"string indexing/outside", "\"apple\" @ (-1)", nil, value.IndexError},

	{"string concatenation", "\"abc\" + \"def\"", nil, value.NewString("abcdef")},
	{"string equality/false", "\"abc\" == \"ab\"", nil, value.NewBool(false)},
	{"string equality/true", "\"abc\" == \"abc\"", nil, value.NewBool(true)},

	{"arithmetics/left assoc", "1-2+1", nil, value.NewInt(0)},
	{"arithmetics/parenthesis", "1-(2+1)", nil, value.NewInt(-2)},

	{"variable/not defined", "a", nil, value.NewError(&messages[0])},
	{"variable/lookup", "{\na=3\na+1\n}", nil, value.NewInt(4)},

	{"relop/int==int true", "1==1", nil, value.NewBool(true)},
	{"relop/int==float true", "1==1.0", nil, value.NewBool(true)},
	{"relop/float==int true", "1.0==1", nil, value.NewBool(true)},
	{"relop/float==float true", "1.0==1.0", nil, value.NewBool(true)},
	{"relop/bool==bool true", "false==false", nil, value.NewBool(true)},

	{"relop/int!=int false", "1!=1", nil, value.NewBool(false)},
	{"relop/int!=float false", "1!=1.0", nil, value.NewBool(false)},
	{"relop/float!=int false", "1.0!=1", nil, value.NewBool(false)},
	{"relop/float!=float false", "1.0!=1.0", nil, value.NewBool(false)},
	{"relop/bool!=bool false", "false!=false", nil, value.NewBool(false)},

	{"relop/float accuracy", "1==0.9999999", nil, value.NewBool(false)},

	{"relop/int<int false", "1<1", nil, value.NewBool(false)},
	{"relop/int<float false", "1<1.0", nil, value.NewBool(false)},
	{"relop/float<int false", "1.0<1", nil, value.NewBool(false)},
	{"relop/float<float false", "1.0<1.0", nil, value.NewBool(false)},
	{"relop/bool<bool", "false<false", nil, value.InvalidOpError},

	{"relop/int<=int true", "1<=1", nil, value.NewBool(true)},
	{"relop/int<=float true", "1<=1.0", nil, value.NewBool(true)},
	{"relop/float<=int true", "1.0<=1", nil, value.NewBool(true)},
	{"relop/float<=float true", "1.0<=1.0", nil, value.NewBool(true)},
	{"relop/bool<=bool", "true<=true", nil, value.InvalidOpError},

	{"logicop/bool&bool true", "true&true", nil, value.NewBool(true)},
	{"logicop/bool&bool false", "true&false", nil, value.NewBool(false)},
	{"logicop/bool|bool true", "true|false", nil, value.NewBool(true)},
	{"logicop/bool|bool false", "false|false", nil, value.NewBool(false)},
	{"logicop/bool|int", "false|1", nil, value.TypeError},

	{"block/single line", "{\n1\n}", nil, value.NewInt(1)},
	{"block/multi line", "{\n1\n2\n}", nil, value.NewInt(2)},

	{"conditional/single line no else", "if true 1", nil, value.NewInt(1)},
	{"conditional/single line else", "if false 1 else 2", nil, value.NewInt(2)},
	{"conditional/incorrect condition", "if 1 1", nil, value.TypeError},
	{"conditional/no result", "if false 1", nil, value.NoResultError},
	{"conditional/blocks no else", "if true {\n1\n}", nil, value.NewInt(1)},
	{"conditional/blocks with else", "if false {\n1\n} else {\n2\n}", nil, value.NewInt(2)},

	{"loop/single line",
		`{
		a = 1
		while a < 10 a = a + 1
		a
	}`, nil, value.NewInt(10)},
	{"loop/block",
		`{
		a = 1
		while a < 10 {
			a = a + 1
		}
		a
	}`, nil, value.NewInt(10)},
	{"loop/false initial condition",
		`{
		while false {
			a = a + 1
		}
	}`, nil, value.NoResultError},
	{"loop/incorrect condition",
		`{
		while 13 {
			a = a + 1
		}
	}`, nil, value.TypeError},

	{"function definition", "(n) -> 1", nil, value.NewFunction(nil, nil)},
	{"function/no argument", "() -> 1", nil, value.NewFunction(nil, nil)},
	{"function/block",
		`(n) -> {
			n + 1
	  }`, nil, value.NewFunction(nil, nil)},

	{"call",
		`{
			a = (n) -> 1
			a(2)
		}`, nil, value.NewInt(1),
	},
	{"call/no argument",
		`{
			a = () -> 1
			a()
		}`, nil, value.NewInt(1),
	},
	{"function/return",
		`{
			a = (n) -> {
	       return 1
	       2
	     }
			a(2)
		}`, nil, value.NewInt(1),
	},
	{"function/closure",
		`{
			f = (a) -> {
	       (b) -> a + b
	     }
			x = f(1)
	     x(2)
		}`, nil, value.NewInt(3),
	},
	{"keyword violation", "true = false", errors.New("Parser: "), value.Type{}},
	{"builtin/aton int", "aton(\"12\")", nil, value.NewInt(12)},
	{"builtin/aton float", "aton(\"1.2\")", nil, value.NewFloat(1.2)},
	{"builtin/aton error", "aton(\"abc\")", nil, value.ConversionError},

	{"builtin/error", "error(\"hi\")", nil, value.NewError(&messages[1])},
	{"builtin/error type error", "error(1)", nil, value.TypeError},
	{"qsort",
		`{
	       filter = (pred, ary) -> {
	         i = 0
	         r = []
	         while i < #ary {
	           if pred(ary@i) r = r + [ary@i]
	           i = i + 1
	         }
	         r
	       }
	       qsort = (ary) -> {
	         if #ary <= 1 ary else {
	           pivot = ary@0
	           tail = ary @ 1 : #ary
	           qsort(filter((n) -> n <= pivot, tail)) + [pivot] + qsort(filter((n) -> n > pivot, tail))
	         }
	       }
	       qsort([5, 2, 4, 3, 1, 8])
	    }`,
		nil,
		value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3), value.NewInt(4), value.NewInt(5), value.NewInt(8)}),
	},
}

func TestCalc(t *testing.T) {
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			m := memory.NewMemory()
			builtin.Load(m)
			ast, err := parser.Parse(test.input)
			if test.parseError == nil {
				assert.NoError(t, err)
				var v value.Type
				for _, stmnt := range ast {
					stmnt = stmnt.STRewrite(node.SymTbl{})
					v = node.Evaluate(m, stmnt)
				}

				if !test.value.StrictEq(v) {
					t.Errorf("expected %v got %v", test.value, v)
				}
			} else {
				if !strings.HasPrefix(err.Error(), test.parseError.Error()) {
					t.Errorf("not the expected error: %s %s", test.parseError.Error(), err.Error())
				}
			}
		})
	}
}

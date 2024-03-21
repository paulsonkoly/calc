package main_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
	"github.com/paulsonkoly/calc/vm"
	"github.com/stretchr/testify/assert"
)

type TestDatum struct {
	name       string
	input      string
	parseError error
	value      value.Type
}

var messages = []string{"a not defined", "hi"}
var emptyFunction = value.NewFunction(0, nil, 0, 0)

var testData = [...]TestDatum{
	{"simple literal/integer", "1", nil, value.NewInt(1)},
	{"simple literal/float", "3.14", nil, value.NewFloat(3.14)},
	{"simple literal/bool", "false", nil, value.NewBool(false)},
	{"simple literal/string", "\"abc\"", nil, value.NewString("abc")},
	{"simple literal/array empty", "[]", nil, value.NewArray([]value.Type{})},
	{"simple literal/array", "[1, false]", nil, value.NewArray([]value.Type{value.NewInt(1), value.NewBool(false)})},
	{"array lit with newline", "[1,2,\n3,4]", nil, value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3), value.NewInt(4)})},
	{"array lit with leading newline", "[\n1,2,\n3,4]", nil, value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3), value.NewInt(4)})},

	{"simple arithmetic/addition", "1+2", nil, value.NewInt(3)},

	{"string indexing/simple", "\"apple\"[1]", nil, value.NewString("p")},
	{"string indexing/complex empty", "\"apple\" [ 1 : 1]", nil, value.NewString("")},

	{"indices/all from stack", "\"apple\"[ 1+0 : 3+0 ]", nil, value.NewString("pp")},

	{"string concatenation", "\"abc\" + \"def\"", nil, value.NewString("abcdef")},

	{"arithmetics/left assoc", "1-2+1", nil, value.NewInt(0)},
	{"arithmetics/parenthesis", "1-(2+1)", nil, value.NewInt(-2)},

	{"variable/not defined", "a", nil, value.NewError(&messages[0])},
	{"variable/lookup", "{\na=3\na+1\n}", nil, value.NewInt(4)},

	{"relop/int==int true", "1==1", nil, value.NewBool(true)},

	{"relop/int!=int false", "1!=1", nil, value.NewBool(false)},

	{"relop/float accuracy", "1==0.9999999", nil, value.NewBool(false)},

	{"relop/int<int false", "1<1", nil, value.NewBool(false)},

	{"relop/int<=int true", "1<=1", nil, value.NewBool(true)},

	{"logicop/bool&bool true", "true&true", nil, value.NewBool(true)},

  {"bool or/low precedence", "true||false == false", nil, value.NewBool(true)},
  {"bool or/high precedence", "true|false == false", nil, value.NewBool(false)},

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

	{"iterator/elems",
		`{
      c = 0
      for i <- elems([2,5,7]) c = c+i
   }`, nil, value.NewInt(14)},
	{"iterator/no yield", "for i <- 1 2", nil, value.NoResultError},
	{"iterator/return", "for i<- fromto(5, 10) if i == 8 return 3*i else 2*i", nil, value.NewInt(24)},
	{"iterator/yield in for",
		`{
    f = () -> for i <- fromto(2,5) {
      if i % 2 == 0 yield i
    }
    c = 0
    for i <- f() c = c+i
    }`, nil, value.NewInt(6)},

	{"iterator/for in for",
		`{
      c = 0
      for i <- fromto(1,2) {
        for j <- fromto(10, 12) {
          c = c + i + j
        }
      }
    }`, nil, value.NewInt(23)},

	{"function definition", "(n) -> 1", nil, emptyFunction},
	{"function/no argument", "() -> 1", nil, emptyFunction},
	{"function/block",
		`(n) -> {
			n + 1
	  }`, nil, emptyFunction},

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
	{"naked return", "return 1", nil, value.NewInt(1)},
	{"function/closure",
		`{
			f = (a) -> {
	       (b) -> a + b
	     }
			x = f(1)
	     x(2)
		}`, nil, value.NewInt(3),
	},

  {"array addition/doesn't share sub-slices",
  `{
    a = [1,2,3]
    b = a[1:2] + [1]
    a
  }`, nil, value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)}),
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
	           if pred(ary[i]) r = r + [ary[i]]
	           i = i + 1
	         }
	         r
	       }
	       qsort = (ary) -> {
	         if #ary <= 1 ary else {
	           pivot = ary[0]
	           tail = ary [1:#ary]
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

			m := memory.New()
			cs := []bytecode.Type{}
			ds := []value.Type{}
			builtin.Load(&cs, &ds)
			virtM := vm.New(m, &cs, &ds)

			ast, err := parser.Parse(test.input)
			if test.parseError == nil {
				assert.NoError(t, err)
				var v value.Type
				for _, stmnt := range ast {
					stmnt = stmnt.STRewrite(node.SymTbl{})
					node.ByteCode(stmnt, &cs, &ds)
					v = virtM.Run(true)
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

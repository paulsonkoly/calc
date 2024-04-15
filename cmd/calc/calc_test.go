package main_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/compresult"
	"github.com/paulsonkoly/calc/types/dbginfo"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
	"github.com/paulsonkoly/calc/vm"
)

type TestDatum struct {
	name         string
	input        string
	parseError   error
	value        value.Type
	runtimeError error
}

var emptyFunction = value.NewFunction(0, nil, 0, 0)

var testData = [...]TestDatum{
	{"simple literal/integer", "1", nil, value.NewInt(1), nil},
	{"simple literal/float", "3.14", nil, value.NewFloat(3.14), nil},
	{"simple literal/bool", "false", nil, value.NewBool(false), nil},
	{"simple literal/string", "\"abc\"", nil, value.NewString("abc"), nil},
	{"simple literal/array empty", "[]", nil, value.NewArray([]value.Type{}), nil},
	{"simple literal/array", "[1, false]", nil, value.NewArray([]value.Type{value.NewInt(1), value.NewBool(false)}), nil},
	{"array lit with newline", "[1,2,\n3,4]", nil, value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3), value.NewInt(4)}), nil},
	{"array lit with leading newline", "[\n1,2,\n3,4]", nil, value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3), value.NewInt(4)}), nil},

	{"simple arithmetic/addition", "1+2", nil, value.NewInt(3), nil},
	{"bitwise logic", "~(1<<1) & 7", nil, value.NewInt(5), nil},

	{"string indexing/simple", "\"apple\"[1]", nil, value.NewString("p"), nil},
	{"string indexing/complex empty", "\"apple\" [ 1 : 1]", nil, value.NewString(""), nil},
	{"indices/all from stack", "\"apple\"[ 1+0 : 3+0 ]", nil, value.NewString("pp"), nil},
	{"indexing/multidimensional", "[[1,2], 3, 4][0][1]", nil, value.NewInt(2), nil},
	{"indexing/from stack", "[1, 2, 3][4-2]", nil, value.NewInt(3), nil},

	{"string concatenation", "\"abc\" + \"def\"", nil, value.NewString("abcdef"), nil},

	{"arithmetics/left assoc", "1-2+1", nil, value.NewInt(0), nil},
	{"arithmetics/parenthesis", "1-(2+1)", nil, value.NewInt(-2), nil},

	{"variable/not defined", "a", nil, value.Nil, nil},
	{"variable/lookup", "{\na=3\na+1\n}", nil, value.NewInt(4), nil},

	{"relop/int==int true", "1==1", nil, value.NewBool(true), nil},

	{"relop/int!=int false", "1!=1", nil, value.NewBool(false), nil},

	{"relop/float accuracy", "1==0.9999999", nil, value.NewBool(false), nil},

	{"relop/int<int false", "1<1", nil, value.NewBool(false), nil},

	{"relop/int<=int true", "1<=1", nil, value.NewBool(true), nil},

	{"logicop/bool&bool true", "true&true", nil, value.NewBool(true), nil},

	{"bool or/low precedence", "true||false == false", nil, value.NewBool(true), nil},
	{"bool or/high precedence", "true|false == false", nil, value.NewBool(false), nil},

	{"block/single line", "{\n1\n}", nil, value.NewInt(1), nil},
	{"block/multi line", "{\n1\n2\n}", nil, value.NewInt(2), nil},

	{"conditional/single line no else", "if true 1", nil, value.NewInt(1), nil},
	{"conditional/single line else", "if false 1 else 2", nil, value.NewInt(2), nil},
	{"conditional/incorrect condition", "if 1 1", nil, value.Nil, value.ErrType},
	{"conditional/no result", "if false 1", nil, value.Nil, nil},
	{"conditional/blocks no else", "if true {\n1\n}", nil, value.NewInt(1), nil},
	{"conditional/blocks with else", "if false {\n1\n} else {\n2\n}", nil, value.NewInt(2), nil},

	{"loop/single line",
		`{
		a = 1
		while a < 10 a = a + 1
		a
	}`, nil, value.NewInt(10), nil},
	{"loop/block",
		`{
		a = 1
		while a < 10 {
			a = a + 1
		}
		a
	}`, nil, value.NewInt(10), nil},
	{"loop/false initial condition",
		`{
		while false {
			a = a + 1
		}
	}`, nil, value.Nil, nil},
	{"loop/incorrect condition",
		`{
		while 13 {
			a = a + 1
		}
	}`, nil, value.Nil, value.ErrType},

	{"iterator/elems",
		`{
      c = 0
      for i <- elems([2,5,7]) c = c+i
   }`, nil, value.NewInt(14), nil},
	{"iterator/no yield", "for i <- 1 2", nil, value.Nil, nil},
	{"iterator/return", "for i<- fromto(5, 10) if i == 8 return 3*i else 2*i", nil, value.NewInt(24), nil},
	{"iterator/yield in for",
		`{
    f = () -> for i <- fromto(2,5) {
      if i % 2 == 0 yield i
    }
    c = 0
    for i <- f() c = c+i
    }`, nil, value.NewInt(6), nil},

	{"iterator/for in for",
		`{
      c = 0
      for i <- fromto(1,2) {
        for j <- fromto(10, 12) {
          c = c + i + j
        }
      }
    }`, nil, value.NewInt(23), nil},
	{"iterator/parallel for",
		`{
      c = ""
      for i, j <- fromto(1,3), elems("ab") {
        c = c + toa(i) + " " + j + " "
      }
    }`, nil, value.NewString("1 a 2 b "), nil},
	{"iterator/parallel for mismatching finish",
		`{
      c = ""
      for i, j <- fromto(1,5), elems("ab") {
        c = c + toa(i) + " " + j + " "
      }
    }`, nil, value.NewString("1 a 2 b "), nil},
	{"iterator/parallel for in recursion",
		`{
       f = (n) -> {
         c = 0
         if n == 0 return 1
         for i, j <- fromto(1, 3), fromto(10, 13) c = c + i + j + f(n - 1) 
       }
       f(2)
    }`, nil, value.NewInt(76), nil},

	{"function definition", "(n) -> 1", nil, emptyFunction, nil},
	{"function/no argument", "() -> 1", nil, emptyFunction, nil},
	{"function/block",
		`(n) -> {
			n + 1
	  }`, nil, emptyFunction, nil},

	{"call",
		`{
			a = (n) -> 1
			a(2)
		}`, nil, value.NewInt(1), nil,
	},
	{"call/no argument",
		`{
			a = () -> 1
			a()
		}`, nil, value.NewInt(1), nil,
	},
	{"function/return",
		`{
			a = (n) -> {
	       return 1
	       2
	     }
			a(2)
		}`, nil, value.NewInt(1), nil,
	},
	{"naked return", "return 1", nil, value.NewInt(1), nil},
	{"function/closure",
		`{
			f = (a) -> {
	       (b) -> a + b
	     }
			x = f(1)
	     x(2)
		}`, nil, value.NewInt(3), nil,
	},
	{"function/closure variable updates",
		`{
			f = () -> {
        x = 1
        g = () -> x
        x = 2
        g
	    }
			g = f()
	    g()
		}`, nil, value.NewInt(2), nil,
	},

	{"array addition/doesn't share sub-slices",
		`{
    a = [1,2,3]
    b = a[1:2] + [1]
    a
  }`, nil, value.NewArray([]value.Type{value.NewInt(1), value.NewInt(2), value.NewInt(3)}), nil,
	},
	{"keyword violation", "true = false", errors.New("Parser: "), value.Nil, nil},
	{"builtin/aton int", "aton(\"12\")", nil, value.NewInt(12), nil},
	{"builtin/aton float", "aton(\"1.2\")", nil, value.NewFloat(1.2), nil},
	{"builtin/aton error", "aton(\"abc\")", nil, value.Nil, vm.ErrConversion},

	{"uninitialised local",
		`{
    f = () -> {
      if false a = 1
      a
    }
    f()
  }`, nil, value.Nil, nil,
	},

	{"regression/tmp optimisation", `aton("1" + "2") + 3`, nil, value.NewInt(15), nil},
	{"regression/function in for",
		`{
       f = () -> for i <- fromto(1,2) {
         yield () -> return 13
       }

       for i <- f() i()
     }`, nil, value.NewInt(13), nil,
	},
	// when one of the iterators never yield
	{"regression/parallel for value", "for i, j <- elems(\"ab\"), fromto(1, 1) 10", nil, value.Nil, nil},

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
		nil,
	},
}

func TestCalc(t *testing.T) {
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			m := memory.New()
			cs := []bytecode.Type{}
			ds := []value.Type{}
			dbg := make(dbginfo.Type)
			cr := compresult.Type{CS: &cs, DS: &ds, Dbg: &dbg}
			builtin.Load(cr)
			virtM := vm.New(m, cr)

			ast, err := parser.Parse(test.input)
			if test.parseError == nil {
				if err != nil {
					t.Errorf("expected no error got %s", err.Error())
					return
				}
				var v value.Type
				var err error
				for _, stmnt := range ast {
					stmnt = stmnt.STRewrite(node.SymTbl{})
					node.ByteCode(stmnt, cr)
					v, err = virtM.Run(true)
				}

				if !test.value.StrictEq(v) || err != test.runtimeError {
					t.Errorf("expected (%v, %v) got (%v, %v)", test.value, test.runtimeError, v, err)
				}
			} else if !strings.HasPrefix(err.Error(), test.parseError.Error()) {
				t.Errorf("not the expected error: %s %s", test.parseError.Error(), err.Error())
			}
		})
	}
}

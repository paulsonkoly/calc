package main_test

import (
	"testing"

	"github.com/phaul/calc/evaluator"
	"github.com/phaul/calc/parser"
	"github.com/phaul/calc/stack"
	"github.com/phaul/calc/types/node"
	"github.com/phaul/calc/types/value"
	"github.com/stretchr/testify/assert"
)

type TestDatum struct {
	name       string
	input      string
	parseError error
	value      value.Type
}

var testData = [...]TestDatum{
	{"simple literal/integer", "1", nil, value.Int(1)},
	{"simple literal/float", "3.14", nil, value.Float(3.14)},
	{"simple literal/bool", "false", nil, value.Bool(false)},

	{"simple arithmetic/addition", "1+2", nil, value.Int(3)},
	{"simple arithmetic/coercion", "1+2.0", nil, value.Float(3)},
	{"simple arithmetic/coercion", "1.0+2", nil, value.Float(3)},

	{"arithmetics/left assoc", "1-2+1", nil, value.Int(0)},
	{"arithmetics/parenthesis", "1-(2+1)", nil, value.Int(-2)},

	{"variable/not defined", "a", nil, value.Error("a not defined")},
	{"variable/lookup", "{\na=3\na+1\n}", nil, value.Int(4)},

	{"relop/int==int true", "1==1", nil, value.Bool(true)},
	{"relop/int==float true", "1==1.0", nil, value.Bool(true)},
	{"relop/float==int true", "1.0==1", nil, value.Bool(true)},
	{"relop/float==float true", "1.0==1.0", nil, value.Bool(true)},
	{"relop/bool==bool true", "false==false", nil, value.Bool(true)},

	{"relop/int!=int false", "1!=1", nil, value.Bool(false)},
	{"relop/int!=float false", "1!=1.0", nil, value.Bool(false)},
	{"relop/float!=int false", "1.0!=1", nil, value.Bool(false)},
	{"relop/float!=float false", "1.0!=1.0", nil, value.Bool(false)},
	{"relop/bool!=bool false", "false!=false", nil, value.Bool(false)},

	{"relop/float accuracy", "1==0.9999999", nil, value.Bool(false)},

	{"relop/int<int false", "1<1", nil, value.Bool(false)},
	{"relop/int<float false", "1<1.0", nil, value.Bool(false)},
	{"relop/float<int false", "1.0<1", nil, value.Bool(false)},
	{"relop/float<float false", "1.0<1.0", nil, value.Bool(false)},
	{"relop/bool<bool", "false<false", nil, value.InvalidOpError},

	{"relop/int<=int true", "1<=1", nil, value.Bool(true)},
	{"relop/int<=float true", "1<=1.0", nil, value.Bool(true)},
	{"relop/float<=int true", "1.0<=1", nil, value.Bool(true)},
	{"relop/float<=float true", "1.0<=1.0", nil, value.Bool(true)},
	{"relop/bool<=bool", "true<=true", nil, value.InvalidOpError},

	{"logicop/bool&bool true", "true&true", nil, value.Bool(true)},
	{"logicop/bool&bool false", "true&false", nil, value.Bool(false)},
	{"logicop/bool|bool true", "true|false", nil, value.Bool(true)},
	{"logicop/bool|bool false", "false|false", nil, value.Bool(false)},
	{"logicop/bool|int", "false|1", nil, value.TypeError},

	{"block/single line", "{\n1\n}", nil, value.Int(1)},
	{"block/multi line", "{\n1\n2\n}", nil, value.Int(2)},

	{"conditional/single line no else", "if true 1", nil, value.Int(1)},
	{"conditional/single line else", "if false 1 else 2", nil, value.Int(2)},
	{"conditional/incorrect condition", "if 1 1", nil, value.TypeError},
	{"conditional/no result", "if false 1", nil, value.NoResultError},
	{"conditional/blocks no else", "if true {\n1\n}", nil, value.Int(1)},
	{"conditional/blocks with else", "if false {\n1\n} else {\n2\n}", nil, value.Int(2)},

	{"loop/single line",
		`{
	a = 1
	while a < 10 a = a + 1
	a
}`, nil, value.Int(10)},
	{"loop/block",
		`{
	a = 1
	while a < 10 {
		a = a + 1
	}
	a
}`, nil, value.Int(10)},
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

	{"function definition", "(n) -> 1", nil, value.Function{Node: &node.Function{}}},
	{"function/no argument", "() -> 1", nil, value.Function{Node: &node.Function{}}},
	{"function/block",
		`(n) -> {
		n + 1
  }`, nil, value.Function{Node: &node.Function{}}},

	{"call",
		`{
			a = (n) -> 1
			a(2)
		}`, nil, value.Int(1),
	},
	{"call/no argument",
		`{
			a = () -> 1
			a()
		}`, nil, value.Int(1),
	},
	{"function/return",
		`{
			a = (n) -> {
        return 1
        2
      }
			a(2)
		}`, nil, value.Int(1),
	},
	{"function/closure",
		`{
			f = (a) -> {
        (b) -> a + b
      }
			x = f(1)
      x(2)
		}`, nil, value.Int(3),
	},
}

func TestCalc(t *testing.T) {
	for _, test := range testData {
		s := stack.NewStack()
		ast, err := parser.Parse(test.input)
		if test.parseError == nil {
			assert.NoError(t, err, test.name)
			var v value.Type
			for _, stmnt := range ast {
				v = evaluator.Evaluate(s, stmnt)
			}
			if f, ok := test.value.(value.Function); ok {
				assert.IsType(t, f, v, "test.name")
			} else {
				assert.Equal(t, test.value, v, test.name)
			}
		} else {
			assert.EqualError(t, err, test.parseError.Error(), test.name)
		}
	}
}

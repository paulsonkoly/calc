package main_test

import (
	// "errors"
	"errors"
	"testing"

	"github.com/phaul/calc/evaluator"
	"github.com/phaul/calc/parser"
	"github.com/phaul/calc/stack"
	"github.com/phaul/calc/types"
	"github.com/stretchr/testify/assert"
)

type TestDatum struct {
	name       string
	input      string
	parseError error
	value      types.Value
}

var testData = [...]TestDatum{
	{"simple literal/integer", "1", nil, types.ValueInt(1)},
	{"simple literal/float", "3.14", nil, types.ValueFloat(3.14)},
	{"simple literal/bool", "false", nil, types.ValueBool(false)},

	{"simple arithmetic/addition", "1+2", nil, types.ValueInt(3)},
	{"simple arithmetic/coercion", "1+2.0", nil, types.ValueFloat(3)},
	{"simple arithmetic/coercion", "1.0+2", nil, types.ValueFloat(3)},

	{"arithmetics/left assoc", "1-2+1", nil, types.ValueInt(0)},
	{"arithmetics/parenthesis", "1-(2+1)", nil, types.ValueInt(-2)},

	{"variable/not defined", "a", nil, types.ValueError("a not defined")},
	{"variable/lookup", "{\na=3\na+1\n}", nil, types.ValueInt(4)},

	{"relop/int==int true", "1==1", nil, types.ValueBool(true)},
	{"relop/int==float true", "1==1.0", nil, types.ValueBool(true)},
	{"relop/float==int true", "1.0==1", nil, types.ValueBool(true)},
	{"relop/float==float true", "1.0==1.0", nil, types.ValueBool(true)},
	{"relop/bool==bool true", "false==false", nil, types.ValueBool(true)},

	{"relop/int!=int false", "1!=1", nil, types.ValueBool(false)},
	{"relop/int!=float false", "1!=1.0", nil, types.ValueBool(false)},
	{"relop/float!=int false", "1.0!=1", nil, types.ValueBool(false)},
	{"relop/float!=float false", "1.0!=1.0", nil, types.ValueBool(false)},
	{"relop/bool!=bool false", "false!=false", nil, types.ValueBool(false)},

	{"relop/float accuracy", "1==0.9999999", nil, types.ValueBool(false)},

	{"relop/int<int false", "1<1", nil, types.ValueBool(false)},
	{"relop/int<float false", "1<1.0", nil, types.ValueBool(false)},
	{"relop/float<int false", "1.0<1", nil, types.ValueBool(false)},
	{"relop/float<float false", "1.0<1.0", nil, types.ValueBool(false)},
	{"relop/bool<bool", "false<false", nil, types.InvalidOpError},

	{"relop/int<=int true", "1<=1", nil, types.ValueBool(true)},
	{"relop/int<=float true", "1<=1.0", nil, types.ValueBool(true)},
	{"relop/float<=int true", "1.0<=1", nil, types.ValueBool(true)},
	{"relop/float<=float true", "1.0<=1.0", nil, types.ValueBool(true)},
	{"relop/bool<=bool", "true<=true", nil, types.InvalidOpError},

	{"logicop/bool&bool true", "true&true", nil, types.ValueBool(true)},
	{"logicop/bool&bool false", "true&false", nil, types.ValueBool(false)},
	{"logicop/bool|bool true", "true|false", nil, types.ValueBool(true)},
	{"logicop/bool|bool false", "false|false", nil, types.ValueBool(false)},
	{"logicop/bool|int", "false|1", nil, types.TypeError},

	{"block/single line", "{\n1\n}", nil, types.ValueInt(1)},
	{"block/multi line", "{\n1\n2\n}", nil, types.ValueInt(2)},

	{"conditional/single line no else", "if true 1", nil, types.ValueInt(1)},
	{"conditional/single line else", "if false 1 else 2", nil, types.ValueInt(2)},
	{"conditional/incorrect condition", "if 1 1", nil, types.TypeError},
	{"conditional/no result", "if false 1", nil, types.NoResultError},
	{"conditional/blocks no else", "if true {\n1\n}", nil, types.ValueInt(1)},
	{"conditional/blocks with else", "if false {\n1\n} else {\n2\n}", nil, types.ValueInt(2)},

	{"loop/single line",
		`{
	a = 1
	while a < 10 a = a + 1
	a
}`, nil, types.ValueInt(10)},
	{"loop/block",
		`{
	a = 1
	while a < 10 {
		a = a + 1
	}
	a
}`, nil, types.ValueInt(10)},
	{"loop/false initial condition",
		`{
	while false {
		a = a + 1
	}
}`, nil, types.NoResultError},
	{"loop/incorrect condition",
		`{
	while 13 {
		a = a + 1
	}
}`, nil, types.TypeError},

	{"function definition", "(n) -> 1", nil, types.ValueFunction(types.Node{})},
	{"function/no argument", "() -> 1", errors.New("Parser: ( expected, got )"), nil},
	{"function/block",
		`(n) -> {
		n + 1
}`, nil, types.ValueFunction(types.Node{})},

	{"call",
		`{
			a = (n) -> 1
			a(2)
		}`, nil, types.ValueInt(1),
	},
}

func TestCalc(t *testing.T) {
	for _, test := range testData {
		s := stack.NewStack()
		ast, err := parser.Parse(test.input)
		if test.parseError == nil {
			assert.NoError(t, err, test.name)
			var v types.Value
			for _, stmnt := range ast {
				v = evaluator.Evaluate(s, stmnt)
			}
			if f, ok := test.value.(types.ValueFunction); ok {
				assert.IsType(t, f, v, "test.name")
			} else {
				assert.Equal(t, test.value, v, test.name)
			}
		} else {
			assert.EqualError(t, err, test.parseError.Error(), test.name)
		}
	}
}

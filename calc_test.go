package main_test

import (
	"testing"

	"github.com/phaul/calc/evaluator"
	"github.com/phaul/calc/parser"
	"github.com/stretchr/testify/assert"
)

type TestDatum struct {
	name       string
	input      string
	parseError error
	value      evaluator.Value
}

var testData = [...]TestDatum{
	{"simple literal/integer", "1", nil, evaluator.ValueInt(1)},
	{"simple literal/float", "3.14", nil, evaluator.ValueFloat(3.14)},
	{"simple literal/bool", "false", nil, evaluator.ValueBool(false)},

	{"simple arithmetic/addition", "1+2", nil, evaluator.ValueInt(3)},
	{"simple arithmetic/coercion", "1+2.0", nil, evaluator.ValueFloat(3)},
	{"simple arithmetic/coercion", "1.0+2", nil, evaluator.ValueFloat(3)},

	{"arithmetics/left assoc", "1-2+1", nil, evaluator.ValueInt(0)},
	{"arithmetics/parenthesis", "1-(2+1)", nil, evaluator.ValueInt(-2)},

	{"variable/not defined", "a", nil, evaluator.ValueError("variable a not defined")},
	{"variable/lookup", "{\na=3\na+1\n}", nil, evaluator.ValueInt(4)},

	{"relop/int==int true", "1==1", nil, evaluator.ValueBool(true)},
	{"relop/int==float true", "1==1.0", nil, evaluator.ValueBool(true)},
	{"relop/float==int true", "1.0==1", nil, evaluator.ValueBool(true)},
	{"relop/float==float true", "1.0==1.0", nil, evaluator.ValueBool(true)},
	{"relop/bool==bool true", "false==false", nil, evaluator.ValueBool(true)},

	{"relop/int!=int false", "1!=1", nil, evaluator.ValueBool(false)},
	{"relop/int!=float false", "1!=1.0", nil, evaluator.ValueBool(false)},
	{"relop/float!=int false", "1.0!=1", nil, evaluator.ValueBool(false)},
	{"relop/float!=float false", "1.0!=1.0", nil, evaluator.ValueBool(false)},
	{"relop/bool!=bool false", "false!=false", nil, evaluator.ValueBool(false)},

	{"relop/float accuracy", "1==0.9999999", nil, evaluator.ValueBool(false)},

	{"relop/int<int false", "1<1", nil, evaluator.ValueBool(false)},
	{"relop/int<float false", "1<1.0", nil, evaluator.ValueBool(false)},
	{"relop/float<int false", "1.0<1", nil, evaluator.ValueBool(false)},
	{"relop/float<float false", "1.0<1.0", nil, evaluator.ValueBool(false)},
	{"relop/bool<bool", "false<false", nil, evaluator.InvalidOpError},

	{"relop/int<=int true", "1<=1", nil, evaluator.ValueBool(true)},
	{"relop/int<=float true", "1<=1.0", nil, evaluator.ValueBool(true)},
	{"relop/float<=int true", "1.0<=1", nil, evaluator.ValueBool(true)},
	{"relop/float<=float true", "1.0<=1.0", nil, evaluator.ValueBool(true)},
	{"relop/bool<=bool", "true<=true", nil, evaluator.InvalidOpError},

	{"block/single line", "{\n1\n}", nil, evaluator.ValueInt(1)},
	{"block/multi line", "{\n1\n2\n}", nil, evaluator.ValueInt(2)},

	{"conditional/single line no else", "if true 1", nil, evaluator.ValueInt(1)},
	{"conditional/single line else", "if false 1 else 2", nil, evaluator.ValueInt(2)},
	{"conditional/incorrect condition", "if 1 1", nil, evaluator.TypeError},
	{"conditional/no result", "if false 1", nil, evaluator.NoResultError},
	{"conditional/blocks no else", "if true {\n1\n}", nil, evaluator.ValueInt(1)},
	{"conditional/blocks with else", "if false {\n1\n} else {\n2\n}", nil, evaluator.ValueInt(2)},
}

func TestCalc(t *testing.T) {
	for _, test := range testData {
		vars := make(evaluator.Variables)
		ast, err := parser.Parse(test.input)
		if test.parseError == nil {
			assert.NoError(t, err, test.name)
			var v evaluator.Value
			for _, stmnt := range ast {
				v = evaluator.Evaluate(vars, stmnt)
			}
			assert.Equal(t, test.value, v, test.name)
		} else {
			assert.EqualError(t, err, test.parseError.Error(), test.name)
		}
	}
}

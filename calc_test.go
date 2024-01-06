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
	{"variable/lookup", "a=3\na+1", nil, evaluator.ValueInt(4)},
}

func TestCalc(t *testing.T) {
	for _, test := range testData {
		vars := make(evaluator.Variables)
		ast, err := parser.Parse(test.input)
		if test.parseError == nil {
			assert.NoError(t, err)
      var v evaluator.Value
      for _, stmnt := range(ast) {
        v = evaluator.Evaluate(vars, stmnt)  
      }
			assert.Equal(t, v, test.value)
		} else {
			assert.EqualError(t, err, test.parseError.Error())
		}
	}
}

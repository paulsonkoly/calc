package lexer_test

import (
	"testing"

	"github.com/phaul/calc/lexer"
)

type testDatum struct {
	title string
	input string
	dat   []lexer.Token
}

var testData = []testDatum{
	{"single lexeme", "13", []lexer.Token{{Value: "13", Type: lexer.IntLit}}},
	{"single lexeme", "a", []lexer.Token{{Value: "a", Type: lexer.VarName}}},
	{"single lexeme", "ab", []lexer.Token{{Value: "ab", Type: lexer.VarName}}},
	{"single lexeme", "13.6", []lexer.Token{{Value: "13.6", Type: lexer.FloatLit}}},
	{"single lexeme", "+", []lexer.Token{{Value: "+", Type: lexer.Op}}},
	{"single lexeme", "-", []lexer.Token{{Value: "-", Type: lexer.Op}}},
	{"single lexeme", "*", []lexer.Token{{Value: "*", Type: lexer.Op}}},
	{"single lexeme", "/", []lexer.Token{{Value: "/", Type: lexer.Op}}},
	{"single lexeme", "(", []lexer.Token{{Value: "(", Type: lexer.Paren}}},
	{"single lexeme", ")", []lexer.Token{{Value: ")", Type: lexer.Paren}}},
	{"whitespace at front", "   )", []lexer.Token{{Value: ")", Type: lexer.Paren}}},
	{"whitespace at back", ")    ", []lexer.Token{{Value: ")", Type: lexer.Paren}}},
	{"complex example",
		"13.6+a-(3 / 9)",
		[]lexer.Token{
			{"13.6", lexer.FloatLit},
			{"+", lexer.Op},
			{"a", lexer.VarName},
			{"-", lexer.Op},
			{"(", lexer.Paren},
			{"3", lexer.IntLit},
			{"/", lexer.Op},
			{"9", lexer.IntLit},
			{")", lexer.Paren},
		},
	},
	{"assignment",
		"a=2+3",
		[]lexer.Token{
			{"a", lexer.VarName},
			{"=", lexer.Assign},
			{"2", lexer.IntLit},
			{"+", lexer.Op},
			{"3", lexer.IntLit},
		},
	},
}

func TestLexer(t *testing.T) {
	for _, test := range testData {
		l := lexer.NewLexer(test.input)
		i := 0
		for l.Next() {
			if i >= len(test.dat) {
				t.Fatalf("%s/%s Next returns true when out of lexemes", test.title, test.input)
			}
			if l.Token != test.dat[i] {
				t.Fatalf("%s/%s returns unexpected token %v (expecting %v)", test.title, test.input, l.Token, test.dat[i])
			}
			i++
		}
		if l.Err != nil {
			t.Fatalf("%s/%s caused %s", test.title, test.input, l.Err)
		}
		if i != len(test.dat) {
			t.Fatalf("%s/%s doesn't consume all input", test.title, test.input)
		}
	}
}

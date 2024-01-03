package lexer_test

import (
	"testing"

	"github.com/phaul/calc/lexer"
	t "github.com/phaul/calc/types"
)

type testDatum struct {
	title string
	input string
	dat   []t.Token
}

var testData = []testDatum{
	{"single lexeme", "13", []t.Token{{Value: "13", Type: t.IntLit}}},
	{"single lexeme", "a", []t.Token{{Value: "a", Type: t.VarName}}},
	{"single lexeme", "ab", []t.Token{{Value: "ab", Type: t.VarName}}},
	{"single lexeme", "13.6", []t.Token{{Value: "13.6", Type: t.FloatLit}}},
	{"single lexeme", "+", []t.Token{{Value: "+", Type: t.Op}}},
	{"single lexeme", "-", []t.Token{{Value: "-", Type: t.Op}}},
	{"single lexeme", "*", []t.Token{{Value: "*", Type: t.Op}}},
	{"single lexeme", "/", []t.Token{{Value: "/", Type: t.Op}}},
	{"single lexeme", "(", []t.Token{{Value: "(", Type: t.Paren}}},
	{"single lexeme", ")", []t.Token{{Value: ")", Type: t.Paren}}},
	{"whitespace at front", "   )", []t.Token{{Value: ")", Type: t.Paren}}},
	{"whitespace at back", ")    ", []t.Token{{Value: ")", Type: t.Paren}}},
	{"complex example",
		"13.6+a-(3 / 9)",
		[]t.Token{
			{Value: "13.6", Type: t.FloatLit},
			{Value: "+", Type: t.Op},
			{Value: "a", Type: t.VarName},
			{Value: "-", Type: t.Op},
			{Value: "(", Type: t.Paren},
			{Value: "3", Type: t.IntLit},
			{Value: "/", Type: t.Op},
			{Value: "9", Type: t.IntLit},
			{Value: ")", Type: t.Paren},
		},
	},
	{"assignment",
		"a=2+3",
		[]t.Token{
			{Value: "a", Type: t.VarName},
			{Value: "=", Type: t.Assign},
			{Value: "2", Type: t.IntLit},
			{Value: "+", Type: t.Op},
			{Value: "3", Type: t.IntLit},
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

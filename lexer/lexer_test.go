package lexer_test

import (
	"testing"

	"github.com/phaul/calc/lexer"
	t "github.com/phaul/calc/types"
	"github.com/stretchr/testify/assert"
)

type testDatum struct {
	title string
	input string
	dat   []t.Token
}

var eol = t.Token{Value: "\n", Type: t.EOL}

var testData = []testDatum{
	{"single lexeme", "13", []t.Token{{Value: "13", Type: t.IntLit}, eol}},
	{"single lexeme", "a", []t.Token{{Value: "a", Type: t.VarName}, eol}},
	{"single lexeme", "ab", []t.Token{{Value: "ab", Type: t.VarName}, eol}},
	{"single lexeme", "13.6", []t.Token{{Value: "13.6", Type: t.FloatLit}, eol}},
	{"single lexeme", "+", []t.Token{{Value: "+", Type: t.SingleChar}, eol}},
	{"single lexeme", "-", []t.Token{{Value: "-", Type: t.SingleChar}, eol}},
	{"single lexeme", "*", []t.Token{{Value: "*", Type: t.SingleChar}, eol}},
	{"single lexeme", "/", []t.Token{{Value: "/", Type: t.SingleChar}, eol}},
	{"single lexeme", "(", []t.Token{{Value: "(", Type: t.SingleChar}, eol}},
	{"single lexeme", ")", []t.Token{{Value: ")", Type: t.SingleChar}, eol}},
	{"new line lexeme", "a\nb", []t.Token{{Value: "a", Type: t.VarName}, eol, {Value: "b", Type: t.VarName}, eol}},
	{"whitespace at front", "   )", []t.Token{{Value: ")", Type: t.SingleChar}, eol}},
	{"whitespace at back", ")    ", []t.Token{{Value: ")", Type: t.SingleChar}, eol}},
	{"complex example",
		"13.6+a-(3 / 9)\n",
		[]t.Token{
			{Value: "13.6", Type: t.FloatLit},
			{Value: "+", Type: t.SingleChar},
			{Value: "a", Type: t.VarName},
			{Value: "-", Type: t.SingleChar},
			{Value: "(", Type: t.SingleChar},
			{Value: "3", Type: t.IntLit},
			{Value: "/", Type: t.SingleChar},
			{Value: "9", Type: t.IntLit},
			{Value: ")", Type: t.SingleChar},
			eol, eol,
		},
	},
	{"assignment",
		"a=2+3",
		[]t.Token{
			{Value: "a", Type: t.VarName},
			{Value: "=", Type: t.SingleChar},
			{Value: "2", Type: t.IntLit},
			{Value: "+", Type: t.SingleChar},
			{Value: "3", Type: t.IntLit},
			eol,
		},
	},
}

func TestLexer(t *testing.T) {
	for _, test := range testData {
		l := lexer.NewLexer(test.input)
		i := 0
		for l.Next() {
			assert.Less(t, i, len(test.dat), "%s/%s Next returns true when out of lexemes", test.title, test.input)
			if i < len(test.dat) {
				assert.Equal(t,
					test.dat[i],
					l.Token,
					"%s/%s returns unexpected token %v (expecting %v)",
					test.title, test.input, l.Token, test.dat[i])
			}
			i++
		}
		assert.NoError(t, l.Err, "%s/%s caused %s", test.title, test.input, l.Err)
		assert.Equal(t, len(test.dat), i, "%s/%s doesn't consume all input", test.title, test.input)
	}
}

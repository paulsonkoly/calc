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
var eof = t.Token{Value: string(lexer.EOF), Type: t.EOF}

var testData = []testDatum{
	{"single lexeme", "13", []t.Token{{Value: "13", Type: t.IntLit}, eof}},
	{"single lexeme", "a", []t.Token{{Value: "a", Type: t.Name}, eof}},
	{"single lexeme", "ab", []t.Token{{Value: "ab", Type: t.Name}, eof}},
	{"single lexeme", "13.6", []t.Token{{Value: "13.6", Type: t.FloatLit}, eof}},
	{"single lexeme", "+", []t.Token{{Value: "+", Type: t.Sticky}, eof}},
	{"single lexeme", "-", []t.Token{{Value: "-", Type: t.Sticky}, eof}},
	{"single lexeme", "*", []t.Token{{Value: "*", Type: t.Sticky}, eof}},
	{"single lexeme", "/", []t.Token{{Value: "/", Type: t.Sticky}, eof}},
	{"single lexeme", "(", []t.Token{{Value: "(", Type: t.NotSticky}, eof}},
	{"single lexeme", ")", []t.Token{{Value: ")", Type: t.NotSticky}, eof}},
	{"sticky double", "<=", []t.Token{{Value: "<=", Type: t.Sticky}, eof}},
	{"non-sticky double", "((", []t.Token{{Value: "(", Type: t.NotSticky}, {Value: "(", Type: t.NotSticky}, eof}},
	{"new line lexeme", "a\nb", []t.Token{{Value: "a", Type: t.Name}, eol, {Value: "b", Type: t.Name}, eof}},
	{"whitespace at front", "   )", []t.Token{{Value: ")", Type: t.NotSticky}, eof}},
	{"whitespace at back", ")    ", []t.Token{{Value: ")", Type: t.NotSticky}, eof}},
	{"complex example",
		"13.6+a-(3 / 9)\n",
		[]t.Token{
			{Value: "13.6", Type: t.FloatLit},
			{Value: "+", Type: t.Sticky},
			{Value: "a", Type: t.Name},
			{Value: "-", Type: t.Sticky},
			{Value: "(", Type: t.NotSticky},
			{Value: "3", Type: t.IntLit},
			{Value: "/", Type: t.Sticky},
			{Value: "9", Type: t.IntLit},
			{Value: ")", Type: t.NotSticky},
			eol, eof,
		},
	},
	{"assignment",
		"a=2+3",
		[]t.Token{
			{Value: "a", Type: t.Name},
			{Value: "=", Type: t.Sticky},
			{Value: "2", Type: t.IntLit},
			{Value: "+", Type: t.Sticky},
			{Value: "3", Type: t.IntLit},
			eof,
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

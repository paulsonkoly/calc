package lexer_test

import (
	"testing"

	"github.com/paulsonkoly/calc/lexer"
	"github.com/paulsonkoly/calc/types/token"
	"github.com/stretchr/testify/assert"
)

type testDatum struct {
	title string
	input string
	dat   []token.Type
}

var eol = token.Type{Value: "\n", Type: token.EOL}
var eof = token.Type{Value: string(lexer.EOF), Type: token.EOF}

var testData = []testDatum{
	{"single lexeme", "13", []token.Type{{Value: "13", Type: token.IntLit}, eol, eof}},
	{"single lexeme", "a", []token.Type{{Value: "a", Type: token.Name}, eol, eof}},
	{"single lexeme", "ab", []token.Type{{Value: "ab", Type: token.Name}, eol, eof}},
	{"single lexeme", "13.6", []token.Type{{Value: "13.6", Type: token.FloatLit}, eol, eof}},
	{"single lexeme", "+", []token.Type{{Value: "+", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "-", []token.Type{{Value: "-", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "*", []token.Type{{Value: "*", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "/", []token.Type{{Value: "/", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "#", []token.Type{{Value: "#", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "%", []token.Type{{Value: "%", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "(", []token.Type{{Value: "(", Type: token.NotSticky}, eol, eof}},
	{"single lexeme", ")", []token.Type{{Value: ")", Type: token.NotSticky}, eol, eof}},
	{"single lexeme", "{", []token.Type{{Value: "{", Type: token.NotSticky}, eol, eof}},
	{"single lexeme", "}", []token.Type{{Value: "}", Type: token.NotSticky}, eol, eof}},
	{"single lexeme", "[", []token.Type{{Value: "[", Type: token.NotSticky}, eol, eof}},
	{"single lexeme", "]", []token.Type{{Value: "]", Type: token.NotSticky}, eol, eof}},
	{"single lexeme", "&", []token.Type{{Value: "&", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "|", []token.Type{{Value: "|", Type: token.Sticky}, eol, eof}},
	{"single lexeme", "~", []token.Type{{Value: "~", Type: token.Sticky}, eol, eof}},
	{"string literal", "\"abc\"", []token.Type{{Value: "\"abc\"", Type: token.StringLit}, eol, eof}},
	{"escaped string literal", "\"a\\\"bc\"", []token.Type{{Value: "\"a\\\"bc\"", Type: token.StringLit}, eol, eof}},
	{"string literal with new line", "\"a\nbc\"", []token.Type{{Value: "\"a\nbc\"", Type: token.StringLit}, eol, eof}},
	{"string literal with escaped line", "\"a\\nbc\"", []token.Type{{Value: "\"a\nbc\"", Type: token.StringLit}, eol, eof}},
	{"sticky double", "<=", []token.Type{{Value: "<=", Type: token.Sticky}, eol, eof}},
	{"non-sticky double", "((", []token.Type{{Value: "(", Type: token.NotSticky}, {Value: "(", Type: token.NotSticky}, eol, eof}},
	{"new line lexeme", "a\nb", []token.Type{{Value: "a", Type: token.Name}, eol, {Value: "b", Type: token.Name}, eol, eof}},
	{"whitespace at front", "   )", []token.Type{{Value: ")", Type: token.NotSticky}, eol, eof}},
	{"whitespace at back", ")    ", []token.Type{{Value: ")", Type: token.NotSticky}, eol, eof}},
	{"complex example",
		"13.6+a-(3 / 9)\n",
		[]token.Type{
			{Value: "13.6", Type: token.FloatLit},
			{Value: "+", Type: token.Sticky},
			{Value: "a", Type: token.Name},
			{Value: "-", Type: token.Sticky},
			{Value: "(", Type: token.NotSticky},
			{Value: "3", Type: token.IntLit},
			{Value: "/", Type: token.Sticky},
			{Value: "9", Type: token.IntLit},
			{Value: ")", Type: token.NotSticky},
			eol, eof,
		},
	},
	{"assignment",
		"a=2+3",
		[]token.Type{
			{Value: "a", Type: token.Name},
			{Value: "=", Type: token.Sticky},
			{Value: "2", Type: token.IntLit},
			{Value: "+", Type: token.Sticky},
			{Value: "3", Type: token.IntLit},
			eol, eof,
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

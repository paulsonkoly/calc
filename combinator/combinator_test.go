package combinator_test

import (
	"strings"
	"testing"

	"github.com/paulsonkoly/calc/combinator"
	"github.com/stretchr/testify/assert"
)

type testDatum struct {
	name      string
	parser    combinator.Parser
	lexerOut  []testToken
	parserOut []testNode
	err       string
}

type tokenWrapper struct{}

func (tokenWrapper) Wrap(t combinator.Token) combinator.Node {
	return testNode{token: t.(testToken)}
}

func accept(t string) combinator.Parser {
	tokenWrap := tokenWrapper{}
	return combinator.Accept(func(a combinator.Token) bool { return string(a.(testToken)) == t }, "?", tokenWrap)
}

var testData = []testDatum{
	{
		name:      "Accept",
		parser:    accept("a"),
		lexerOut:  []testToken{"a"},
		parserOut: []testNode{{token: testToken("a")}},
		err:       "",
	},
	{
		name:      "AcceptFail",
		parser:    accept("b"),
		lexerOut:  []testToken{"a"},
		parserOut: nil,
		err:       "Parser: ? expected, got a",
	},
	{
		name:      "And",
		parser:    combinator.And(accept("a"), accept("b")),
		lexerOut:  []testToken{"a", "b"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("b")}},
		err:       "",
	},
	{
		name:      "Seq",
		parser:    combinator.Seq(accept("a"), accept("b"), accept("c")),
		lexerOut:  []testToken{"a", "b", "c"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("b")}, {token: testToken("c")}},
		err:       "",
	},
	{
		name:      "OneOf",
		parser:    combinator.OneOf(accept("a"), accept("b")),
		lexerOut:  []testToken{"b"},
		parserOut: []testNode{{token: testToken("b")}},
		err:       "",
	},
	{
		name:      "OneOf failed",
		parser:    combinator.OneOf(accept("a"), accept("b")),
		lexerOut:  []testToken{"c"},
		parserOut: nil,
		err:       "Parser: ? expected, got c",
	},
	{
		name: "Backtrack aab -> a(aa|ab)",
		parser: combinator.And(
			accept("a"),
			combinator.OneOf(combinator.And(accept("a"), accept("a")), combinator.And(accept("a"), accept("b"))),
		),
		lexerOut:  []testToken{"a", "a", "b"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("a")}, {token: testToken("b")}},
		err:       "",
	},
	{
		name:      "OneOf",
		parser:    combinator.OneOf(accept("a"), accept("b"), accept("c")),
		lexerOut:  []testToken{"b"},
		parserOut: []testNode{{token: testToken("b")}},
		err:       "",
	},
	{
		name:      "Choose",
		parser:    combinator.Choose(combinator.Conditional{Gate: accept("a"), OnSuccess: accept("b")}),
		lexerOut:  []testToken{"a", "b"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("b")}},
		err:       "",
	},
	{
		name:      "Any (none)",
		parser:    combinator.Any(combinator.Conditional{Gate: accept("a"), OnSuccess: combinator.Ok()}),
		lexerOut:  []testToken{"b", "b", "b", "b"},
		parserOut: []testNode{},
		err:       "",
	},
	{
		name:      "Any (some)",
		parser:    combinator.Any(combinator.Conditional{Gate: accept("a"), OnSuccess: combinator.Ok()}),
		lexerOut:  []testToken{"a", "a", "b", "b"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("a")}},
		err:       "",
	},
	{
		name:      "Separated by (nothing)",
		parser:    combinator.SeparatedBy(accept("a"), accept("b")),
		lexerOut:  []testToken{"c"},
		parserOut: []testNode{},
		err:       "",
	},
	{
		name:      "Separated by (single token)",
		parser:    combinator.SeparatedBy(accept("a"), accept("b")),
		lexerOut:  []testToken{"a"},
		parserOut: []testNode{{token: testToken("a")}},
		err:       "",
	},
	{
		name:      "Separated by",
		parser:    combinator.SeparatedBy(accept("a"), accept("b")),
		lexerOut:  []testToken{"a", "b", "a", "b", "a"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("a")}, {token: testToken("a")}},
		err:       "",
	},
	{
		name:      "Separated by/doesnt consume last spearator",
		parser:    combinator.And(combinator.SeparatedBy(accept("a"), accept("b")), combinator.And(accept("b"), accept("c"))),
		lexerOut:  []testToken{"a", "b", "c"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("b")}, {token: testToken("c")}},
		err:       "",
	},
	{
		name:      "Surrounded by",
		parser:    combinator.SurroundedBy(accept("a"), accept("b"), accept("c")),
		lexerOut:  []testToken{"a", "b", "c"},
		parserOut: []testNode{{token: testToken("b")}},
		err:       "",
	},
	{
		name: "Fmap",
		parser: combinator.Fmap(
			func(i []combinator.Node) []combinator.Node {
				// sigh
				return []combinator.Node{
					combinator.Node(testNode{token: testToken(strings.Repeat(string(i[0].(testNode).token), 2))}),
				}
			},
			accept("a"),
		),
		lexerOut:  []testToken{"a"},
		parserOut: []testNode{{token: testToken("aa")}},
		err:       "",
	},
}

type testToken string

func (t testToken) From() int { return 0 }
func (t testToken) To() int   { return 0 }

type testNode struct{ token testToken }

func (t testToken) Node() combinator.Node { return testNode{token: t} }

type lexerStub struct {
	tokens   []testToken
	readP    int
	pointers []int
}

func (l lexerStub) From() int               { return 0 }
func (l lexerStub) To() int                 { return 0 }
func (l lexerStub) Token() combinator.Token { return combinator.Token(l.tokens[l.readP]) }
func (l lexerStub) Err() error              { return nil }
func (l *lexerStub) Next() bool             { l.readP++; return l.readP < len(l.tokens) }
func (l *lexerStub) Snapshot()              { l.pointers = append(l.pointers, l.readP) }
func (l *lexerStub) Commit()                { l.pointers = l.pointers[:len(l.pointers)-1] }
func (l *lexerStub) Rollback() {
	l.readP = l.pointers[len(l.pointers)-1]
}

func TestCombinator(t *testing.T) {
	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			l := lexerStub{tokens: tt.lexerOut, readP: -1, pointers: make([]int, 0)}
			p := tt.parser

			n, err := p(combinator.RollbackLexer(&l))
			if tt.err != "" {
				if assert.Error(t, err) {
					assert.Equal(t, tt.err, err.Error())
				}
			} else {
				assert.Nil(t, err)

				convert := []testNode{}
				for _, a := range n {
					convert = append(convert, a.(testNode))
				}
				assert.Equal(t, tt.parserOut, convert)
			}
		})
	}
}

package combinator_test

import (
	"strings"
	"testing"

	"github.com/phaul/calc/combinator"
	"github.com/stretchr/testify/assert"
)

type testDatum struct {
	name      string
	parser    combinator.Parser
	lexerOut  []testToken
	parserOut []testNode
	err       string
}

func accept(t string) combinator.Parser {
	return combinator.Accept(func(a combinator.Token) bool { return string(a.(testToken)) == t })
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
		err:       "Parser: a failed",
	},
	{
		name:      "And",
		parser:    combinator.And(accept("a"), accept("b")),
		lexerOut:  []testToken{"a", "b"},
		parserOut: []testNode{{token: testToken("a")}, {token: testToken("b")}},
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

type testNode struct{ token testToken }

func (t testToken) Node() combinator.Node { return testNode{token: t} }

type lexerStub struct {
	tokens   []testToken
	readP    int
	pointers []int
}

func (l lexerStub) Token() combinator.Token { return combinator.Token(l.tokens[l.readP]) }
func (l lexerStub) Err() error              { return nil }
func (l *lexerStub) Next() bool             { l.readP++; return l.readP < len(l.tokens) }
func (l *lexerStub) Snapshot()              { l.pointers = append(l.pointers, l.readP) }
func (l *lexerStub) Rollback() {
	l.readP = l.pointers[len(l.pointers)-1]
	l.pointers = l.pointers[:len(l.pointers)-1]
}

func TestCombinator(t *testing.T) {
	for _, tt := range testData {
		l := lexerStub{tokens: tt.lexerOut, readP: -1, pointers: make([]int, 0)}
		p := tt.parser

		n, err := p(combinator.RollbackLexer(&l))
    if tt.err != "" {
      if assert.Error(t, err) {
        assert.Equal(t, tt.err, err.Error())
      }
    } else {
      assert.NoError(t, err)

			convert := []testNode{}
			for _, a := range n {
				convert = append(convert, a.(testNode))
			}
			assert.Equal(t, tt.parserOut, convert)
		}
	}
}

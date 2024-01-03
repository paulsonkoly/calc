// package lexer is a lexer for our tiny language with an iterator style
// interface
//
//	   l := NewLexer("12+13")
//
//			for l.Next() {
//	     tok := l.Token
//	     ...
//			}
//			if l.Err != nil {
//	      // some lexing issues
//			}
package lexer

import (
	"strings"

	t "github.com/phaul/calc/types"
)

type Lexer struct {
	input    string
	rdr      strings.Reader
	from, to int
	Token    t.Token // the next token
	Err      error   // if there was an error this will be set
	state    stateFunc
}

// NewLexer a new lexer with input string
func NewLexer(input string) Lexer {
	return Lexer{input: input, rdr: *strings.NewReader(input), state: whiteSpace}
}

// Next advances the lexer to a new token.

// returns false if an error happened or there are no tokens left
func (l *Lexer) Next() bool {
	var st stateFunc

	for st = l.state; !l.finished(); {
		var c rune
		var s int
		var err error

		if c, s, err = l.nextRune(); err != nil {
			l.Err = err
			return false
		}

		str := st(c)

		if str.err != nil {
			l.Err = str.err
			return false
		}

		st = str.next

		if str.doEmit {
			l.Token = t.Token{Value: l.input[l.from:l.to], Type: str.typ}
			l.state = str.next
			l.from = l.to
			l.to += s
			return true
		} else if str.doAdv {
			l.from = l.to
		}
		l.to += s
	}
	return false
}

const eof = rune(0)

func (l *Lexer) nextRune() (rune, int, error) {
	if l.to >= len(l.input) {
		return eof, 0, nil
	}

	c, s, err := l.rdr.ReadRune()
	return c, s, err
}

func (l Lexer) finished() bool {
	return l.from == len(l.input) && l.to == len(l.input)
}

// Package lexer is a lexer for our tiny language with an iterator style
// interface.
//
//	l := NewLexer("12+13")
//
//	for l.Next() {
//	 tok := l.Token
//	 ...
//	}
//	if l.Err != nil {
//	  // some lexing issues
//	}
package lexer

import (
	"strings"

	"github.com/paulsonkoly/calc/types/token"
)

type Lexer struct {
	input    string
	rdr      strings.Reader
	from, to int
	Token    token.Type // the next token
	Err      error      // if there was an error this will be set
	state    stateFunc
	eof      bool
}

// NewLexer creates a new lexer with input string.
func NewLexer(input string) Lexer {
	return Lexer{input: input, rdr: *strings.NewReader(input), state: whiteSpace}
}

// Next advances the lexer to a new token.
//
// It returns false if an error happened or there are no tokens left.
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
			// return false
			return true
		}

		st = str.next

		if str.doEmit {
			word := l.input[l.from:l.to]
			if str.typ == token.StringLit {
				word = strings.ReplaceAll(word, "\\n", "\n")
			}
			l.Token = token.WithFromTo(str.typ, word, l.from, l.to)
			l.state = str.next
			l.from = l.to
			l.to += s
			return true
		} else if str.doAdv {
			l.from = l.to
		}
		l.to += s
	}

	l.Err = nil
	switch {
	case !l.eof && l.Token.Type != token.EOL:
		l.Token = token.Type{Value: "\n", Type: token.EOL}
		return true
	case !l.eof:
		l.Token = token.Type{Value: string(EOF), Type: token.EOF}
		l.eof = true
		return true
	}
	return false
}

const EOF = rune(0)

func (l *Lexer) nextRune() (rune, int, error) {
	if l.to >= len(l.input) {
		return EOF, 0, nil
	}

	c, s, err := l.rdr.ReadRune()
	return c, s, err
}

func (l Lexer) finished() bool {
	return l.from == len(l.input) && l.to == len(l.input)
}

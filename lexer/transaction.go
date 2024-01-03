package lexer

import (
	"github.com/phaul/calc/combinator"
	t "github.com/phaul/calc/types"
)

type lexerResult struct {
	token t.Token
	err   error
}

// TLexer is a lexer satisfying combinators.RollbackLexer
type TLexer struct {
	stack    []lexerResult
	pointers []int
	writep   int
	readp    int
	nextF    bool
	lexer    Lexer
}

func NewTLexer(input string) TLexer {
	return TLexer{
		stack:    make([]lexerResult, 0),
		pointers: make([]int, 0),
    readp:    -1,
    nextF:    true,
		lexer:    NewLexer(input),
	}
}

func (tl *TLexer) Next() bool {
	if !tl.nextF && tl.readp < tl.writep {
		return true
	}
	tl.nextF = true
	if tl.lexer.Next() {
		e := lexerResult{token: tl.lexer.Token, err: tl.lexer.Err}
		tl.stack = append(tl.stack, e)
		tl.writep++
		return true
	}

	return false
}

func (tl *TLexer) Token() combinator.Token {
	if tl.nextF {
		tl.readp++
		tl.nextF = false
	}
	return combinator.Token(tl.stack[tl.readp].token)
}

func (tl *TLexer) Err() error {
	if tl.nextF {
		tl.readp++
		tl.nextF = false
	}
	return tl.stack[tl.writep].err
}

func (tl *TLexer) Snapshot() {
	tl.pointers = append(tl.pointers, tl.readp)
}

func (tl *TLexer) Rollback() {
	tl.readp = tl.pointers[len(tl.pointers)-1]
	tl.pointers = tl.pointers[:len(tl.pointers)-1]
}

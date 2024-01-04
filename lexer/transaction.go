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
//
// It's the same as Lexer, but it's state can be snap-shotted and rolled back,
// so if the parser backtracks, the parser can instruct the lexer to restore a
// previous point in lexing and start emitting tokens it has already emitted
// before.
type TLexer struct {
	stack    []lexerResult
	pointers []int
	writep   int
	readp    int
	lexer    Lexer
}

func NewTLexer(input string) TLexer {
	return TLexer{
		stack:    make([]lexerResult, 0),
		pointers: make([]int, 0),
    readp:    -1,
		lexer:    NewLexer(input),
	}
}

// Next advances the lexer to a new token.
//
// returns false if an error happened or there are no tokens left
func (tl *TLexer) Next() bool {
	if tl.readp < tl.writep - 1 {
    tl.readp++
		return true
	}
	if tl.lexer.Next() {
		e := lexerResult{token: tl.lexer.Token, err: tl.lexer.Err}
		tl.stack = append(tl.stack, e)
    tl.readp++
		tl.writep++
		return true
	}

	return false
}

// Token gives the next token
func (tl *TLexer) Token() combinator.Token {
	return combinator.Token(tl.stack[tl.readp].token)
}

// Err gives the next lexer error if any
func (tl *TLexer) Err() error {
	return tl.stack[tl.readp].err
}

// Snapshot snapshots the lexer state
func (tl *TLexer) Snapshot() {
	tl.pointers = append(tl.pointers, tl.readp)
}

// Rollback rolls back to the last snapshot
func (tl *TLexer) Rollback() {
	tl.readp = tl.pointers[len(tl.pointers)-1]
	tl.pointers = tl.pointers[:len(tl.pointers)-1]
}

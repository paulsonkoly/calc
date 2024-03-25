package lexer

import (
	"github.com/paulsonkoly/calc/combinator"
	"github.com/paulsonkoly/calc/types/token"
)

type lexerResult struct {
	token    token.Type
	err      error
	from, to int
}

// TLexer is a lexer satisfying combinators.RollbackLexer.
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
// It returns false if an error happened or there are no tokens left.
func (tl *TLexer) Next() bool {
	if tl.readp < tl.writep-1 {
		tl.readp++
		return true
	}
	if tl.lexer.Next() {
		e := lexerResult{token: tl.lexer.Token, err: tl.lexer.Err, from: tl.lexer.from, to: tl.lexer.to}
		tl.stack = append(tl.stack, e)
		tl.readp++
		tl.writep++
		return true
	}

	return false
}

// Token gives the next token.
func (tl *TLexer) Token() combinator.Token {
	return combinator.Token(tl.stack[tl.readp].token)
}

// Err gives the next lexer error if any.
func (tl *TLexer) Err() error {
	return tl.stack[tl.readp].err
}

// From gives the next token starting position.
func (tl *TLexer) From() int {
	return tl.stack[tl.readp].from
}

// To gives the next token starting position.
func (tl *TLexer) To() int {
	return tl.stack[tl.readp].to
}

// Snapshot snapshots the lexer state.
func (tl *TLexer) Snapshot() {
	tl.pointers = append(tl.pointers, tl.readp)
}

func (tl *TLexer) Commit() {
	tl.pointers = tl.pointers[:len(tl.pointers)-1]
}

// Rollback rolls back to the last snapshot.
func (tl *TLexer) Rollback() {
	tl.readp = tl.pointers[len(tl.pointers)-1]
	tl.pointers = tl.pointers[:len(tl.pointers)-1]
}

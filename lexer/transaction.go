package lexer

type lexerResult struct {
	token Token
	err   error
}

// TLexer is a lexer satisfying combinators.RollbackLexer
type TLexer struct {
	stack    []lexerResult
	pointers []int
	writep   int
	readp    int
	lexer    Lexer
}

// NewTLexer a new rollback capable lexer with input string
func NewTLexer(input string) TLexer {
	return TLexer{
		stack:    make([]lexerResult, 0),
		pointers: make([]int, 0),
    readp: -1,
		lexer:    NewLexer(input),
	}
}

// Next advances the lexer to a new token.

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
func (tl TLexer) Token() Token {
  return tl.stack[tl.readp].token
}

// Err gives the next lexer error if any
func (tl TLexer) Err() error {
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

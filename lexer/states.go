package lexer

import (
	"fmt"
	"log"
	"strings"

	"github.com/paulsonkoly/calc/types/token"
)

const (
	stickyChars     = "+*/=<>!-&|@#%"
	nonStrickyChars = "(){}[],:"
)

// stateFuncResult
type str struct {
	next   stateFunc // next state function
	doEmit bool      // lexer emits token
	doAdv  bool      // lexer advances from to to (current token becomes empty)
	typ    token.TokenType
	err    error
}

type stateFunc func(c rune) str

func newSTR(c rune, typ token.TokenType, emit, adv bool, format string, args ...any) str {
	switch {
	case c == ' ' || c == '\t':
		return str{next: whiteSpace, doEmit: emit, doAdv: adv, typ: typ}

	case c == '\n':
		return str{next: eol, doEmit: emit, doAdv: adv, typ: typ}

	case c == EOF:
		return str{next: eof, doEmit: emit, doAdv: adv, typ: typ}

	case '0' <= c && c <= '9':
		return str{next: intLit, doEmit: emit, doAdv: adv, typ: typ}

	case 'a' <= c && c <= 'z':
		return str{next: varName, doEmit: emit, doAdv: adv, typ: typ}

	case c == '"':
		return str{next: stringLit, doEmit: emit, doAdv: adv, typ: typ}

	case strings.Contains(nonStrickyChars, string(c)):
		return str{next: notSticky, doEmit: emit, doAdv: adv, typ: typ}

	case strings.Contains(stickyChars, string(c)):
		return str{next: sticky, doEmit: emit, doAdv: adv, typ: typ}

	default:
		return str{err: fmt.Errorf(format, args...)}
	}
}

func whiteSpace(c rune) str {
	return newSTR(c, token.Invalid, false, true, "Lexer: unexpected char %c", c)
}

func intLit(c rune) str {
	switch {
	case '0' <= c && c <= '9':
		return str{next: intLit}

	case c == '.':
		return str{next: floatLit}

	default:
		return newSTR(c, token.IntLit, true, false, "Lexer: unexpected char %c in integer literal", c)
	}
}

func floatLit(c rune) str {
	switch {
	case '0' <= c && c <= '9':
		return str{next: floatLit}

	default:
		return newSTR(c, token.FloatLit, true, false, "Lexer: unexpected char %c in float literal", c)
	}
}

func varName(c rune) str {
	switch {
	case 'a' <= c && c <= 'z':
		return str{next: varName}

	default:
		return newSTR(c, token.Name, true, false, "Lexer: unexpected char %c in variable name", c)
	}
}

func stringLit(c rune) str {
	switch {
	case c == '"':
		return str{next: stringLitEnd}

	case c == '\\':
		return str{next: escapeStringLit}

	default:
		return str{next: stringLit}
	}
}

func escapeStringLit(c rune) str {
	return str{next: stringLit}
}

func stringLitEnd(c rune) str {
	return newSTR(c, token.StringLit, true, false, "Lexer: unexpected char %c in string literal", c)
}

func notSticky(c rune) str {
	return newSTR(c, token.NotSticky, true, false, "Lexer: unexpected char %c", c)
}

func sticky(c rune) str {
	switch {
	case strings.Contains(stickyChars, string(c)):
		return str{next: sticky}

	default:
		return newSTR(c, token.Sticky, true, false, "Lexer: unexpected char %c following operator", c)
	}
}

func eol(c rune) str {
	return newSTR(c, token.EOL, true, false, "Lexer: unexpected char %c following new line", c)
}

func eof(c rune) str {
	log.Panicf("Lexer: %c character after end of input", c)
	return str{}
}

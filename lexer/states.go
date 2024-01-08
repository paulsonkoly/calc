package lexer

import (
	"fmt"
	"log"
	"strings"

	t "github.com/phaul/calc/types"
)

// stateFuncResult
type str struct {
	next   stateFunc // next state function
	doEmit bool      // lexer emits token
	doAdv  bool      // lexer advances from to to (current token becomes empty)
	typ    t.TokenType
	err    error
}

type stateFunc func(c rune) str

func newSTR(c rune, typ t.TokenType, emit, adv bool, format string, args ...any) str {
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

	case strings.Contains("(){},", string(c)):
		return str{next: notSticky, doEmit: emit, doAdv: adv, typ: typ}

	case strings.Contains("+*/=<>!-", string(c)):
		return str{next: sticky, doEmit: emit, doAdv: adv, typ: typ}

	default:
		return str{err: fmt.Errorf(format, args...)}
	}
}

func whiteSpace(c rune) str {
	return newSTR(c, t.InvalidToken, false, true, "Lexer: unexpected char %c", c)
}

func intLit(c rune) str {
	switch {
	case '0' <= c && c <= '9':
		return str{next: intLit}

	case c == '.':
		return str{next: floatLit}

	default:
		return newSTR(c, t.IntLit, true, false, "Lexer: unexpected char %c in integer literal", c)
	}
}

func floatLit(c rune) str {
	switch {
	case '0' <= c && c <= '9':
		return str{next: floatLit}

	default:
		return newSTR(c, t.FloatLit, true, false, "Lexer: unexpected char %c in float literal", c)
	}
}

func varName(c rune) str {
	switch {
	case 'a' <= c && c <= 'z':
		return str{next: varName}

	default:
		return newSTR(c, t.Name, true, false, "Lexer: unexpected char %c in variable name", c)
	}
}

func notSticky(c rune) str {
	return newSTR(c, t.NotSticky, true, false, "Lexer: unexpected char %c", c)
}

func sticky(c rune) str {
	switch {
	case strings.Contains("+*/=<>!-", string(c)):
		return str{next: sticky}

	default:
		return newSTR(c, t.Sticky, true, false, "Lexer: unexpected char %c following operator", c)
	}
}

func eol(c rune) str {
	return newSTR(c, t.EOL, true, false, "Lexer: unexpected char %c following new line", c)
}

func eof(c rune) str {
	log.Panicf("Lexer: %c character after end of input", c)
	return str{}
}

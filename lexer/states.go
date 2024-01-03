package lexer

import (
	"fmt"

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
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: whiteSpace, doEmit: emit, doAdv: adv, typ: typ}

	case '0' <= c && c <= '9':
		return str{next: intLit, doEmit: emit, doAdv: adv, typ: typ}

	case 'a' <= c && c <= 'z':
		return str{next: varName, doEmit: emit, doAdv: adv, typ: typ}

	case c == '(', c == ')':
		return str{next: paren, doEmit: emit, doAdv: adv, typ: typ}

	case c == '+' || c == '-' || c == '*' || c == '/':
		return str{next: op, doEmit: emit, doAdv: adv, typ: typ}

	case c == '=':
		return str{next: assign, doEmit: emit, doAdv: adv, typ: typ}

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
		return newSTR(c, t.VarName, true, false, "Lexer: unexpected char %c in variable name", c)
	}
}

func op(c rune) str {
	return newSTR(c, t.Op, true, false, "Lexer: unexpected char %c following operator", c)
}

func paren(c rune) str {
	return newSTR(c, t.Paren, true, false, "Lexer: unexpected char %c following parenthesis", c)
}

func assign(c rune) str {
	return newSTR(c, t.Assign, true, false, "Lexer: unexpected char %c following assignment", c)
}

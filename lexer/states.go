package lexer

import (
	"errors"
	"fmt"
)

// stateFuncResult
type str struct {
	next   stateFunc
	doEmit bool
	doAdv  bool
	typ    TokenType
	err    error
}

type stateFunc func(c rune) str

func start(c rune) str {
	switch {
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: start, doAdv: true}

	case '0' <= c && c <= '9':
		return str{next: intLit, doAdv: true}

	case 'a' <= c && c <= 'z':
		return str{next: varName, doAdv: true}

	case c == '(' || c == ')':
		return str{next: paren, doAdv: true}

	case c == '+' || c == '-' || c == '*' || c == '/':
		return str{next: op, doAdv: true}

	case c == '=':
		return str{next: assign, doAdv: true}

	default:
		return str{err: errors.New(fmt.Sprintf("Lexer: unexpected char %c", c))}
	}
}

func intLit(c rune) str {
	switch {
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: start, doEmit: true, typ: IntLit}

	case '0' <= c && c <= '9':
		return str{next: intLit}

	case c == '+' || c == '-' || c == '*' || c == '/':
		return str{next: op, doEmit: true, typ: IntLit}

	case c == ')':
		return str{next: paren, doEmit: true, typ: IntLit}

	case c == '.':
		return str{next: floatLit}

	default:
		return str{err: errors.New(fmt.Sprintf("Lexer: unexpected char %c in integer literal", c))}
	}
}

func floatLit(c rune) str {
	switch {
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: start, doEmit: true, typ: FloatLit}

	case '0' <= c && c <= '9':
		return str{next: floatLit}

	case c == '+' || c == '-' || c == '*' || c == '/':
		return str{next: op, doEmit: true, typ: FloatLit}

	case c == ')':
		return str{next: paren, doEmit: true, typ: FloatLit}

	default:
		return str{err: errors.New(fmt.Sprintf("Lexer: unexpected char %c in float literal", c))}
	}
}

func varName(c rune) str {
	switch {
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: start, doEmit: true, typ: VarName}

	case 'a' <= c && c <= 'z':
		return str{next: varName}

	case c == '+' || c == '-' || c == '*' || c == '/':
		return str{next: op, doEmit: true, typ: VarName}

	case c == ')':
		return str{next: paren, doEmit: true, typ: VarName}

	case c == '=':
		return str{next: assign, doEmit: true, typ: VarName}

	default:
		return str{err: errors.New(fmt.Sprintf("Lexer: unexpected char %c in variable name", c))}
	}
}

func op(c rune) str {
	switch {
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: start, doEmit: true, typ: Op}

	case '0' <= c && c <= '9':
		return str{next: intLit, doEmit: true, typ: Op}

	case 'a' <= c && c <= 'z':
		return str{next: varName, doEmit: true, typ: Op}

	case c == '(':
		return str{next: paren, doEmit: true, typ: Op}

	case c == '-':
		return str{next: op, doEmit: true, typ: Op}

	default:
		return str{err: errors.New(fmt.Sprintf("Lexer: unexpected char %c following operator", c))}
	}
}

func paren(c rune) str {
	switch {
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: start, doEmit: true, typ: Paren}

	case '0' <= c && c <= '9':
		return str{next: intLit, doEmit: true, typ: Paren}

	case 'a' <= c && c <= 'z':
		return str{next: varName, doEmit: true, typ: Paren}

	case c == '(':
		return str{next: paren, doEmit: true, typ: Paren}

	case c == '-':
		return str{next: op, doEmit: true, typ: Paren}

	default:
		return str{err: errors.New(fmt.Sprintf("Lexer: unexpected char %c following parenthesis", c))}
	}
}

func assign(c rune) str {
	switch {
	case c == ' ' || c == '\t' || c == '\n' || c == eof:
		return str{next: start, doEmit: true, typ: Assign}

	case '0' <= c && c <= '9':
		return str{next: intLit, doEmit: true, typ: Assign}

	case 'a' <= c && c <= 'z':
		return str{next: varName, doEmit: true, typ: Assign}

	case c == '(':
		return str{next: paren, doEmit: true, typ: Assign}

	case c == '-':
		return str{next: op, doEmit: true, typ: Assign}

	default:
		return str{err: errors.New(fmt.Sprintf("Lexer: unexpected char %c following assignment", c))}
	}
}

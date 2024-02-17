package node

import (
	"fmt"
	"strings"
)

type PrettyPrinter interface {
	PrettyPrint(indent int)
}

func indent(i int, s string) string {
	return strings.Repeat(" ", i*3) + s
}

func (c Call) PrettyPrint(d int) {
	fmt.Println(indent(d, "("))
	for _, t := range c.Arguments.Elems {
		t.PrettyPrint(d + 1)
	}
	fmt.Println(indent(d, ")"))
}

func (f Function) PrettyPrint(d int) {
	f.Parameters.PrettyPrint(d + 1)
	fmt.Println(indent(d, "->"))
	f.Body.PrettyPrint(d + 1)
}

func (i Int) PrettyPrint(d int) { fmt.Println(indent(d, string(i))) }

func (f Float) PrettyPrint(d int) { fmt.Println(indent(d, string(f))) }

func (s String) PrettyPrint(d int) { fmt.Println(indent(d, string(s))) }

func (b BinOp) PrettyPrint(d int) {
	fmt.Println(indent(d, b.Op))
	b.Left.PrettyPrint(d + 1)
	b.Right.PrettyPrint(d + 1)
}

func (u UnOp) PrettyPrint(d int) {
	fmt.Println(indent(d, u.Op))
	u.Target.PrettyPrint(d + 1)
}

func (i IndexAt) PrettyPrint(d int) {
	i.Ary.PrettyPrint(d)
	fmt.Print("@")
	i.At.PrettyPrint(0)
}

func (i IndexFromTo) PrettyPrint(d int) {
	i.Ary.PrettyPrint(d)
	fmt.Print("@")
	i.From.PrettyPrint(0)
	fmt.Print(":")
	i.To.PrettyPrint(0)
}

func (i If) PrettyPrint(d int) {
	fmt.Print(indent(d, "if "))
	i.Condition.PrettyPrint(0)
	i.TrueCase.PrettyPrint(d + 1)
}

func (ie IfElse) PrettyPrint(d int) {
	fmt.Print(indent(d, "if "))
	ie.Condition.PrettyPrint(0)
	ie.TrueCase.PrettyPrint(d + 1)
	fmt.Println(indent(d, "else"))
	ie.FalseCase.PrettyPrint(d + 1)
}

func (w While) PrettyPrint(d int) {
	fmt.Print(indent(d, "while "))
	w.Condition.PrettyPrint(0)
	w.Body.PrettyPrint(d + 1)
}

func (r Return) PrettyPrint(d int) {
	fmt.Print(indent(d, "return "))
	r.Target.PrettyPrint(0)
}

func (r Read) PrettyPrint(d int) {
	fmt.Print(indent(d, "read "))
	r.Target.PrettyPrint(0)
}

func (w Write) PrettyPrint(d int) {
	fmt.Print(indent(d, "write "))
	w.Value.PrettyPrint(0)
}

func (r Repl) PrettyPrint(d int) { fmt.Print(indent(d, "repl")) }

func (n Name) PrettyPrint(d int) { fmt.Print(indent(d, string(n))) }

func (b Block) PrettyPrint(d int) {
	fmt.Println(indent(d, "{"))
	for _, t := range b.Body {
		t.PrettyPrint(d + 1)
	}
	fmt.Println(indent(d, "}"))
}

func (l List) PrettyPrint(d int) {
	fmt.Println(indent(d, "("))
	for _, t := range l.Elems {
		t.PrettyPrint(d + 1)
	}
	fmt.Println(indent(d, ")"))
}

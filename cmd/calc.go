package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/stack"
	"github.com/paulsonkoly/calc/types/node"
)

func main() {
  b := builtin.Type{}
	s := stack.NewStack(b)
  p := parser.Type{}

	switch len(os.Args) {

	case 1: // REPL mode
		rl := node.NewRLReader()
		defer rl.Close()
		node.Loop(rl, p, s, true)

	case 2: // file mode
		fileName := os.Args[1]
		fr := node.NewFReader(fileName)
		defer fr.Close()
		node.Loop(fr, p, s, false)

	case 3: // command line mode
		var eval string
		flag.StringVar(&eval, "eval", "", "string to evaluate")
		flag.Parse()
		cmdLine(eval)
	}
}

func cmdLine(line string) {
  b := builtin.Type{}
	s := stack.NewStack(b)
	t, err := parser.Parse(line)
	if len(t) > 0 {
		fmt.Println(node.Evaluate(s, t[0]))
	}
	if err != nil {
		fmt.Println(err)
	}
}

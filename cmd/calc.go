package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/node"
)

func main() {
	m := memory.NewMemory()
	p := parser.Type{}

  builtin.Load(m)

	switch len(os.Args) {

	case 1: // REPL mode
		rl := node.NewRLReader()
		defer rl.Close()
		node.Loop(rl, p, m, true)

	case 2: // file mode
		fileName := os.Args[1]
		fr := node.NewFReader(fileName)
		defer fr.Close()
		node.Loop(fr, p, m, false)

	case 3: // command line mode
		var eval string
		flag.StringVar(&eval, "eval", "", "string to evaluate")
		flag.Parse()
		cmdLine(eval)
	}
}

func cmdLine(line string) {
	m := memory.NewMemory()
  builtin.Load(m)

	t, err := parser.Parse(line)
	if len(t) > 0 {
		fmt.Println(node.Evaluate(m, t[0]))
	}
	if err != nil {
		fmt.Println(err)
	}
}

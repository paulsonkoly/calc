package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/paulsonkoly/calc/evaluator"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/stack"
)

func main() {
	s := stack.NewStack()

	switch len(os.Args) {

	case 1: // REPL mode
		rl := evaluator.NewRLReader()
		defer rl.Close()
		evaluator.Loop(rl, s, true)

	case 2: // file mode
		fileName := os.Args[1]
		fr := evaluator.NewFReader(fileName)
		defer fr.Close()
		evaluator.Loop(fr, s, false)

	case 3: // command line mode
		var eval string
		flag.StringVar(&eval, "eval", "", "string to evaluate")
		flag.Parse()
		cmdLine(eval)
	}
}

func cmdLine(line string) {
	s := stack.NewStack()
	t, err := parser.Parse(line)
	if len(t) > 0 {
		fmt.Println(evaluator.Evaluate(s, t[0]))
	}
	if err != nil {
		fmt.Println(err)
	}
}

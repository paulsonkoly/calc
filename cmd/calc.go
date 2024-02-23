package main

import (
	"flag"
	"fmt"
	"runtime/pprof"

	"os"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/node"
)

func main() {
	var eval string
	var cpuprof string
	var ast bool
	flag.StringVar(&eval, "eval", "", "string to evaluate")
	flag.StringVar(&cpuprof, "cpuprof", "", "filename for go pprof")
	flag.BoolVar(&ast, "ast", false, "repl outputs AST instead of evaluating")

	flag.Parse()

	m := memory.NewMemory()
	p := parser.Type{}

	builtin.Load(m)

	if cpuprof != "" {
		f, err := os.Create(cpuprof)
		if err != nil {
			panic(err)
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}

		defer pprof.StopCPUProfile()
	}

	if eval != "" { // cmd line mode
		cmdLine(eval)
		return
	}

	if flag.NArg() >= 1 { // file mode
		fileName := flag.Arg(0)
		fr := node.NewFReader(fileName)
		defer fr.Close()
		node.Loop(fr, p, m, false, ast)
		return
	}

	// REPL mode
	fmt.Println("calc repl")
	rl := node.NewRLReader()
	defer rl.Close()
	node.Loop(rl, p, m, true, ast)
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

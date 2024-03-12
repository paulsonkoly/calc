package main

import (
	"flag"
	"fmt"
	"runtime/pprof"

	"os"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
)

func main() {
	var eval string
	var cpuprof string
	var ast bool
	var bcFlag bool
	flag.StringVar(&eval, "eval", "", "string to evaluate")
	flag.StringVar(&cpuprof, "cpuprof", "", "filename for go pprof")
	flag.BoolVar(&ast, "ast", false, "repl outputs AST instead of evaluating")
	flag.BoolVar(&bcFlag, "bytecode", false, "repl outputs bytecode instead of evaluating")

	flag.Parse()

	m := memory.NewMemory()
	p := parser.Type{}
	cs := []bytecode.Type{}
	ds := []value.Type{}

	builtin.Load(m, &cs, &ds)

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
		node.Loop(fr, p, m, &cs, &ds, false, ast, bcFlag)
		return
	}

	// REPL mode
	fmt.Println("calc repl")
	rl := node.NewRLReader()
	defer rl.Close()
	node.Loop(rl, p, m, &cs, &ds, true, ast, bcFlag)
}

func cmdLine(line string) {
	m := memory.NewMemory()
	cs := []bytecode.Type{}
	ds := []value.Type{}

	builtin.Load(m, &cs, &ds)

	t, err := parser.Parse(line)
	if len(t) > 0 {
		fmt.Println(node.Evaluate(m, t[0]))
	}
	if err != nil {
		fmt.Println(err)
	}
}

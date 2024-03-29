package main

import (
	"flag"
	"fmt"
	"runtime/pprof"

	"os"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/flags"
	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/compresult"
	"github.com/paulsonkoly/calc/types/dbginfo"
	"github.com/paulsonkoly/calc/types/node"
	"github.com/paulsonkoly/calc/types/value"
	"github.com/paulsonkoly/calc/vm"
)

func main() {
	flag.Parse()

	m := memory.New()
	p := parser.Type{}
	cs := []bytecode.Type{}
	ds := []value.Type{}
	dbg := make(dbginfo.Type)
	cr := compresult.Type{CS: &cs, DS: &ds, Dbg: &dbg}

	builtin.Load(cr)
	virtM := vm.New(m, cr)

	if *flags.CPUProfFlag != "" {
		f, err := os.Create(*flags.CPUProfFlag)
		if err != nil {
			panic(err)
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}

		defer pprof.StopCPUProfile()
	}

	if *flags.EvalFlag != "" { // cmd line mode
		t, err := parser.Parse(*flags.EvalFlag)

		if err != nil {
			fmt.Println(err)
		}

		if len(t) > 0 {
			n := t[0]
			node.ByteCode(n, cr)
			if v, err := virtM.Run(true); err == nil {
				fmt.Println(v)
			}
		}
		return
	}

	if flag.NArg() >= 1 { // file mode
		fileName := flag.Arg(0)
		fr := node.NewFReader(fileName)
		defer fr.Close()
		node.Loop(fr, p, virtM, false)
		return
	}

	// REPL mode
	fmt.Println("calc repl")
	rl := node.NewRLReader()
	defer rl.Close()
	node.Loop(rl, p, virtM, true)
}

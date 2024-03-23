package node

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/paulsonkoly/calc/flags"
	"github.com/paulsonkoly/calc/vm"
)

type lineReader interface {
	read() (string, error)
	io.Closer
}

type RLReader struct{ r *readline.Instance }

func NewRLReader() RLReader {
	r, err := readline.New("")
	if err != nil {
		panic(err)
	}
	return RLReader{r: r}
}

func (rl RLReader) read() (string, error) { return rl.r.Readline() }

func (rl RLReader) Close() error { return rl.r.Close() }

type FReader struct {
	r *os.File
	b *bufio.Reader
}

func NewFReader(fn string) FReader {
	r, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}

	b := bufio.NewReader(r)
	return FReader{r: r, b: b}
}

func (f FReader) read() (string, error) { return f.b.ReadString('\n') }

func (f FReader) Close() error { return f.r.Close() }

type Parser interface {
	Parse(string) ([]Type, error)
}

func Loop(r lineReader, p Parser, vm *vm.Type, doOut bool) {
	blocksOpen := 0
	quotesOpen := 0
	bracketsOpen := 0
	input := ""
	sep := ""

	for {
		line, err := r.read()
		if err != nil { // io.EOF
			break
		}

		blocksOpen += strings.Count(line, "{") - strings.Count(line, "}")
		quotesOpen += strings.Count(line, "\"") - strings.Count(line, "\\\"")
		bracketsOpen += strings.Count(line, "[") - strings.Count(line, "]")
		input += sep + line
		sep = "\n"

		if blocksOpen == 0 && quotesOpen%2 == 0 && bracketsOpen == 0 {
			t, err := p.Parse(input)
			input = ""
			sep = ""
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, e := range t {
				e := e.STRewrite(SymTbl{})

				if *flags.AstFlag {
					Graphviz(e)
				}

				ip := len(*vm.CS)
				if doOut {
					ByteCode(e, vm.CS, vm.DS)
				} else {
					ByteCodeNoStck(e, vm.CS, vm.DS)
				}

				if *flags.ByteCodeFlag {
					for i, c := range (*vm.CS)[ip:] {
						fmt.Printf(" %8d | %v\n", ip+i, c)
					}
				}

				v := vm.Run(doOut)
				if doOut {
					fmt.Printf("> %s\n", v.Display())
				}
			}
		}
	}
}

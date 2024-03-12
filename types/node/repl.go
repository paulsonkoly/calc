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
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/value"
	"github.com/paulsonkoly/calc/vm"
)

type lineReader interface {
	read() (string, error)
	io.Closer
}

type rlReader struct{ r *readline.Instance }

func NewRLReader() rlReader {
	r, err := readline.New("")
	if err != nil {
		panic(err)
	}
	return rlReader{r: r}
}

func (rl rlReader) read() (string, error) { return rl.r.Readline() }

func (r rlReader) Close() error { return r.r.Close() }

type fReader struct {
	r *os.File
	b *bufio.Reader
}

func NewFReader(fn string) fReader {
	r, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}

	b := bufio.NewReader(r)
	return fReader{r: r, b: b}
}

func (f fReader) read() (string, error) { return f.b.ReadString('\n') }

func (f fReader) Close() error { return f.r.Close() }

type Parser interface {
	Parse(string) ([]Type, error)
}

func Loop(r lineReader, p Parser, vm *vm.Type, cs *[]bytecode.Type, ds *[]value.Type, doOut bool) {
	blocksOpen := 0
	quotesOpen := 0
	input := ""

	for {
		line, err := r.read()
		if err != nil { // io.EOF
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		blocksOpen += strings.Count(line, "{") - strings.Count(line, "}")
		quotesOpen += strings.Count(line, "\"") - strings.Count(line, "\\\"")
		input = join(input, line)

		if blocksOpen == 0 && quotesOpen%2 == 0 {
			t, err := p.Parse(input)
			if err != nil {
				fmt.Println(err)
				input = ""
				continue
			}
			if len(t) < 1 {
				continue
			}

			for _, e := range t {
				e := e.STRewrite(SymTbl{})

				if *flags.AstFlag {
					Graphviz(e)
				}

				ip := len(*cs)
				if doOut {
					ByteCode(e, cs, ds)
				} else {
					ByteCodeNoStck(e, cs, ds)
				}

				if *flags.ByteCodeFlag {
					for i, c := range (*cs)[ip:] {
						fmt.Printf(" %8d | %v\n", ip+i, c)
					}
				}

				vm.SetSegments(*cs, *ds)
				v := vm.Run(doOut)
				if doOut {
					fmt.Printf("> %v\n", v)
				}
			}
			input = ""
		}
	}
}

func join(a, b string) string {
	if a == "" {
		return b
	}
	return a + "\n" + b
}

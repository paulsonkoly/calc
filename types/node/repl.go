package node

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/paulsonkoly/calc/combinator"
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

type ParserError = *combinator.Error 

type Parser interface {
	Parse(string) ([]Type, ParserError)
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
			sep = ""
			if err != nil {
				reportError(err, input)
				input = ""
				continue
			}
			input = ""

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

func reportError(err ParserError, line string) {
	fmt.Println(err.Message())
	start := strings.LastIndex(line[0:err.From()], "\n")
	if start == -1 {
		start = 0
	}
	end := strings.Index(line[err.To():], "\n")
	if end == -1 {
		end = len(line)
	} else {
    end += err.To()
  }
	fmt.Println(line[start:end])
  empty := ""
  if err.From() > start {
    empty = strings.Repeat(" ", err.From() - start - 1)
  }
	squiggly := ""
	if err.To() > err.From() {
		squiggly = strings.Repeat("~", err.To()-err.From() - 1)
	}
	fmt.Println(empty + "^" + squiggly + "^")
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"

	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/paulsonkoly/calc/evaluator"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/stack"
)

func main() {
	switch len(os.Args) {

	case 1: // REPL mode
		rl := newRLReader()
		defer rl.Close()
		loop(rl, true)

	case 2: // file mode
		fileName := os.Args[1]
		fr := newFReader(fileName)
		defer fr.Close()
		loop(fr, false)

	case 3: // command line mode
		var eval string
		flag.StringVar(&eval, "eval", "", "string to evaluate")
		flag.Parse()
		cmdLine(eval)
	}
}

type lineReader interface {
	read() (string, error)
	io.Closer
}

type rlReader struct{ r *readline.Instance }

func newRLReader() rlReader {
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

func newFReader(fn string) fReader {
	r, err := os.Open(fn)
	if err != nil {
		panic(err)
	}

	b := bufio.NewReader(r)
	return fReader{r: r, b: b}
}

func (f fReader) read() (string, error) { return f.b.ReadString('\n') }

func (f fReader) Close() error { return f.r.Close() }

func loop(r lineReader, doOut bool) {
	s := stack.NewStack()
	blocksOpen := 0
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
		input = join(input, line)

		if blocksOpen == 0 {
			t, err := parser.Parse(input)
			if err != nil {
				fmt.Println(err)
				input = ""
				continue
			}
			if len(t) < 1 {
				continue
			}
			v := evaluator.Evaluate(s, t[0])
			for _, e := range t[1:] {
				v = evaluator.Evaluate(s, e)
			}

			if doOut {
				fmt.Println("> ", v)
			}
			input = ""
		}
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

func join(a, b string) string {
	if a == "" {
		return b
	}
	return a + "\n" + b
}

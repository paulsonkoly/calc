package evaluator

import (
	"bufio"
	"fmt"
	"io"

	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/paulsonkoly/calc/parser"
	"github.com/paulsonkoly/calc/stack"
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
		panic(err)
	}

	b := bufio.NewReader(r)
	return fReader{r: r, b: b}
}

func (f fReader) read() (string, error) { return f.b.ReadString('\n') }

func (f fReader) Close() error { return f.r.Close() }

func Loop(r lineReader, s stack.Stack, doOut bool) {
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
			t, err := parser.Parse(input)
			if err != nil {
				fmt.Println(err)
				input = ""
				continue
			}
			if len(t) < 1 {
				continue
			}
			v := Evaluate(s, t[0])
			for _, e := range t[1:] {
				v = Evaluate(s, e)
			}

			if doOut {
				fmt.Println("> ", v)
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

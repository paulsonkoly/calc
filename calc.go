package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/phaul/calc/evaluator"
	"github.com/phaul/calc/parser"
)

var eval = flag.String("eval", "", "string to evaluate")

func main() {
	vars := make(evaluator.Variables)
	flag.Parse()
	if *eval != "" {
		t, err := parser.Parse(*eval)
		if len(t) > 0 {
			fmt.Println("> ", evaluator.Evaluate(vars, t[0]))
		}
		if err != nil {
			fmt.Println(err)
		}
	} else {
		r := bufio.NewReader(os.Stdin)
		blocksOpen := 0
		input := ""
		for {
			line, _ := r.ReadString('\n')
			line = strings.TrimSpace(line)
			blocksOpen += strings.Count(line, "{") - strings.Count(line, "}")
			input = join(input, line)
			if blocksOpen == 0 {
				t, err := parser.Parse(input)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if len(t) < 1 {
					continue
				}
				v := evaluator.Evaluate(vars, t[0])
				for _, e := range t[1:] {
					v = evaluator.Evaluate(vars, e)
				}
				fmt.Println("> ", v)
				input = ""
			}
		}
	}
}

func join(a, b string) string {
	if a == "" {
		return b
	}
	return a + "\n" + b
}
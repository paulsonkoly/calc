// stack provides the call stack to our interpreter.
//
// It implements variable lookup and pushing stack frames on function calls,
// and popping stack frames on returns.
// We have very simple lookup rules. In case of a variable name lookup start at
// the most recent frame and keep going up the stack until a variable is found.
// In case of setting a variable we always set the variable in the current
// frame. Use v, ok := style signalling if a variable is not defined in any of
// the frames.
package stack

import (
	"fmt"

	"github.com/paulsonkoly/calc/builtin"
	"github.com/paulsonkoly/calc/types/value"
)

type sframe = value.Frame
type svalue = value.Type

type Stack []sframe

// NewStack creates a new stack
func NewStack() Stack {
	topF := make(sframe)
	for name, f := range builtin.All {
		topF[name] = f
	}
	return []sframe{topF}
}

// LookUp looks up a variable value
func (s Stack) LookUp(name string) (svalue, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		f := s[i]
		if v, ok := f[name]; ok {
			return v, true
		}
	}
	return value.Error(fmt.Sprintf("%s not defined", name)), false
}

// Set sets a variable to a value in the current frame, ignoring lookup
func (s Stack) Set(name string, v svalue) {
	s[len(s)-1][name] = v
}

// Push pushes a new frame on the stack
func (s *Stack) Push(fr *sframe) {
	if fr == nil {
		*s = append(*s, make(sframe))
	} else {
		*s = append(*s, *fr)
	}
}

// Pop pops a frame from the stack
//
// It returns a pointer to the frame popped, in case of a function returning a
// function value, as a closure we need to keep reference to the enclosing
// environment. It will be attached to the returned function value. Otherwise
// the value can be ignored by the caller.
func (s *Stack) Pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *Stack) Top() *sframe {
	r := (*s)[len(*s)-1]
	return &r
}

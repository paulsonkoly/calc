// stack provides the call stack to our interpretter.
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

	"github.com/phaul/calc/types"
)

type Stack []frame

type frame map[string]types.Value

// NewStack creates a new stack
func NewStack() Stack {
	topF := make(frame)
	return []frame{topF}
}

// LookUp looks up a variable value
func (s Stack) LookUp(name string) (types.Value, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		f := s[i]
		if v, ok := f[name]; ok {
			return v, true
		}
	}
	return types.ValueError(fmt.Sprintf("%s not defined", name)), false
}

// Set sets a variable to a value in the current frame, ignoring lookup
func (s Stack) Set(name string, v types.Value) {
	s[len(s)-1][name] = v
}

// Push pushes a new frame on the stack
func (s *Stack) Push() {
	*s = append(*s, make(frame))
}

// Pop pos a frame from the stack
func (s *Stack) Pop() {
	*s = (*s)[:len(*s)-1]
}

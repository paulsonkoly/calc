// memory provides our memory model
//
// Local and closure variables are accessed via symbol tbl index. Global
// variables are accessed using their name.
//
// The global frame is special, it's a map from variable names to values. This
// is because we can gradually parse more and more code that can define new
// global variables, so the symbol table phase can't work out a symbol tbl
// index for these variables.
//
// Other frames are slices, whose size don't change. These are indexed by the
// symbol table index, and they belong to a lexical scope.
package memory

import (
	"fmt"
	"github.com/paulsonkoly/calc/types/value"
)

type Frame []value.Type
type gframe map[string]value.Type

// NewFrame creates a new frame, either Closure or Local
func NewFrame(size int) Frame {
	return make(Frame, size)
}

// Set sets a variable value
func (f Frame) Set(symIdx int, v value.Type) { f[symIdx] = v }

// Memory holds all variables
type Type struct {
	global gframe
	stack  []Frame
}

// NewType creates a new memory, with an empty global frame and an empty stack
func NewMemory() *Type { return &Type{global: gframe{}, stack: []Frame{}} }

// SetGlobal sets a global variable
func (m *Type) SetGlobal(name string, v value.Type) { m.global[name] = v }

// Set sets a local variable
func (m *Type) Set(symIdx int, v value.Type) { m.stack[len(m.stack)-1][symIdx] = v }

// LookUpLocal looks up a local variable
func (m *Type) LookUpLocal(symIdx int) value.Type { return m.stack[len(m.stack)-1][symIdx] }

// LookUpClosure looks up a closure variable. A variable that was local in the
// containing lexical scope
func (m *Type) LookUpClosure(symIdx int) value.Type { return m.stack[len(m.stack)-2][symIdx] }

// LookUpGlobal looks up a global variable
func (m *Type) LookUpGlobal(name string) value.Type {
	v, ok := m.global[name]
	if !ok {
		s := fmt.Sprintf("%s not defined", name)
		return value.NewError(&s)
	}
	return v
}

// PushFrame pushes a stack frame
func (m *Type) PushFrame(f Frame) { m.stack = append(m.stack, f) }

// PopFrame pops a stack frame
func (m *Type) PopFrame() { m.stack = m.stack[0 : len(m.stack)-1] }

// Top is the last stack frame pushed
func (m *Type) Top() Frame {
	if len(m.stack) < 1 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

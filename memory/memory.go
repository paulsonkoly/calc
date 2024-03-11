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
// Other frames are slices on the stack. These are indexed by the fp (frame
// pointer) offseted by symbol table index. Local variables of the frame are
// between fp and le (frame pointer and local end).
package memory

import (
	"fmt"
	"slices"

	"github.com/paulsonkoly/calc/types/value"
)

const minStackSize = 128

const (
	localFE   = -1
	localFP   = -2
	closureFE = -3
	closureFP = -4
)

type Frame []value.Type
type gframe map[string]value.Type

// Memory holds all variables
type Type struct {
	sp     int
	fp     []int
	global gframe
	stack  []value.Type
}

// NewType creates a new memory, with an empty global frame and an empty stack
func NewMemory() *Type { return &Type{fp: []int{}, global: gframe{}, stack: []value.Type{}} }

// SetGlobal sets a global variable
func (m *Type) SetGlobal(name string, v value.Type) { m.global[name] = v }

// Set sets a local variable
func (m *Type) Set(symIdx int, v value.Type) {
	fp := m.fp[len(m.fp)+localFP]
	sp := fp + symIdx
	m.stack[sp] = v
}

// LookUpLocal looks up a local variable
func (m *Type) LookUpLocal(symIdx int) value.Type {
	fp := m.fp[len(m.fp)+localFP]
	sp := fp + symIdx
	return m.stack[sp]
}

// LookUpClosure looks up a closure variable. A variable that was local in the
// containing lexical scope
func (m *Type) LookUpClosure(symIdx int) value.Type {
	fp := m.fp[len(m.fp)+closureFP]
	sp := fp + symIdx
	return m.stack[sp]
}

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
func (m *Type) PushFrame(f Frame) {
	m.fp = append(m.fp, m.sp, m.sp+len(f))
	if len(f) > 0 {
		m.growStack(len(f))
		m.stack = slices.Replace(m.stack, m.sp, m.sp+len(f), f...)
		m.sp += len(f)
	}
}

func (m *Type) Push(v value.Type) {
	m.growStack(1)
	m.stack[m.sp] = v
	m.sp++
}

// PopFrame pops a stack frame
func (m *Type) PopFrame() {
	fp := m.fp[len(m.fp)+localFP]
	m.sp = fp
	m.fp = m.fp[:len(m.fp)-2]
}

func (m *Type) Pop() value.Type {
  m.sp--
	return m.stack[m.sp]
}

// Top is the last stack frame pushed
func (m *Type) Top() Frame {
	if len(m.fp) < 1 {
		return nil
	}
	fp := m.fp[len(m.fp)+localFP]
	le := m.fp[len(m.fp)+localFE]
	return m.stack[fp:le]
}

func (m *Type) growStack(size int) {
	if m.sp+size >= len(m.stack) {
		m.stack = append(m.stack, make([]value.Type, max(minStackSize, size))...)
	}
}

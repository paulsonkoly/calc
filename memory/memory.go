// Package memory provides our memory model.
//
// There are 3 main regions. Global, the closure stack and the normal stack.
//
// Local and closure variables are accessed via symbol tbl index. Global
// variables are accessed using their name.
//
// The global region is special, it's a map from variable names to values. This
// is because we can gradually parse more and more code that can define new
// global variables, so the symbol table phase can't work out a symbol tbl
// index for these variables.
//
// The closure region is pointers to cloned slices of the normal stack. When a
// function returns a function we save the frame of the defining function in
// the returned function value. When a function is called apart from pushing
// the normal frame on the normal stack we need to push the closure frame from
// the function value  on the closure stack.
//
// Normal stack is an ever growing slice of values. The fp has a pair of
// pointers into the stack per frame: fp and le the frame pointer and local
// end. Local variables live on the normal stack. Function arguments count as
// local variables. Local variables of the frame are between fp and le . le
// points to the function return address. After le it's stack scratch area.
package memory

import (
	"fmt"

	"github.com/paulsonkoly/calc/types/dbginfo"
	"github.com/paulsonkoly/calc/types/value"
)

const minStackSize = 128

const (
	localFE = -1
	localFP = -2
)

type Frame []value.Type
type gframe map[string]value.Type

var id int

// Memory holds all variables.
type Type struct {
	id      int
	sp      int
	fp      []int
	global  gframe
	closure []Frame
	stack   []value.Type
}

// New creates a new memory, with an empty global frame and an empty stack.
func New() *Type {
	id = -1
	return &Type{id: id, fp: []int{}, global: gframe{}, closure: []Frame{}, stack: []value.Type{}}
}

// ID returns a unique ID for this memory.
func (m *Type) ID() int {
	return m.id
}

// Clone does a memory copy for context switching.
//
// The clone would point to the same global, same closure, and the last frame
// of the stack will be deep copied. reuse can be nil, when it's not it's
// resources are re-used to create a new memory.
var TotalCnt = 0
var AllocCnt = 0

func (m *Type) Clone(reuse *Type) *Type {
	var newStackSize int
	var newStack []value.Type

	id++
	TotalCnt++

	if len(m.fp) < 2 {
		newStackSize = minStackSize
	} else {
		newStackSize = max(m.sp-m.fp[len(m.fp)+localFP], minStackSize)
	}

	if reuse != nil {
		if len(reuse.stack) < newStackSize {
			reuse.growStack(newStackSize - len(reuse.stack))
		}
		newStack = reuse.stack
	} else {
		newStack = make([]value.Type, newStackSize)
		AllocCnt++
	}

	if len(m.fp) < 2 {
		return &Type{id: id, sp: 0, fp: []int{}, global: m.global, closure: m.closure, stack: newStack}
	}

	fp := m.fp[len(m.fp)+localFP]
	le := m.fp[len(m.fp)+localFE]

	copy(newStack, m.stack[fp:m.sp])

	return &Type{id: id, sp: m.sp - fp, fp: []int{0, le - fp}, global: m.global, closure: m.closure, stack: newStack}
}

// CallDepth is the number of call frames.
func (m *Type) CallDepth() int {
	return len(m.fp) / 2
}

// SetGlobal sets a global variable.
func (m *Type) SetGlobal(name string, v value.Type) { m.global[name] = v }

// Set sets a local variable.
func (m *Type) Set(symIdx int, v value.Type) {
	fp := m.fp[len(m.fp)+localFP]
	sp := fp + symIdx
	m.stack[sp] = v
}

// LookUpLocal looks up a local variable.
func (m *Type) LookUpLocal(symIdx int) value.Type {
	fp := m.fp[len(m.fp)+localFP]
	sp := fp + symIdx
	return m.stack[sp]
}

// LookUpClosure looks up a closure variable. A variable that was local in the
// containing lexical scope.
func (m *Type) LookUpClosure(symIdx int) value.Type {
	return m.closure[len(m.closure)-1][symIdx]
}

// LookUpGlobal looks up a global variable.
func (m *Type) LookUpGlobal(name string) value.Type {
	v, ok := m.global[name]
	if !ok {
		return value.Nil
	}
	return v
}

// PushFrame pushes a stack frame.
func (m *Type) PushFrame(argsCnt, localCnt int) {
	locals := localCnt - argsCnt
	m.growStack(localCnt - argsCnt)
	for i := m.sp; i < m.sp+locals; i++ {
		m.stack[i] = value.Nil
	}
	m.sp += localCnt - argsCnt
	m.fp = append(m.fp, m.sp-localCnt, m.sp)
}

// Push pushes a value.
func (m *Type) Push(v value.Type) {
	m.growStack(1)
	m.stack[m.sp] = v
	m.sp++
}

// PushClosure pushes the closure frame.
func (m *Type) PushClosure(f Frame) {
	m.closure = append(m.closure, f)
}

// PopFrame pops a stack frame.
func (m *Type) PopFrame() {
	fp := m.fp[len(m.fp)+localFP]
	m.sp = fp
	m.fp = m.fp[:len(m.fp)-2]
}

// Pop pops the last pushed value decrementing the stack pointer.
func (m *Type) Pop() value.Type {
	m.sp--
	return m.stack[m.sp]
}

// PopClosure pops a frame from the closure region.
func (m *Type) PopClosure() {
	m.closure = m.closure[:len(m.closure)-1]
}

// Top is the last stack frame pushed.
func (m *Type) Top() Frame {
	if len(m.fp) < 1 {
		return nil
	}
	fp := m.fp[len(m.fp)+localFP]
	le := m.fp[len(m.fp)+localFE]
	return m.stack[fp:le]
}

// IP returns the function return address.
func (m *Type) IP() *value.Type {
	if len(m.fp)+localFE < 0 {
		return nil
	}
	le := m.fp[len(m.fp)+localFE]
	return &m.stack[le]
}

// ResetSP resets the stack pointer to 0.
func (m *Type) ResetSP() {
	m.sp = 0
}

func (m *Type) growStack(size int) {
	if m.sp+size >= len(m.stack) {
		m.stack = append(m.stack, make([]value.Type, max(minStackSize, size))...)
	}
}

// DumpStack is a debug dump of the stack.
func (m *Type) DumpStack(dbg *dbginfo.Type) {
	fmt.Println("= stack =============================================")
	for i := len(m.fp) - 1; i >= 0; i -= 2 {
		ipAddr := m.fp[i]
		ipv := m.stack[ipAddr]
		ip, ok := ipv.ToInt()
		if !ok {
			fmt.Println("corrupt stack. giving up")
			return
		}
		info, ok := (*dbg)[ip]
		if !ok {
			fmt.Println("No debug info found for call. giving up")
			return
		}
		name := info.Name
		argCnt := info.ArgCnt

		if i < 1 {
			fmt.Println("corrupt frame pointer. giving up")
			return
		}
		fp := m.fp[i-1]
		argv := m.stack[fp : fp+argCnt]

		args := ""
		sep := ""
		for i, v := range argv {
			args += fmt.Sprintf("%sarg[%d]: %s", sep, i, v.Abbrev())
			sep = " "
		}

		fmt.Printf("IP: %d %s() args: %s\n", ip, name, args)
	}
	fmt.Println("=====================================================")
}

// Reset drops all stack local allocations.
func (m *Type) Reset() {
	m.sp = 0
	m.closure = []Frame{}
	m.fp = []int{}
}

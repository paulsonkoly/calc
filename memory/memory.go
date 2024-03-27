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

// Memory holds all variables.
type Type struct {
	sp      int
	fp      []int
	global  gframe
	closure []Frame
	stack   []value.Type
}

// New creates a new memory, with an empty global frame and an empty stack.
func New() *Type {
	return &Type{fp: []int{}, global: gframe{}, closure: []Frame{}, stack: []value.Type{}}
}

// Clone does a memory copy for context switching.
//
// The clone would point to the same global, same closure, and the last frame
// of the stack will be deep copied.
func (m *Type) Clone() *Type {
	if len(m.fp) < 2 {
		frm := slices.Clone(m.stack)
		return &Type{sp: m.sp, fp: []int{}, global: m.global, closure: m.closure, stack: frm}
	}
	fp := m.fp[len(m.fp)+localFP]
	le := m.fp[len(m.fp)+localFE]
	frm := slices.Clone(m.stack[fp:m.sp])
	return &Type{sp: len(frm), fp: []int{0, le - fp}, global: m.global, closure: m.closure, stack: frm}
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
	m.sp += localCnt - argsCnt
	m.fp = append(m.fp, m.sp-localCnt, m.sp)
}

func (m *Type) Push(v value.Type) {
	m.growStack(1)
	m.stack[m.sp] = v
	m.sp++
}

func (m *Type) PushClosure(f Frame) {
	m.closure = append(m.closure, f)
}

// PopFrame pops a stack frame.
func (m *Type) PopFrame() {
	fp := m.fp[len(m.fp)+localFP]
	m.sp = fp
	m.fp = m.fp[:len(m.fp)-2]
}

func (m *Type) Pop() value.Type {
	m.sp--
	return m.stack[m.sp]
}

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

func (m *Type) IP() *value.Type {
	if len(m.fp)+localFE < 0 {
		return nil
	}
	le := m.fp[len(m.fp)+localFE]
	return &m.stack[le]
}

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

// Package vm implements the virtual machine.
package vm

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"

	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/compresult"
	"github.com/paulsonkoly/calc/types/value"

	"github.com/kamstrup/intmap"
)

var (
	ErrConversion = errors.New("conversion error")
	ErrArity      = errors.New("arity mismatch")
)

const (
	minAllocContexts = 16
)

type context struct {
	ip       int                           // instruction pointer
	m        *memory.Type                  // variables
	parent   *context                      // parent context
	children *intmap.Map[uint64, *context] // child contexts
}

type Type struct {
	main *context        // main context
	CR   compresult.Type // cr is the compilation result
}

// New creates a new virtual machine using memory from m and code and data from cr.
func New(m *memory.Type, cr compresult.Type) *Type {
	contexts := intmap.New[uint64, *context](minAllocContexts)
	main := context{m: m, children: contexts}
	return &Type{main: &main, CR: cr}
}

// Run executes the run loop.
// nolint:maintidx // the only thing we care about here is making it faster
func (vm *Type) Run(retResult bool) (value.Type, error) {
	ctxp := vm.main

	m := ctxp.m
	ip := ctxp.ip
	ds := vm.CR.DS
	cs := vm.CR.CS

	// temp/accumulator "register"
	tmp := value.Nil

	var err error

	freeList := list.New()

	for ip < len(*cs) {
		instr := (*cs)[ip]

		// TODO allow tracing flag
		// fmt.Printf("%8d | %8p | %v\n", ip, ctxp, instr)

		opCode := instr.OpCode()

		switch opCode {
		case bytecode.ADD, bytecode.SUB, bytecode.MUL, bytecode.DIV:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Arith(opCode, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.ADDTMP, bytecode.SUBTMP, bytecode.MULTMP, bytecode.DIVTMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			tmp, err = tmp.Arith(opCode-bytecode.ADDTMP+bytecode.ADD, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.INC:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			val, err := src0.Arith(bytecode.ADD, value.NewInt(1))
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}
			src0T := instr.Src0()

			switch src0T {
			case bytecode.AddrLcl:
				m.Set(instr.Src0Addr(), val)
			case bytecode.AddrGbl:
				name, ok := (*ds)[instr.Src0Addr()].ToString()
				if !ok {
					log.Panicf("unknown global\n %8d | %v\n", ip, instr)
				}
				m.SetGlobal(name, val)

			default:
				log.Panicf("unexpected dst in INC\n %8d | %v\n", ip, instr)
			}

		case bytecode.MOD:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Mod(src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.MODTMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			tmp, err = tmp.Mod(src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.AND, bytecode.OR:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Logic(opCode, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.ANDTMP, bytecode.ORTMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			tmp, err = tmp.Logic(opCode-bytecode.ANDTMP+bytecode.AND, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.LSH, bytecode.RSH:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Shift(opCode, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.LSHTMP, bytecode.RSHTMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			tmp, err = tmp.Shift(opCode-bytecode.LSHTMP+bytecode.LSH, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.NOT:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			val, err := src0.Not()
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

			m.Push(val)

		case bytecode.NOTTMP:
			tmp, err = tmp.Not()
			if err != nil {
				return vm.dumpStack(ctxp, ip, err)
			}

		case bytecode.FLIP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			val, err := src0.Flip()
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

			m.Push(val)

		case bytecode.FLIPTMP:
			tmp, err = tmp.Flip()
			if err != nil {
				return vm.dumpStack(ctxp, ip, err)
			}

		case bytecode.LT, bytecode.GT, bytecode.LE, bytecode.GE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Relational(opCode, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.LTTMP, bytecode.GTTMP, bytecode.LETMP, bytecode.GETMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			tmp, err = tmp.Relational(opCode-bytecode.LTTMP+bytecode.LT, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.EQ, bytecode.NE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Eq(opCode, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.EQTMP, bytecode.NETMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			tmp, err = tmp.Eq(opCode-bytecode.EQTMP+bytecode.EQ, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.LEN:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			val, err := src0.Len()
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

			m.Push(val)

		case bytecode.LENTMP:
			tmp, err = tmp.Len()
			if err != nil {
				return vm.dumpStack(ctxp, ip, err)
			}

		case bytecode.IX1:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Index(src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.IX2:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			src2 := vm.fetch(instr.Src2(), instr.Src2Addr(), m, ds)

			val, err := src2.Index(src1, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src2, src1, src0)
			}

			m.Push(val)

		case bytecode.JMP:
			src0Imm := instr.Src0Addr()

			ip += src0Imm - 1

		case bytecode.JMPF, bytecode.JMPT:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1Imm := instr.Src1Addr()

			b, ok := src0.ToBool()

			if !ok {
				return vm.dumpStack(ctxp, ip, value.ErrType, src0)
			}
			if (opCode == bytecode.JMPF && !b) || (opCode == bytecode.JMPT && b) {
				ip += src1Imm - 1
			}

		case bytecode.PUSH:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			m.Push(src0)

		case bytecode.PUSHTMP:
			m.Push(tmp)

		case bytecode.POP:
			m.Pop()

		case bytecode.MOV:
			var val value.Type
			if instr.Src0() == bytecode.AddrTmp {
				val = tmp
			} else {
				val = vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			}

			if val.IsNil() {
				return vm.dumpStack(ctxp, ip, value.ErrNil, val)
			}

			src1T := instr.Src1()

			switch src1T {
			case bytecode.AddrLcl:
				m.Set(instr.Src1Addr(), val)
			case bytecode.AddrGbl:
				name, ok := (*ds)[instr.Src1Addr()].ToString()
				if !ok {
					log.Panicf("unknown global\n %8d | %v\n", ip, instr)
				}
				m.SetGlobal(name, val)
			case bytecode.AddrTmp:
				tmp = val

			default:
				log.Panicf("unexpected dst in MOV\n %8d | %v\n", ip, instr)
			}

		case bytecode.ARR:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			ary := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			slc, ok := ary.ToArray()
			if !ok {
				log.Panicf("cannot convert value to array\n %8d | %v\n", ip, instr)
			}

			slc = slices.Clone(slc)
			slc = append(slc, val)

			val = value.NewArray(slc)
			m.Push(val)

		case bytecode.FUNC:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			frame := m.Top()
			val.SetFrame(&frame)
			m.Push(val)

		case bytecode.CALL:
			f := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			args := instr.Src1Addr()

			fVal, ok := f.ToFunction()

			if !ok {
				return vm.dumpStack(ctxp, ip, value.ErrType, f)
			}

			if fVal.ParamCnt != args {
				return vm.dumpStack(ctxp, ip, ErrArity, f)
			}

			m.PushFrame(args, fVal.LocalCnt)
			m.PushClosure(*fVal.Frame)
			m.Push(value.NewInt(ip))

			ip = fVal.Node - 1

		case bytecode.RET:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			f, ok := val.ToFunction()
			if ok && f.Frame != nil {
				frame := slices.Clone(*f.Frame)
				val.SetFrame(&frame)
			}

			nip := m.IP()
			if nip == nil {
				m.ResetSP()
				m.Push(val)
				ip = len(*cs) - 1
				break
			}
			lip, ok := nip.ToInt()
			if !ok {
				log.Panicf("can't pop instruction pointer\n %8d | %v\n", lip, instr)
			}

			m.PopFrame()
			m.PopClosure()

			m.Push(val)

			ip = lip

		case bytecode.CCONT:
			jmp := instr.Src0Addr()
			ctxID := instr.Src1Addr()
			ctxHash := hashContext(m, ctxID)
			ctxp.ip = ip + jmp - 1

			var childCtx *context
			front := freeList.Front()
			if front != nil {
				freeList.Remove(front)
				childCtx = front.Value.(*context)

				m = m.Clone(childCtx.m)
				childCtx.m = m
				childCtx.parent = ctxp
			} else {
				m = m.Clone(nil)
				childCtx = &context{m: m, parent: ctxp, children: intmap.New[uint64, *context](minAllocContexts)}
			}

			ctxp.children.Put(ctxHash, childCtx)
			ctxp = childCtx

		case bytecode.RCONT:
			lo := instr.Src0Addr()
			hi := instr.Src1Addr()
			for i := lo; i <= hi; i++ {
				hsh := hashContext(m, i)
				if child, ok := ctxp.children.Get(hsh); ok {
					deleteContext(child, freeList)
					ctxp.children.Del(hsh)
				}
			}

		case bytecode.DCONT:
			if ctxp.parent != nil {
				ctxp = ctxp.parent
				m = ctxp.m
			}

			lo := instr.Src0Addr()
			hi := instr.Src1Addr()
			for i := lo; i <= hi; i++ {
				hsh := hashContext(m, i)
				if child, ok := ctxp.children.Get(hsh); ok {
					deleteContext(child, freeList)
					ctxp.children.Del(hsh)
				}
			}

		case bytecode.SCONT:
			ctxID := instr.Src0Addr()
			ctxp.m = m
			ctxp.ip = ip
			var ok bool

			ctxp, ok = ctxp.children.Get(hashContext(m, ctxID))
			if !ok {
				log.Panicf("context not found %016x\n", hashContext(m, ctxID))
			}

			m = ctxp.m
			ip = ctxp.ip

		case bytecode.YIELD:
			tmp = vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			// otherwise naked yield, in the master context
			if ctxp.parent != nil {
				ctxp.m = m
				ctxp.ip = ip

				ctxp = ctxp.parent

				m = ctxp.m
				ip = ctxp.ip

				m.Push(tmp)
			}

		case bytecode.READ:
			b := bufio.NewReader(os.Stdin)
			line, err := b.ReadString('\n')
			if err != nil {
				return vm.dumpStack(ctxp, ip, fmt.Errorf("read error %w", err))
			}
			m.Push(value.NewString(line))

		case bytecode.WRITE:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			fmt.Print(val)
			m.Push(value.Nil)

		case bytecode.ATON:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			sv, ok := val.ToString()
			if !ok {
				return vm.dumpStack(ctxp, ip, value.ErrType, val)
			}

			if v, err := strconv.Atoi(string(sv)); err == nil {
				m.Push(value.NewInt(v))
				break
			}

			if v, err := strconv.ParseFloat(string(sv), 64); err == nil {
				m.Push(value.NewFloat(v))
				break
			}

			return vm.dumpStack(ctxp, ip, ErrConversion, val)

		case bytecode.TOA:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			val = value.NewString(val.String())
			m.Push(val)

		case bytecode.EXIT:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			i, ok := val.ToInt()
			if !ok {
				os.Exit(255)
			}
			os.Exit(i)

		default:
			log.Panicf("unknown opcode: %v\n %8d | %v\n", opCode, ip, instr)
		}

		ip++
	}

	ctxp.ip = ip

	if retResult {
		return m.Pop(), nil
	}

	return value.Type{}, nil
}

func (vm *Type) fetch(src uint64, addr int, m *memory.Type, ds *[]value.Type) value.Type {
	switch src {
	case bytecode.AddrStck:
		return m.Pop()
	case bytecode.AddrDS:
		return (*ds)[addr]
	case bytecode.AddrCls:
		return m.LookUpClosure(addr)
	case bytecode.AddrLcl:
		return m.LookUpLocal(addr)
	case bytecode.AddrGbl:
		name, ok := (*ds)[addr].ToString()
		if !ok {
			log.Panic("unknown global")
		}
		return m.LookUpGlobal(name)
	default:
		log.Panicf("unknown source")
	}
	panic("unreachable code")
}

func (vm *Type) dumpStack(ctx *context, ip int, err error, values ...value.Type) (value.Type, error) {
	fmt.Printf("RUNTIME ERROR : %v\n", err)

	args := ""
	sep := ""
	for _, v := range values {
		args += sep + v.Abbrev()
		sep = ", "
	}

	start, end := max(0, ip-3), min(len(*vm.CR.CS), ip+3)

	for i, v := range (*vm.CR.CS)[start:end] {
		if i+start == ip {
			fmt.Printf("--> %d: %v; %s\n", i+start, v, args)
		} else {
			fmt.Printf("    %d: %v\n", i+start, v)
		}
	}

	for ; ctx != nil; ctx = ctx.parent {
		fmt.Printf("memory context %08p\n", ctx)
		m := ctx.m
		m.DumpStack(vm.CR.Dbg)
	}
	// reset state for the next run
	vm.main.m.Reset()
	vm.main.ip = len(*vm.CR.CS)
	vm.main.children.Clear()

	return value.Nil, err
}

func hashContext(m *memory.Type, id int) uint64 {
	return (uint64(m.CallDepth()) << 15) ^ uint64(id)
}

func deleteContext(ctxp *context, freeList *list.List) {
	ctxp.children.ForEach(func(_ uint64, child *context) bool {
		deleteContext(child, freeList)
		return true
	})

	ctxp.children.Clear()

	freeList.PushFront(ctxp)
}

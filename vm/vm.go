// Package vm implements the virtual machine.
package vm

import (
	"bufio"
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
)

var (
	ErrConversion = errors.New("conversion error")
	ErrArity      = errors.New("arity mismatch")
)

type context struct {
	ip       int          // instruction pointer
	m        *memory.Type // variables
	parent   *context     // parent context
	children []*context   // child contexts
}

type Type struct {
	ctx *context        // ctx is the context tree
	CR  compresult.Type // cr is the compilation result
}

// New creates a new virtual machine using memory from m and code and data from cr.
func New(m *memory.Type, cr compresult.Type) *Type {
	return &Type{ctx: &context{m: m, children: []*context{}}, CR: cr}
}

// Run executes the run loop.
// nolint:maintidx // the only thing we care about here is making it faster
func (vm *Type) Run(retResult bool) (value.Type, error) {
	ctxp := vm.ctx
	m := ctxp.m
	ip := ctxp.ip
	ds := vm.CR.DS
	cs := vm.CR.CS
	tmp := value.Nil
	var err error

	for ip < len(*cs) {
		instr := (*cs)[ip]

		// TODO allow tracing flag
		fmt.Printf("%8d | %8p | %v\n", ip, m, instr)

		opCode := instr.OpCode()

		switch opCode {
		case bytecode.ADD, bytecode.SUB, bytecode.MUL, bytecode.DIV:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"+", "-", "*", "/"}[opCode-bytecode.ADD]

			val, err := src1.Arith(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.ADDTMP, bytecode.SUBTMP, bytecode.MULTMP, bytecode.DIVTMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			op := [...]string{"+", "-", "*", "/"}[opCode-bytecode.ADDTMP]

			tmp, err = tmp.Arith(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.INC:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			val, err := src0.Arith("+", value.NewInt(1))
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
			op := [...]string{"&", "|"}[opCode-bytecode.AND]

			val, err := src1.Logic(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.ANDTMP, bytecode.ORTMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			op := [...]string{"&", "|"}[opCode-bytecode.ANDTMP]

			tmp, err = tmp.Logic(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.LSH, bytecode.RSH:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"<<", ">>"}[opCode-bytecode.LSH]

			val, err := src1.Shift(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.LSHTMP, bytecode.RSHTMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			op := [...]string{"<<", ">>"}[opCode-bytecode.LSHTMP]

			tmp, err = tmp.Shift(op, src0)
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
			op := [...]string{"<", ">", "<=", ">="}[opCode-bytecode.LT]

			val, err := src1.Relational(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.LTTMP, bytecode.GTTMP, bytecode.LETMP, bytecode.GETMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			op := [...]string{"<", ">", "<=", ">="}[opCode-bytecode.LTTMP]

			tmp, err = tmp.Relational(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src0)
			}

		case bytecode.EQ, bytecode.NE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"==", "!="}[opCode-bytecode.EQ]

			val, err := src1.Eq(op, src0)
			if err != nil {
				return vm.dumpStack(ctxp, ip, err, src1, src0)
			}

			m.Push(val)

		case bytecode.EQTMP, bytecode.NETMP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			op := [...]string{"==", "!="}[opCode-bytecode.EQTMP]

			tmp, err = tmp.Eq(op, src0)
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

		case bytecode.JMPF:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1Imm := instr.Src1Addr()

			b, ok := src0.ToBool()

			if !ok {
				return vm.dumpStack(ctxp, ip, value.ErrType, src0)
			}
			if !b {
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

			slc = append(slc, val)

			val = value.NewArray(slc)
			m.Push(val)

		case bytecode.FUNC:
			localCnt := instr.Src2Addr()
			paramCnt := instr.Src1Addr()

			f := value.NewFunction(instr.Src0Addr(), m.Top(), paramCnt, localCnt)
			m.Push(f)

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
			m.PushClosure(fVal.Frame.(memory.Frame))
			m.Push(value.NewInt(ip))

			ip = fVal.Node - 1

		case bytecode.RET:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			f, ok := val.ToFunction()
			if ok && f.Frame != nil {
				f.Frame = slices.Clone(f.Frame.(memory.Frame))
				val = value.NewFunction(f.Node, f.Frame, f.ParamCnt, f.LocalCnt)
			}

			nip := m.IP()
			if nip == nil {
				m.ResetSP()
				m.Push(val)
				ip = len(*cs)
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
			ctxp.ip = ip + jmp - 1

			m = m.Clone()
			childCtx := context{m: m, parent: ctxp, children: make([]*context, 0)}
			ctxp.children = append(ctxp.children, &childCtx)
			ctxp = &childCtx

		case bytecode.RCONT:
			ctxNum := instr.Src0Addr()
			ctxp.children = ctxp.children[:len(ctxp.children)-ctxNum]

		case bytecode.DCONT:
			ctxNum := instr.Src0Addr()
			ctxp = ctxp.parent
			if len(ctxp.children) > 0 {
				ctxp.children = ctxp.children[:len(ctxp.children)-ctxNum]
			}
			m = ctxp.m

		case bytecode.SCONT:
			ctxID := instr.Src0Addr()
			ctxp.m = m
			ctxp.ip = ip

			ctxp = ctxp.children[len(ctxp.children)-ctxID]

			m = ctxp.m
			ip = ctxp.ip

		case bytecode.YIELD:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			// yield needs to push in the slave context because a subsequent pop, we
			// should optimise this away
			m.Push(val)

			// otherwise naked yield, in the master context
			if ctxp.parent != nil {
				ctxp.m = m
				ctxp.ip = ip

				ctxp = ctxp.parent

				m = ctxp.m
				ip = ctxp.ip

				m.Push(val)
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
			val = value.NewString(fmt.Sprint(val))
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

func (vm Type) fetch(src uint64, addr int, m *memory.Type, ds *[]value.Type) value.Type {
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
		// reset state for the next run
		m.Reset()
		ctx.children = []*context{}
		ctx.ip = len(*vm.CR.CS)
		vm.ctx = ctx
	}

	return value.Nil, err
}

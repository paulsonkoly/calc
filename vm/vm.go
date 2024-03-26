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
	ctx *context         // ctx is the context tree
	CS  *[]bytecode.Type // CS is the code segment
	DS  *[]value.Type    // DS is the data segment
}

func New(m *memory.Type, cs *[]bytecode.Type, ds *[]value.Type) *Type {
	return &Type{ctx: &context{m: m, children: []*context{}}, CS: cs, DS: ds}
}

// nolint:maintidx // the only thing we care about here is making it faster
func (vm *Type) Run(retResult bool) (value.Type, error) {
	ctxp := vm.ctx
	m := ctxp.m
	ip := ctxp.ip
	ds := vm.DS
	cs := vm.CS

	for ip < len(*cs) {
		instr := (*cs)[ip]

		// TODO allow tracing flag
		// fmt.Printf("%8d | %8p | %v\n", ip, m, instr)

		opCode := instr.OpCode()

		switch opCode {
		case bytecode.ADD, bytecode.SUB, bytecode.MUL, bytecode.DIV:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"+", "-", "*", "/"}[opCode-bytecode.ADD]

			val, err := src1.Arith(op, src0)
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.MOD:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Mod(src0)
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.AND, bytecode.OR:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"&", "|"}[opCode-bytecode.AND]

			val, err := src1.Logic(op, src0)
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.LSH, bytecode.RSH:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"<<", ">>"}[opCode-bytecode.LSH]

			val, err := src1.Shift(op, src0)
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.NOT:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			val, err := src0.Not()
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.FLIP:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			val, err := src0.Flip()
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.LT, bytecode.GT, bytecode.LE, bytecode.GE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"<", ">", "<=", ">="}[opCode-bytecode.LT]

			val, err := src1.Relational(op, src0)
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.EQ, bytecode.NE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := [...]string{"==", "!="}[opCode-bytecode.EQ]

			val, err := src1.Eq(op, src0)
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.LEN:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			val, err := src0.Len()
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.IX1:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val, err := src1.Index(src0)
			if err != nil {
				return dumpStack(m, err)
			}

			m.Push(val)

		case bytecode.IX2:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			src2 := vm.fetch(instr.Src2(), instr.Src2Addr(), m, ds)

			val, err := src2.Index(src1, src0)
			if err != nil {
				return dumpStack(m, err)
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
				return dumpStack(m, value.ErrType)
			}
			if !b {
				ip += src1Imm - 1
			}

		case bytecode.PUSH:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			m.Push(src0)

		case bytecode.POP:
			m.Pop()

		case bytecode.MOV:
			val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

			if val.IsNil() {
				return dumpStack(m, value.ErrNil)
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
				return dumpStack(m, value.ErrType)
			}

			if fVal.ParamCnt != args {
				return dumpStack(m, ErrArity)
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
			ctxp.children = ctxp.children[:len(ctxp.children)-1]

		case bytecode.DCONT:
			ctxp = ctxp.parent
			if len(ctxp.children) > 0 {
				ctxp.children = ctxp.children[:len(ctxp.children)-1]
			}
			m = ctxp.m

		case bytecode.SCONT:
			ctxp.m = m
			ctxp.ip = ip

			ctxp = ctxp.children[len(ctxp.children)-1]

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
				return dumpStack(m, fmt.Errorf("read error %w", err))
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
				return dumpStack(m, value.ErrType)
			}

			if v, err := strconv.Atoi(string(sv)); err == nil {
				m.Push(value.NewInt(v))
				break
			}

			if v, err := strconv.ParseFloat(string(sv), 64); err == nil {
				m.Push(value.NewFloat(v))
				break
			}

			return dumpStack(m, ErrConversion)

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

func dumpStack(_ *memory.Type, err error) (value.Type, error) {
	fmt.Println(err)
	return value.Type{}, err
}

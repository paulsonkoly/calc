package vm

import (
	"log"
	"slices"

	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/value"
)

type Type struct {
	ip int
}

func NewVM() *Type {
	return &Type{}
}

func (vm *Type) Run(m *memory.Type, cs *[]bytecode.Type, ds []value.Type) value.Type {
	var val value.Type

	for vm.ip < len(*cs) {
		instr := (*cs)[vm.ip]

		opCode := instr.OpCode()

		switch opCode {
		case bytecode.ADD, bytecode.SUB, bytecode.MUL, bytecode.DIV:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := map[bytecode.OpCode]string{bytecode.ADD: "+", bytecode.SUB: "-", bytecode.MUL: "*", bytecode.DIV: "/"}[opCode]

			val = src1.Arith(op, src0)

			m.Push(val)

    case bytecode.MOD:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val = src1.Mod(src0)

			m.Push(val)

    case bytecode.AND, bytecode.OR:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := map[bytecode.OpCode]string{bytecode.AND: "&", bytecode.OR: "|"}[opCode]

			val = src1.Logic(op, src0)

			m.Push(val)

		case bytecode.LT, bytecode.GT, bytecode.LE, bytecode.GE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := map[bytecode.OpCode]string{bytecode.LT: "<", bytecode.GT: ">", bytecode.LE: "<=", bytecode.GE: ">="}[opCode]

			val = src1.Relational(op, src0)

			m.Push(val)

		case bytecode.EQ, bytecode.NE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)
			op := map[bytecode.OpCode]string{bytecode.EQ: "==", bytecode.NE: "!="}[opCode]

			val = src1.Eq(op, src0)

			m.Push(val)

		case bytecode.IX1:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			val = src1.Index(src0)

			m.Push(val)

		case bytecode.JMP:
			src0Imm := instr.Src0Addr()

			vm.ip += src0Imm - 1

		case bytecode.JMPF:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1Imm := instr.Src1Addr()
			// TODO - how do we do type errors?
			if b, ok := src0.ToBool(); !b && ok {
				vm.ip += src1Imm - 1
			}

		case bytecode.PUSH:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			m.Push(src0)

		case bytecode.POP:
			m.Pop()

		case bytecode.MOV:
			val = vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			src1T := instr.Src1()

			switch src1T {
			case bytecode.ADDR_LCL:
				m.Set(instr.Src1Addr(), val)
			case bytecode.ADDR_GBL:
				name, ok := ds[instr.Src1Addr()].ToString()
				if !ok {
					panic("unknown global")
				}
				m.SetGlobal(name, val)

			default:
				panic("unexpected dst in MOV")
			}

		case bytecode.ARR:
			val = vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
			ary := vm.fetch(instr.Src1(), instr.Src1Addr(), m, ds)

			slc, ok := ary.ToArray()
			if !ok {
				panic("cannot convert value to array")
			}

			slc = append(slc, val)

			val = value.NewArray(slc)
			m.Push(val)

		case bytecode.FUNC:
			f := value.NewFunction(instr.Src0Addr(), slices.Clone(m.Top()), instr.Src1Addr())
			m.Push(f)

    case bytecode.CALL:
      f := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)
      args := instr.Src1Addr()

      m.PushFrame(args)

      fVal, ok := f.ToFunction()
      if !ok {
        m.PopFrame()
        m.Push(value.TypeError)
        break
      }
      if fVal.ParamCnt != args {
        m.PopFrame()
        m.Push(value.ArgumentError)
        break
      }

      m.PushClosure(fVal.Frame.(memory.Frame))
      m.Push(value.NewInt(vm.ip))

      vm.ip = fVal.Node - 1

    case bytecode.RET:
      ip, ok := m.IP().ToInt()
      if !ok {
        panic("can't pop instruction pointer")
      }

      val := vm.fetch(instr.Src0(), instr.Src0Addr(), m, ds)

      m.PopFrame()
      m.PopClosure()

      m.Push(val)

      vm.ip = ip
		default:
			log.Panicf("unknown opcode: %v", opCode)
		}

		vm.ip++
	}

	return val
}

func (vm Type) fetch(src uint64, addr int, m *memory.Type, ds []value.Type) value.Type {
	switch src {
	case bytecode.ADDR_STCK:
		return m.Pop()
	case bytecode.ADDR_DS:
		return ds[addr]
	case bytecode.ADDR_CLS:
		return m.LookUpClosure(addr)
	case bytecode.ADDR_LCL:
		return m.LookUpLocal(addr)
	case bytecode.ADDR_GBL:
		name, ok := ds[addr].ToString()
		if !ok {
			panic("unknown global")
		}
		return m.LookUpGlobal(name)
	default:
		panic("unknown source")
	}
}

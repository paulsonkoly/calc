package vm

import (
	"fmt"
	"log"
	"slices"
	"strconv"

	"github.com/paulsonkoly/calc/memory"
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/value"
)

type Type struct {
	ip int             // instruction pointer
	m  *memory.Type    // variables
	cs []bytecode.Type // CS is the code segment
	ds []value.Type    // DS is the data segment
}

func New(m *memory.Type, cs []bytecode.Type, ds []value.Type) *Type {
	return &Type{m: m, cs: cs, ds: ds}
}

func (vm *Type) SetSegments(cs []bytecode.Type, ds []value.Type) {
	vm.cs = cs
	vm.ds = ds
}

func (vm *Type) Run(retResult bool) value.Type {
	for vm.ip < len(vm.cs) {
		instr := vm.cs[vm.ip]

		// TODO allow tracing flag
		// fmt.Printf("%8d | %v\n", vm.ip, instr)

		opCode := instr.OpCode()

		switch opCode {
		case bytecode.ADD, bytecode.SUB, bytecode.MUL, bytecode.DIV:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr())
			op := [...]string{"+", "-", "*", "/"}[opCode-bytecode.ADD]

      val := src1.Arith(op, src0)

			vm.m.Push(val)

		case bytecode.MOD:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr())

      val := src1.Mod(src0)

			vm.m.Push(val)

		case bytecode.AND, bytecode.OR:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr())
			op := [...]string{"&", "|"}[opCode-bytecode.AND]

      val := src1.Logic(op, src0)

			vm.m.Push(val)

		case bytecode.NOT:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
      val := src0.Not()
			vm.m.Push(val)

		case bytecode.LT, bytecode.GT, bytecode.LE, bytecode.GE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr())
			op := [...]string{"<", ">", "<=", ">="}[opCode-bytecode.LT]

      val := src1.Relational(op, src0)

			vm.m.Push(val)

		case bytecode.EQ, bytecode.NE:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr())
			op := [...]string{"==", "!="}[opCode-bytecode.EQ]

      val := src1.Eq(op, src0)

			vm.m.Push(val)

		case bytecode.LEN:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())

			vm.m.Push(src0.Len())

		case bytecode.IX1:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr())

      val := src1.Index(src0)

			vm.m.Push(val)

		case bytecode.IX2:
			ary := vm.m.Pop()
			src1 := vm.fetch(instr.Src1(), instr.Src1Addr())
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())

      val := ary.Index(src1, src0)

			vm.m.Push(val)

		case bytecode.JMP:
			src0Imm := instr.Src0Addr()

			vm.ip += src0Imm - 1

		case bytecode.JMPF:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1Imm := instr.Src1Addr()
			// TODO - how do we do type errors?
			if b, ok := src0.ToBool(); !b && ok {
				vm.ip += src1Imm - 1
			}

		case bytecode.PUSH:
			src0 := vm.fetch(instr.Src0(), instr.Src0Addr())
			vm.m.Push(src0)

		case bytecode.POP:
			vm.m.Pop()

		case bytecode.MOV:
      val := vm.fetch(instr.Src0(), instr.Src0Addr())
			src1T := instr.Src1()

			switch src1T {
			case bytecode.ADDR_LCL:
				vm.m.Set(instr.Src1Addr(), val)
			case bytecode.ADDR_GBL:
				name, ok := vm.ds[instr.Src1Addr()].ToString()
				if !ok {
					log.Panicf("unknown global\n %8d | %v\n", vm.ip, instr)
				}
				vm.m.SetGlobal(name, val)

			default:
				log.Panicf("unexpected dst in MOV\n %8d | %v\n", vm.ip, instr)
			}

		case bytecode.ARR:
      val := vm.fetch(instr.Src0(), instr.Src0Addr())
			ary := vm.fetch(instr.Src1(), instr.Src1Addr())

			slc, ok := ary.ToArray()
			if !ok {
				log.Panicf("cannot convert value to array\n %8d | %v\n", vm.ip, instr)
			}

			slc = append(slc, val)

			val = value.NewArray(slc)
			vm.m.Push(val)

		case bytecode.FUNC:
			funInfo := instr.Src1Addr()
			localCnt := funInfo >> 12
			paramCnt := funInfo & 0xfff
			f := value.NewFunction(instr.Src0Addr(), slices.Clone(vm.m.Top()), paramCnt, localCnt)
			vm.m.Push(f)

		case bytecode.CALL:
			f := vm.fetch(instr.Src0(), instr.Src0Addr())
			args := instr.Src1Addr()

			fVal, ok := f.ToFunction()

			if !ok {
				vm.m.Push(value.TypeError)
				break
			}

			if fVal.ParamCnt != args {
				vm.m.Push(value.ArgumentError)
				break
			}

			vm.m.PushFrame(args, fVal.LocalCnt)
			vm.m.PushClosure(fVal.Frame.(memory.Frame))
			vm.m.Push(value.NewInt(vm.ip))

			vm.ip = fVal.Node - 1

		case bytecode.RET:
			ip, ok := vm.m.IP().ToInt()
			if !ok {
				log.Panicf("can't pop instruction pointer\n %8d | %v\n", vm.ip, instr)
			}

			val := vm.fetch(instr.Src0(), instr.Src0Addr())

			vm.m.PopFrame()
			vm.m.PopClosure()

			vm.m.Push(val)

			vm.ip = ip

		case bytecode.WRITE:
			val := vm.fetch(instr.Src0(), instr.Src0Addr())
			fmt.Println(val)
			vm.m.Push(value.NoResultError)

		case bytecode.ATON:
			val := vm.fetch(instr.Src0(), instr.Src0Addr())

			sv, ok := val.ToString()
			if !ok {
				vm.m.Push(value.TypeError)
				break
			}

			if v, err := strconv.Atoi(string(sv)); err == nil {
				vm.m.Push(value.NewInt(v))
				break
			}

			if v, err := strconv.ParseFloat(string(sv), 64); err == nil {
				vm.m.Push(value.NewFloat(v))
				break
			}

			vm.m.Push(value.ConversionError)

		case bytecode.TOA:
      val := vm.fetch(instr.Src0(), instr.Src0Addr())
			val = value.NewString(fmt.Sprint(val))
			vm.m.Push(val)

		default:
			log.Panicf("unknown opcode: %v\n %8d | %v\n", opCode, vm.ip, instr)
		}

		vm.ip++
	}

  if retResult {
    return vm.m.Pop()
  }

	return value.Type{}
}

func (vm Type) fetch(src uint64, addr int) value.Type {
	switch src {
	case bytecode.ADDR_STCK:
		return vm.m.Pop()
	case bytecode.ADDR_DS:
		return vm.ds[addr]
	case bytecode.ADDR_CLS:
		return vm.m.LookUpClosure(addr)
	case bytecode.ADDR_LCL:
		return vm.m.LookUpLocal(addr)
	case bytecode.ADDR_GBL:
		name, ok := vm.ds[addr].ToString()
		if !ok {
			log.Panicf("unknown global\n %8d | %v\n", vm.ip, vm.cs[vm.ip])
		}
		return vm.m.LookUpGlobal(name)
	default:
		log.Panicf("unknown source\n %8d | %v\n", vm.ip, vm.cs[vm.ip])
	}
	panic("unreachable code")
}

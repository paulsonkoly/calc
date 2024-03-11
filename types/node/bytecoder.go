package node

import (
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/value"
)

type ByteCoder interface {
	byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type
}

func ByteCode(bc ByteCoder, cs *[]bytecode.Type, ds *[]value.Type) {
	instr := bc.byteCode(0, cs, ds)
	if instr.Src0() != bytecode.ADDR_STCK { // leave the final result on the stack
		instr |= bytecode.NewByteCode(bytecode.PUSH, 0, 0, 0, 0)
		*cs = append(*cs, instr)
	}
}

func (i Int) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewInt(int(i))
	ix := len(*ds)
	*ds = append(*ds, v)

	return encodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (b Bool) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewBool(bool(b))
	ix := len(*ds)
	*ds = append(*ds, v)

	return encodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (f Float) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewFloat(float64(f))
	ix := len(*ds)
	*ds = append(*ds, v)

	return encodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (s String) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewString(string(s))
	ix := len(*ds)
	*ds = append(*ds, v)

	return encodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (l List) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// TODO optimise at least the constant parts of array case to the ds
  v := value.NewArray([]value.Type{})
  ix := len(*ds)
  *ds = append(*ds, v)
  if len(l.Elems) == 0 {
    return encodeSrc(srcsel, bytecode.ADDR_DS, ix)
  }

  instr := l.Elems[0].byteCode(0, cs, ds)
  instr |= bytecode.NewByteCode(bytecode.ARR, bytecode.ADDR_DS, ix, 0, 0)
	*cs = append(*cs, instr)
  
  for _, t := range l.Elems[1:] {
		instr = t.byteCode(0, cs, ds)
		instr |= bytecode.NewByteCode(bytecode.ARR, bytecode.ADDR_STCK, 0, 0, 0)
		*cs = append(*cs, instr)
  }

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (l Local) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	return encodeSrc(srcsel, bytecode.ADDR_LCL, int(l))
}

func (c Closure) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	return encodeSrc(srcsel, bytecode.ADDR_CLS, int(c))
}

func (n Name) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewString(string(n))
	ix := len(*ds)
	*ds = append(*ds, v)
	return encodeSrc(srcsel, bytecode.ADDR_GBL, ix)
}

func (f Function) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.NewByteCode(bytecode.JMP, 0, 0, 0, 0)
	*cs = append(*cs, instr)
	jmpAddr := len(*cs) - 1

	instr = f.Body.byteCode(0, cs, ds)
	if instr.Src0() != bytecode.ADDR_STCK {
		instr |= bytecode.NewByteCode(bytecode.PUSH, 0, 0, 0, 0)
		*cs = append(*cs, instr)
	}

	instr = bytecode.NewByteCode(bytecode.FUNC, 0, 0, bytecode.ADDR_IMM, jmpAddr+1)
	*cs = append(*cs, instr)

	// patch the jmp
	(*cs)[jmpAddr] |= encodeSrc(0, bytecode.ADDR_IMM, len(*cs)-jmpAddr-1)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (a Assign) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	srcInstr := a.Value.byteCode(0, cs, ds)
	instr := srcInstr | a.VarRef.byteCode(1, cs, ds)
	instr |= bytecode.NewByteCode(bytecode.MOV, 0, 0, 0, 0)

	*cs = append(*cs, instr)

	return encodeSrc(srcsel, instr.Src1(), instr.Src1Addr())
}

func encodeSrc(srcsel int, src uint64, srcAddr int) bytecode.Type {
	switch srcsel {
	case 0:
		// TODO check for overflow
		return bytecode.NewByteCode(0, 0, 0, src, srcAddr)
	case 1:
		return bytecode.NewByteCode(0, src, srcAddr, 0, 0)
	default:
		panic("wrong srcsel")
	}
}

func (b BinOp) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	var op bytecode.OpCode
	switch b.Op {
	case "+":
		op = bytecode.ADD
	case "-":
		op = bytecode.SUB
	case "*":
		op = bytecode.MUL
	case "/":
		op = bytecode.DIV
	case "%":
		op = bytecode.MOD
	case "&":
		op = bytecode.AND
	case "|":
		op = bytecode.OR
	case "==":
		op = bytecode.EQ
	case "!=":
		op = bytecode.NE
	case "<":
		op = bytecode.LT
	case "<=":
		op = bytecode.LE
	case ">":
		op = bytecode.GT
	case ">=":
		op = bytecode.GE
	default:
		panic("unexpected op")
	}

	instr := b.Left.byteCode(1, cs, ds)
	instr |= b.Right.byteCode(0, cs, ds)
	instr |= bytecode.NewByteCode(op, 0, 0, 0, 0)

	*cs = append(*cs, instr)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (u UnOp) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	var op bytecode.OpCode
	switch u.Op {
	case "-":
		return BinOp{Op: "*", Left: Int(-1), Right: u.Target}.byteCode(srcsel, cs, ds)
	case "#":
		op = bytecode.LEN
	case "!":
		op = bytecode.NOT
	default:
		panic("unexpected op")
	}

	instr := u.Target.byteCode(0, cs, ds)
	instr |= bytecode.NewByteCode(op, 0, 0, 0, 0)

	*cs = append(*cs, instr)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (b Block) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	var instr bytecode.Type
	for i, t := range b.Body {
		instr = t.byteCode(srcsel, cs, ds)
		if i != len(b.Body)-1 { // throw away mid-block results
			if instr.Src(srcsel) == bytecode.ADDR_STCK {
				instr = bytecode.NewByteCode(bytecode.POP, 0, 0, 0, 0)
				*cs = append(*cs, instr)
			}
		}
	}
	return instr
}

func (i If) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	//
	// JMPF +truecasesize+1, condition         --+
	// truecase                                  |
	// if trucase result is not on the stack     |
	//    PUSH truecase result                   |
	// JMP +2                                  --|-+
	// PUSH no result                          <-+ |
	//                                         <---+
	//
	instr := i.Condition.byteCode(0, cs, ds)
	instr |= bytecode.NewByteCode(bytecode.JMPF, 0, 0, 0, 0)

	*cs = append(*cs, instr)
	jmpfAddr := len(*cs) - 1

	tcInstr := i.TrueCase.byteCode(0, cs, ds)
	if tcInstr.Src0() != bytecode.ADDR_STCK {
		instr = bytecode.NewByteCode(bytecode.PUSH, 0, 0, 0, 0)
		instr |= tcInstr
		*cs = append(*cs, instr)
	}

	instr = bytecode.NewByteCode(bytecode.JMP, 0, 0, bytecode.ADDR_IMM, 2)
	*cs = append(*cs, instr)
	// put no result error on DS
	ix := len(*ds)
	*ds = append(*ds, value.NoResultError)
	instr = bytecode.NewByteCode(bytecode.PUSH, 0, 0, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	// patch the JMPF
	(*cs)[jmpfAddr] |= encodeSrc(1, bytecode.ADDR_IMM, len(*cs)-jmpfAddr-1)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (i IfElse) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := i.Condition.byteCode(0, cs, ds)
	instr |= bytecode.NewByteCode(bytecode.JMPF, 0, 0, 0, 0)
	*cs = append(*cs, instr)
	jmpfAddr := len(*cs) - 1

	instr = i.TrueCase.byteCode(0, cs, ds)
	if instr.Src0() != bytecode.ADDR_STCK {
		instr |= bytecode.NewByteCode(bytecode.PUSH, 0, 0, 0, 0)
		*cs = append(*cs, instr)
	}

	instr = bytecode.NewByteCode(bytecode.JMP, 0, 0, 0, 0)
	*cs = append(*cs, instr)
	jmpAddr := len(*cs) - 1

	instr = i.FalseCase.byteCode(0, cs, ds)
	if instr.Src0() != bytecode.ADDR_STCK {
		instr |= bytecode.NewByteCode(bytecode.PUSH, 0, 0, 0, 0)
		*cs = append(*cs, instr)
	}

	// atch jmpf
	(*cs)[jmpfAddr] |= encodeSrc(1, bytecode.ADDR_IMM, jmpAddr-jmpfAddr+1)
	// patch jmp
	(*cs)[jmpAddr] |= encodeSrc(0, bytecode.ADDR_IMM, len(*cs)-jmpAddr)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (w While) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// PUSH no result
	// JUMPF condition                               <-----+ -+
	// POP                                                 |  |
	// loop body                                           |  |
	// if body didn't leave its result on the stack        |  |
	//    PUSH body result                                 |  |
	// JUMP                                           -----+  |
	//                                                <-------+

	ix := len(*ds)
	*ds = append(*ds, value.NoResultError)
	instr := bytecode.NewByteCode(bytecode.PUSH, 0, 0, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	instr = w.Condition.byteCode(0, cs, ds)
	condAddr := len(*cs) - 1
	instr |= bytecode.NewByteCode(bytecode.JMPF, 0, 0, 0, 0)
	*cs = append(*cs, instr)

	jmpfAddr := len(*cs) - 1

	instr = bytecode.NewByteCode(bytecode.POP, 0, 0, 0, 0)
	*cs = append(*cs, instr)

	instr = w.Body.byteCode(0, cs, ds)

	if instr.Src0() != bytecode.ADDR_STCK {
		instr |= bytecode.NewByteCode(bytecode.PUSH, 0, 0, 0, 0)
		*cs = append(*cs, instr)
	}

	instr = bytecode.NewByteCode(bytecode.JMP, 0, 0, bytecode.ADDR_IMM, condAddr-len(*cs))
	*cs = append(*cs, instr)

	// patch the JMPF
	(*cs)[jmpfAddr] |= encodeSrc(1, bytecode.ADDR_IMM, len(*cs)-jmpfAddr)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (i IndexAt) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := i.Ary.byteCode(1, cs, ds)
	instr |= i.At.byteCode(0, cs, ds)
	instr |= bytecode.NewByteCode(bytecode.IX1, 0, 0, 0, 0)

	*cs = append(*cs, instr)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (u IndexFromTo) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	return 0
}

func (r Read) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.NewByteCode(bytecode.READ, 0, 0, 0, 0)

	*cs = append(*cs, instr)

	return encodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (w Write) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := w.Value.byteCode(0, cs, ds)
	instr |= bytecode.NewByteCode(bytecode.WRITE, 0, 0, 0, 0)

	*cs = append(*cs, instr)

	ix := len(*ds)
	*ds = append(*ds, value.NoResultError)

	return encodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (r Return) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type { return 0 }
func (a Aton) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type   { return 0 }
func (t Toa) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type    { return 0 }
func (e Error) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type  { return 0 }
func (e Exit) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type   { return 0 }

func (c Call) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type { return 0 }

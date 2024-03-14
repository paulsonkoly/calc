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
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}
}

func ByteCodeNoStck(bc ByteCoder, cs *[]bytecode.Type, ds *[]value.Type) {
	instr := bc.byteCode(0, cs, ds)
	if instr.Src0() == bytecode.ADDR_STCK { // don't leave the final result on the stack
		instr = bytecode.New(bytecode.POP)
		*cs = append(*cs, instr)
	}
}

func (i Int) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewInt(int(i))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (b Bool) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewBool(bool(b))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (f Float) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewFloat(float64(f))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (s String) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewString(string(s))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_DS, ix)
}

func (l List) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// TODO optimise at least the constant parts of array case to the ds
	v := value.NewArray([]value.Type{})
	ix := len(*ds)
	*ds = append(*ds, v)
	if len(l.Elems) == 0 {
		return bytecode.EncodeSrc(srcsel, bytecode.ADDR_DS, ix)
	}

	instr := l.Elems[0].byteCode(0, cs, ds)
	instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	for _, t := range l.Elems[1:] {
		instr = t.byteCode(0, cs, ds)
		instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.ADDR_STCK, 0)
		*cs = append(*cs, instr)
	}

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (l Local) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_LCL, int(l))
}

func (c Closure) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_CLS, int(c))
}

func (n Name) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewString(string(n))
	ix := len(*ds)
	*ds = append(*ds, v)
	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_GBL, ix)
}

func (f Function) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.JMP)
	*cs = append(*cs, instr)
	jmpAddr := len(*cs) - 1

	instr = f.Body.byteCode(0, cs, ds)
	instr |= bytecode.New(bytecode.RET)
	*cs = append(*cs, instr)

	instr = bytecode.New(bytecode.FUNC) |
		bytecode.EncodeSrc(2, bytecode.ADDR_IMM, (f.LocalCnt)) |
		bytecode.EncodeSrc(1, bytecode.ADDR_IMM, len(f.Parameters.Elems)) |
		bytecode.EncodeSrc(0, bytecode.ADDR_IMM, jmpAddr+1)
	*cs = append(*cs, instr)

	// patch the jmp
	(*cs)[jmpAddr] |= bytecode.EncodeSrc(0, bytecode.ADDR_IMM, len(*cs)-jmpAddr-1)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (c Call) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// push the arguments
	for _, arg := range c.Arguments.Elems {
		instr := arg.byteCode(0, cs, ds)
		if instr.Src0() != bytecode.ADDR_STCK {
			instr |= bytecode.New(bytecode.PUSH)
			*cs = append(*cs, instr)
		}
	}

	// get the function
	instr := c.Name.byteCode(0, cs, ds)

	instr |= bytecode.New(bytecode.CALL) | bytecode.EncodeSrc(1, bytecode.ADDR_IMM, len(c.Arguments.Elems))
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (r Return) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := r.Target.byteCode(0, cs, ds)
	instr |= bytecode.New(bytecode.RET)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (y Yield) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := y.Target.byteCode(0, cs, ds)
	instr |= bytecode.New(bytecode.YIELD)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (a Assign) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	srcInstr := a.Value.byteCode(0, cs, ds)
	instr := srcInstr | a.VarRef.byteCode(1, cs, ds)
	instr |= bytecode.New(bytecode.MOV)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, instr.Src1(), instr.Src1Addr())
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
	instr |= bytecode.New(op)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
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

	instr := bytecode.New(op) | u.Target.byteCode(0, cs, ds)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (b Block) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	var instr bytecode.Type
	for i, t := range b.Body {
		instr = t.byteCode(srcsel, cs, ds)
		if i != len(b.Body)-1 { // throw away mid-block results
			if instr.Src(srcsel) == bytecode.ADDR_STCK {
				instr = bytecode.New(bytecode.POP)
				*cs = append(*cs, instr)
			}
		}
	}
	return instr
}

func (i If) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	//
	// JMPF ..., condition                     --+
	// truecase                                  |
	// if trucase result is not on the stack     |
	//    PUSH truecase result                   |
	// JMP +4                                  --|-+
	// PUSH no result                          <-+ |
	// JMP +2                                    | |
	// PUSH type error                         <-+ |
	//                                         <---+
	//
	instr := bytecode.New(bytecode.JMPF) | i.Condition.byteCode(0, cs, ds)

	*cs = append(*cs, instr)
	jmpfAddr := len(*cs) - 1

	tcInstr := i.TrueCase.byteCode(0, cs, ds)
	if tcInstr.Src0() != bytecode.ADDR_STCK {
		instr = bytecode.New(bytecode.PUSH) | tcInstr
		*cs = append(*cs, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.ADDR_IMM, 4)
	*cs = append(*cs, instr)

	ix := len(*ds)
	*ds = append(*ds, value.NoResultError)
	instr = bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.ADDR_IMM, 2)
	*cs = append(*cs, instr)

	ix = len(*ds)
	*ds = append(*ds, value.TypeError)
	instr = bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	// patch the JMPF
	(*cs)[jmpfAddr] |=
		bytecode.EncodeSrc(2, bytecode.ADDR_IMM, len(*cs)-jmpfAddr-1) |
			bytecode.EncodeSrc(1, bytecode.ADDR_IMM, len(*cs)-jmpfAddr-3)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (i IfElse) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	//
	// JMPF ..., condition                     --+
	// truecase                                  |
	// if trucase result is not on the stack     |
	//    PUSH truecase result                   |
	// JMP                                     --|-+
	// PUSH type error                         <-+ |
	// falsecase                               <-+ |
	// if falsecase result is not on the stack     |
	//    PUSH falsecase result                    |
	//                                         <---+
	//
	instr := bytecode.New(bytecode.JMPF) | i.Condition.byteCode(0, cs, ds)
	*cs = append(*cs, instr)
	jmpfAddr := len(*cs) - 1

	instr = i.TrueCase.byteCode(0, cs, ds)
	if instr.Src0() != bytecode.ADDR_STCK {
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}

	jmpTAddr := len(*cs)
	instr = bytecode.New(bytecode.JMP)
	*cs = append(*cs, instr)

	typErrAddr := len(*cs)
	ix := len(*ds)
	*ds = append(*ds, value.TypeError)
	instr = bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	instr = i.FalseCase.byteCode(0, cs, ds)
	if instr.Src0() != bytecode.ADDR_STCK {
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}

	// patch jmpf
	(*cs)[jmpfAddr] |=
		bytecode.EncodeSrc(2, bytecode.ADDR_IMM, typErrAddr-jmpfAddr) |
			bytecode.EncodeSrc(1, bytecode.ADDR_IMM, jmpTAddr-jmpfAddr+1)
	// patch jmp
	(*cs)[jmpTAddr] |= bytecode.EncodeSrc(0, bytecode.ADDR_IMM, len(*cs)-jmpTAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (w While) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// PUSH no result
	// JMPF condition                                <-----+ -+
	// POP                                                 |  |
	// loop body                                           |  |
	// if body didn't leave its result on the stack        |  |
	//    PUSH body result                                 |  |
	// JUMP                                           -----+  |
	// PUSH type error                                <-------|
	//                                                <-------+

	ix := len(*ds)
	*ds = append(*ds, value.NoResultError)
	instr := bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	condAddr := len(*cs)
	instr = bytecode.New(bytecode.JMPF) | w.Condition.byteCode(0, cs, ds)
	*cs = append(*cs, instr)

	jmpfAddr := len(*cs) - 1

	instr = bytecode.New(bytecode.POP)
	*cs = append(*cs, instr)

	instr = w.Body.byteCode(0, cs, ds)

	if instr.Src0() != bytecode.ADDR_STCK {
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.ADDR_IMM, condAddr-len(*cs))
	*cs = append(*cs, instr)

	typErrAddr := len(*cs)
	ix = len(*ds)
	*ds = append(*ds, value.TypeError)
	instr = bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.ADDR_DS, ix)
	*cs = append(*cs, instr)

	// patch the JMPF
	(*cs)[jmpfAddr] |=
		bytecode.EncodeSrc(2, bytecode.ADDR_IMM, typErrAddr-jmpfAddr) |
			bytecode.EncodeSrc(1, bytecode.ADDR_IMM, len(*cs)-jmpfAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

var contextId = 0

func (f For) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
  ccontAddr:=len(*cs)
	instr := bytecode.New(bytecode.CCONT) //| bytecode.EncodeSrc(0, bytecode.ADDR_IMM, contextId)
	*cs = append(*cs, instr)
  contextId++

	instr = f.Iterator.byteCode(0, cs, ds)
	if instr.Src0() == bytecode.ADDR_STCK {
		instr = bytecode.New(bytecode.POP)
    *cs = append(*cs, instr)
	}

	instr = bytecode.New(bytecode.DCONT)
	*cs = append(*cs, instr)

	jmpAddr := len(*cs)
	instr = bytecode.New(bytecode.JMP)
	*cs = append(*cs, instr)

	switchAddr := len(*cs)
	instr = bytecode.New(bytecode.SCONT)
	*cs = append(*cs, instr)

	assignAddr := len(*cs)
	vref := f.VarRef.byteCode(1, cs, ds)
	instr = bytecode.New(bytecode.MOV) | vref | bytecode.EncodeSrc(0, bytecode.ADDR_STCK, 0)
	*cs = append(*cs, instr)

	instr = f.Body.byteCode(0, cs, ds)

  if instr.Src0() != bytecode.ADDR_STCK {
    instr = bytecode.New(bytecode.PUSH) | instr
    *cs = append(*cs, instr)
  }

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.ADDR_IMM, switchAddr-len(*cs))
	*cs = append(*cs, instr)

	// patch jump
	(*cs)[jmpAddr] |= bytecode.EncodeSrc(0, bytecode.ADDR_IMM, len(*cs)-jmpAddr)

  // patch ccont
  (*cs)[ccontAddr] |= bytecode.EncodeSrc(0, bytecode.ADDR_IMM, assignAddr-ccontAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (i IndexAt) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	ary := i.Ary.byteCode(1, cs, ds)
	at := i.At.byteCode(0, cs, ds)
	instr := bytecode.New(bytecode.IX1) | ary | at

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (i IndexFromTo) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	ary := i.Ary.byteCode(2, cs, ds)
	from := i.From.byteCode(1, cs, ds)
	to := i.To.byteCode(0, cs, ds)

	instr := bytecode.New(bytecode.IX2) | ary | from | to

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (r Read) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.READ)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (w Write) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.WRITE) | w.Value.byteCode(0, cs, ds)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (a Aton) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.ATON) | a.Value.byteCode(0, cs, ds)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (t Toa) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.TOA) | t.Value.byteCode(0, cs, ds)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (e Error) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.ERROR) | e.Value.byteCode(0, cs, ds)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

func (e Exit) byteCode(srcsel int, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.EXIT) | e.Value.byteCode(0, cs, ds)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.ADDR_STCK, 0)
}

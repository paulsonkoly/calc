package node

import (
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/compresult"
	"github.com/paulsonkoly/calc/types/dbginfo"
	"github.com/paulsonkoly/calc/types/value"
)

type compResult = compresult.Type

type ByteCoder interface {
	byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type
}

func ByteCode(bc ByteCoder, cr compResult) {
	instr := bc.byteCode(0, false, cr)
	if instr.Src0() != bytecode.AddrStck { // leave the final result on the stack
		instr |= bytecode.New(bytecode.PUSH)
		*cr.CS = append(*cr.CS, instr)
	}
}

func ByteCodeNoStck(bc ByteCoder, cr compResult) {
	instr := bc.byteCode(0, false, cr)
	if instr.Src0() == bytecode.AddrStck { // don't leave the final result on the stack
		instr = bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
	}
}

func (i Int) byteCode(srcsel int, _ bool, cr compResult) bytecode.Type {
	v := value.NewInt(int(i))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (b Bool) byteCode(srcsel int, _ bool, cr compResult) bytecode.Type {
	v := value.NewBool(bool(b))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (f Float) byteCode(srcsel int, _ bool, cr compResult) bytecode.Type {
	v := value.NewFloat(float64(f))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (s String) byteCode(srcsel int, _ bool, cr compResult) bytecode.Type {
	v := value.NewString(string(s))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (l List) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	// TODO optimise at least the constant parts of array case to the ds
	v := value.NewArray([]value.Type{})
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)
	if len(l.Elems) == 0 {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
	}

	instr := l.Elems[0].byteCode(0, inFor, cr)
	instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.AddrDS, ix)
	*cr.CS = append(*cr.CS, instr)

	for _, t := range l.Elems[1:] {
		instr = t.byteCode(0, inFor, cr)
		instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.AddrStck, 0)
		*cr.CS = append(*cr.CS, instr)
	}

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (l Local) byteCode(srcsel int, _ bool, _ compResult) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.AddrLcl, l.Ix)
}

func (c Closure) byteCode(srcsel int, _ bool, _ compResult) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.AddrCls, c.Ix)
}

func (n Name) byteCode(srcsel int, _ bool, cr compResult) bytecode.Type {
	v := value.NewString(string(n))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)
	return bytecode.EncodeSrc(srcsel, bytecode.AddrGbl, ix)
}

func (f Function) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.JMP)
	*cr.CS = append(*cr.CS, instr)
	jmpAddr := len(*cr.CS) - 1

	instr = f.Body.byteCode(0, inFor, cr)
	instr |= bytecode.New(bytecode.RET)
	*cr.CS = append(*cr.CS, instr)

	instr = bytecode.New(bytecode.FUNC) |
		bytecode.EncodeSrc(2, bytecode.AddrImm, (f.LocalCnt)) |
		bytecode.EncodeSrc(1, bytecode.AddrImm, len(f.Parameters.Elems)) |
		bytecode.EncodeSrc(0, bytecode.AddrImm, jmpAddr+1)
	*cr.CS = append(*cr.CS, instr)

	// patch the jmp
	(*cr.CS)[jmpAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cr.CS)-jmpAddr-1)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (c Call) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	// push the arguments
	for _, arg := range c.Arguments.Elems {
		instr := arg.byteCode(0, inFor, cr)
		if instr.Src0() != bytecode.AddrStck {
			instr |= bytecode.New(bytecode.PUSH)
			*cr.CS = append(*cr.CS, instr)
		}
	}

	name, ok := c.Name.(Namer)
	if !ok {
		panic("function name is not held by a named node")
	}
	(*cr.Dbg)[len(*cr.CS)] = dbginfo.Call{Name: string(name.Name()), ArgCnt: len(c.Arguments.Elems)}

	// get the function
	instr := c.Name.byteCode(0, inFor, cr)

	instr |= bytecode.New(bytecode.CALL) | bytecode.EncodeSrc(1, bytecode.AddrImm, len(c.Arguments.Elems))
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (r Return) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	if inFor {
		instr := bytecode.New(bytecode.RCONT)
		*cr.CS = append(*cr.CS, instr)
	}
	instr := r.Target.byteCode(0, inFor, cr)
	instr |= bytecode.New(bytecode.RET)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (y Yield) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	instr := y.Target.byteCode(0, inFor, cr)
	instr |= bytecode.New(bytecode.YIELD)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (a Assign) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	srcInstr := a.Value.byteCode(0, inFor, cr)
	instr := srcInstr | a.VarRef.byteCode(1, inFor, cr)
	instr |= bytecode.New(bytecode.MOV)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, instr.Src1(), instr.Src1Addr())
}

func (b BinOp) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
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
	case "&", "&&":
		op = bytecode.AND
	case "|", "||":
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
	case "<<":
		op = bytecode.LSH
	case ">>":
		op = bytecode.RSH
	default:
		panic("unexpected op")
	}
	instr := b.Left.byteCode(1, inFor, cr)
	instr |= b.Right.byteCode(0, inFor, cr)
	instr |= bytecode.New(op)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (u UnOp) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	var op bytecode.OpCode
	switch u.Op {
	case "-":
		return BinOp{Op: "*", Left: Int(-1), Right: u.Target}.byteCode(srcsel, inFor, cr)
	case "#":
		op = bytecode.LEN
	case "!":
		op = bytecode.NOT
	case "~":
		op = bytecode.FLIP
	default:
		panic("unexpected op")
	}

	instr := bytecode.New(op) | u.Target.byteCode(0, inFor, cr)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (b Block) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	var instr bytecode.Type
	for i, t := range b.Body {
		instr = t.byteCode(srcsel, inFor, cr)
		if i != len(b.Body)-1 { // throw away mid-block results
			if instr.Src(srcsel) == bytecode.AddrStck {
				instr = bytecode.New(bytecode.POP)
				*cr.CS = append(*cr.CS, instr)
			}
		}
	}
	return instr
}

func (i If) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	//
	// JMPF ..., condition                     --+
	// truecase                                  |
	// if trucase result is not on the stack     |
	//    PUSH truecase result                   |
	// JMP +2                                  --|-+
	// PUSH no result                          <-+ |
	//                                         <---+
	//
	instr := bytecode.New(bytecode.JMPF) | i.Condition.byteCode(0, inFor, cr)

	*cr.CS = append(*cr.CS, instr)
	jmpfAddr := len(*cr.CS) - 1

	tcInstr := i.TrueCase.byteCode(0, inFor, cr)
	if tcInstr.Src0() != bytecode.AddrStck {
		instr = bytecode.New(bytecode.PUSH) | tcInstr
		*cr.CS = append(*cr.CS, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, 2)
	*cr.CS = append(*cr.CS, instr)

	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, value.Nil)
	instr = bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
	*cr.CS = append(*cr.CS, instr)

	// patch the JMPF
	(*cr.CS)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, len(*cr.CS)-jmpfAddr-1)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IfElse) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	//
	// JMPF ..., condition                     --+
	// truecase                                  |
	// if trucase result is not on the stack     |
	//    PUSH truecase result                   |
	// JMP                                     --|-+
	// falsecase                               <-+ |
	// if falsecase result is not on the stack     |
	//    PUSH falsecase result                    |
	//                                         <---+
	//
	instr := bytecode.New(bytecode.JMPF) | i.Condition.byteCode(0, inFor, cr)
	*cr.CS = append(*cr.CS, instr)
	jmpfAddr := len(*cr.CS) - 1

	instr = i.TrueCase.byteCode(0, inFor, cr)
	if instr.Src0() != bytecode.AddrStck {
		instr |= bytecode.New(bytecode.PUSH)
		*cr.CS = append(*cr.CS, instr)
	}

	jmpTAddr := len(*cr.CS)
	instr = bytecode.New(bytecode.JMP)
	*cr.CS = append(*cr.CS, instr)

	instr = i.FalseCase.byteCode(0, inFor, cr)
	if instr.Src0() != bytecode.AddrStck {
		instr |= bytecode.New(bytecode.PUSH)
		*cr.CS = append(*cr.CS, instr)
	}

	// patch jmpf
	(*cr.CS)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, jmpTAddr-jmpfAddr+1)
	// patch jmp
	(*cr.CS)[jmpTAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cr.CS)-jmpTAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (w While) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	// PUSH no result
	// JMPF condition                                <-----+ -+
	// POP                                                 |  |
	// loop body                                           |  |
	// if body didn't leave its result on the stack        |  |
	//    PUSH body result                                 |  |
	// JUMP                                           -----+  |
	//                                                <-------+

	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, value.Nil)
	instr := bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
	*cr.CS = append(*cr.CS, instr)

	condAddr := len(*cr.CS)
	instr = bytecode.New(bytecode.JMPF) | w.Condition.byteCode(0, inFor, cr)
	*cr.CS = append(*cr.CS, instr)

	jmpfAddr := len(*cr.CS) - 1

	instr = bytecode.New(bytecode.POP)
	*cr.CS = append(*cr.CS, instr)

	instr = w.Body.byteCode(0, inFor, cr)

	if instr.Src0() != bytecode.AddrStck {
		instr |= bytecode.New(bytecode.PUSH)
		*cr.CS = append(*cr.CS, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, condAddr-len(*cr.CS))
	*cr.CS = append(*cr.CS, instr)

	// patch the JMPF
	(*cr.CS)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, len(*cr.CS)-jmpfAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (f For) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	//
	// PUSH nil
	// CCONT                               ------+                  ; fork iterator context
	// Iterator byteCode                         |
	// if iterator returned on the stack         |
	//   POP                                     |
	// DCONT                                     |                  ; destroy ierator context
	// JMP                                  =====|==+
	// SCONT                               <~~~~~|~~|~~+            ; switch to iterator from for loop body
	// MOV lvar, STCK                      <-----+  |  |
	// POP                                          |  |            ; previous loop result
	// loopBody byteCode                            |  |
	// if loopbBody didn't return on stack          |  |
	//     PUSH loopBody result                     |  |
	// JMP                                  ~~~~~~~~|~~+
	//                                      <=======+
	//
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, value.Nil)
	instr := bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
	*cr.CS = append(*cr.CS, instr)

	ccontAddr := len(*cr.CS)
	instr = bytecode.New(bytecode.CCONT)
	*cr.CS = append(*cr.CS, instr)

	instr = f.Iterator.byteCode(0, inFor, cr)
	if instr.Src0() == bytecode.AddrStck {
		instr = bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
	}

	instr = bytecode.New(bytecode.DCONT)
	*cr.CS = append(*cr.CS, instr)

	jmpAddr := len(*cr.CS)
	instr = bytecode.New(bytecode.JMP)
	*cr.CS = append(*cr.CS, instr)

	switchAddr := len(*cr.CS)
	instr = bytecode.New(bytecode.SCONT)
	*cr.CS = append(*cr.CS, instr)

	assignAddr := len(*cr.CS)
	vref := f.VarRef.byteCode(1, inFor, cr)
	instr = bytecode.New(bytecode.MOV) | vref | bytecode.EncodeSrc(0, bytecode.AddrStck, 0)
	*cr.CS = append(*cr.CS, instr)

	instr = bytecode.New(bytecode.POP)
	*cr.CS = append(*cr.CS, instr)

	instr = f.Body.byteCode(0, true, cr)

	if instr.Src0() != bytecode.AddrStck {
		instr = bytecode.New(bytecode.PUSH) | instr
		*cr.CS = append(*cr.CS, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, switchAddr-len(*cr.CS))
	*cr.CS = append(*cr.CS, instr)

	// patch jump
	(*cr.CS)[jmpAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cr.CS)-jmpAddr)

	// patch ccont
	(*cr.CS)[ccontAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, assignAddr-ccontAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IndexAt) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	ary := i.Ary.byteCode(1, inFor, cr)
	at := i.At.byteCode(0, inFor, cr)
	instr := bytecode.New(bytecode.IX1) | ary | at

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IndexFromTo) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	ary := i.Ary.byteCode(2, inFor, cr)
	from := i.From.byteCode(1, inFor, cr)
	to := i.To.byteCode(0, inFor, cr)

	instr := bytecode.New(bytecode.IX2) | ary | from | to

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (r Read) byteCode(srcsel int, _ bool, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.READ)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (w Write) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.WRITE) | w.Value.byteCode(0, inFor, cr)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (a Aton) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.ATON) | a.Value.byteCode(0, inFor, cr)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (t Toa) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.TOA) | t.Value.byteCode(0, inFor, cr)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (e Exit) byteCode(srcsel int, inFor bool, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.EXIT) | e.Value.byteCode(0, inFor, cr)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

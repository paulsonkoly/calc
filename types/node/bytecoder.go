package node

import (
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/value"
)

type ByteCoder interface {
	byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type
}

func ByteCode(bc ByteCoder, cs *[]bytecode.Type, ds *[]value.Type) {
	instr := bc.byteCode(0, false, cs, ds)
	if instr.Src0() != bytecode.AddrStck { // leave the final result on the stack
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}
}

func ByteCodeNoStck(bc ByteCoder, cs *[]bytecode.Type, ds *[]value.Type) {
	instr := bc.byteCode(0, false, cs, ds)
	if instr.Src0() == bytecode.AddrStck { // don't leave the final result on the stack
		instr = bytecode.New(bytecode.POP)
		*cs = append(*cs, instr)
	}
}

func (i Int) byteCode(srcsel int, _ bool, _ *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewInt(int(i))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (b Bool) byteCode(srcsel int, _ bool, _ *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewBool(bool(b))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (f Float) byteCode(srcsel int, _ bool, _ *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewFloat(float64(f))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (s String) byteCode(srcsel int, _ bool, _ *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewString(string(s))
	ix := len(*ds)
	*ds = append(*ds, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (l List) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// TODO optimise at least the constant parts of array case to the ds
	v := value.NewArray([]value.Type{})
	ix := len(*ds)
	*ds = append(*ds, v)
	if len(l.Elems) == 0 {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
	}

	instr := l.Elems[0].byteCode(0, inFor, cs, ds)
	instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.AddrDS, ix)
	*cs = append(*cs, instr)

	for _, t := range l.Elems[1:] {
		instr = t.byteCode(0, inFor, cs, ds)
		instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.AddrStck, 0)
		*cs = append(*cs, instr)
	}

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (l Local) byteCode(srcsel int, _ bool, _ *[]bytecode.Type, _ *[]value.Type) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.AddrLcl, int(l))
}

func (c Closure) byteCode(srcsel int, _ bool, _ *[]bytecode.Type, _ *[]value.Type) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.AddrCls, int(c))
}

func (n Name) byteCode(srcsel int, _ bool, _ *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	v := value.NewString(string(n))
	ix := len(*ds)
	*ds = append(*ds, v)
	return bytecode.EncodeSrc(srcsel, bytecode.AddrGbl, ix)
}

func (f Function) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.JMP)
	*cs = append(*cs, instr)
	jmpAddr := len(*cs) - 1

	instr = f.Body.byteCode(0, inFor, cs, ds)
	instr |= bytecode.New(bytecode.RET)
	*cs = append(*cs, instr)

	instr = bytecode.New(bytecode.FUNC) |
		bytecode.EncodeSrc(2, bytecode.AddrImm, (f.LocalCnt)) |
		bytecode.EncodeSrc(1, bytecode.AddrImm, len(f.Parameters.Elems)) |
		bytecode.EncodeSrc(0, bytecode.AddrImm, jmpAddr+1)
	*cs = append(*cs, instr)

	// patch the jmp
	(*cs)[jmpAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cs)-jmpAddr-1)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (c Call) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// push the arguments
	for _, arg := range c.Arguments.Elems {
		instr := arg.byteCode(0, inFor, cs, ds)
		if instr.Src0() != bytecode.AddrStck {
			instr |= bytecode.New(bytecode.PUSH)
			*cs = append(*cs, instr)
		}
	}

	// get the function
	instr := c.Name.byteCode(0, inFor, cs, ds)

	instr |= bytecode.New(bytecode.CALL) | bytecode.EncodeSrc(1, bytecode.AddrImm, len(c.Arguments.Elems))
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (r Return) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	if inFor {
		instr := bytecode.New(bytecode.RCONT)
		*cs = append(*cs, instr)
	}
	instr := r.Target.byteCode(0, inFor, cs, ds)
	instr |= bytecode.New(bytecode.RET)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (y Yield) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := y.Target.byteCode(0, inFor, cs, ds)
	instr |= bytecode.New(bytecode.YIELD)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (a Assign) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	srcInstr := a.Value.byteCode(0, inFor, cs, ds)
	instr := srcInstr | a.VarRef.byteCode(1, inFor, cs, ds)
	instr |= bytecode.New(bytecode.MOV)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, instr.Src1(), instr.Src1Addr())
}

func (b BinOp) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
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
	instr := b.Left.byteCode(1, inFor, cs, ds)
	instr |= b.Right.byteCode(0, inFor, cs, ds)
	instr |= bytecode.New(op)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (u UnOp) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	var op bytecode.OpCode
	switch u.Op {
	case "-":
		return BinOp{Op: "*", Left: Int(-1), Right: u.Target}.byteCode(srcsel, inFor, cs, ds)
	case "#":
		op = bytecode.LEN
	case "!":
		op = bytecode.NOT
	case "~":
		op = bytecode.FLIP
	default:
		panic("unexpected op")
	}

	instr := bytecode.New(op) | u.Target.byteCode(0, inFor, cs, ds)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (b Block) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	var instr bytecode.Type
	for i, t := range b.Body {
		instr = t.byteCode(srcsel, inFor, cs, ds)
		if i != len(b.Body)-1 { // throw away mid-block results
			if instr.Src(srcsel) == bytecode.AddrStck {
				instr = bytecode.New(bytecode.POP)
				*cs = append(*cs, instr)
			}
		}
	}
	return instr
}

func (i If) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	//
	// JMPF ..., condition                     --+
	// truecase                                  |
	// if trucase result is not on the stack     |
	//    PUSH truecase result                   |
	// JMP +2                                  --|-+
	// PUSH no result                          <-+ |
	//                                         <---+
	//
	instr := bytecode.New(bytecode.JMPF) | i.Condition.byteCode(0, inFor, cs, ds)

	*cs = append(*cs, instr)
	jmpfAddr := len(*cs) - 1

	tcInstr := i.TrueCase.byteCode(0, inFor, cs, ds)
	if tcInstr.Src0() != bytecode.AddrStck {
		instr = bytecode.New(bytecode.PUSH) | tcInstr
		*cs = append(*cs, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, 2)
	*cs = append(*cs, instr)

	ix := len(*ds)
	*ds = append(*ds, value.Nil)
	instr = bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
	*cs = append(*cs, instr)

	// patch the JMPF
	(*cs)[jmpfAddr] |=
		bytecode.EncodeSrc(2, bytecode.AddrImm, len(*cs)-jmpfAddr-1) |
			bytecode.EncodeSrc(1, bytecode.AddrImm, len(*cs)-jmpfAddr-3)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IfElse) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
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
	instr := bytecode.New(bytecode.JMPF) | i.Condition.byteCode(0, inFor, cs, ds)
	*cs = append(*cs, instr)
	jmpfAddr := len(*cs) - 1

	instr = i.TrueCase.byteCode(0, inFor, cs, ds)
	if instr.Src0() != bytecode.AddrStck {
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}

	jmpTAddr := len(*cs)
	instr = bytecode.New(bytecode.JMP)
	*cs = append(*cs, instr)

	instr = i.FalseCase.byteCode(0, inFor, cs, ds)
	if instr.Src0() != bytecode.AddrStck {
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}

	// patch jmpf
	(*cs)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, jmpTAddr-jmpfAddr+1)
	// patch jmp
	(*cs)[jmpTAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cs)-jmpTAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (w While) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	// PUSH no result
	// JMPF condition                                <-----+ -+
	// POP                                                 |  |
	// loop body                                           |  |
	// if body didn't leave its result on the stack        |  |
	//    PUSH body result                                 |  |
	// JUMP                                           -----+  |
	//                                                <-------+

	ix := len(*ds)
	*ds = append(*ds, value.Nil)
	instr := bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
	*cs = append(*cs, instr)

	condAddr := len(*cs)
	instr = bytecode.New(bytecode.JMPF) | w.Condition.byteCode(0, inFor, cs, ds)
	*cs = append(*cs, instr)

	jmpfAddr := len(*cs) - 1

	instr = bytecode.New(bytecode.POP)
	*cs = append(*cs, instr)

	instr = w.Body.byteCode(0, inFor, cs, ds)

	if instr.Src0() != bytecode.AddrStck {
		instr |= bytecode.New(bytecode.PUSH)
		*cs = append(*cs, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, condAddr-len(*cs))
	*cs = append(*cs, instr)

	// patch the JMPF
	(*cs)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, len(*cs)-jmpfAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (f For) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	ix := len(*ds)
	*ds = append(*ds, value.Nil)
	instr := bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
	*cs = append(*cs, instr)

	ccontAddr := len(*cs)
	instr = bytecode.New(bytecode.CCONT)
	*cs = append(*cs, instr)

	instr = f.Iterator.byteCode(0, inFor, cs, ds)
	if instr.Src0() == bytecode.AddrStck {
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
	vref := f.VarRef.byteCode(1, inFor, cs, ds)
	instr = bytecode.New(bytecode.MOV) | vref | bytecode.EncodeSrc(0, bytecode.AddrStck, 0)
	*cs = append(*cs, instr)

	instr = bytecode.New(bytecode.POP)
	*cs = append(*cs, instr)

	instr = f.Body.byteCode(0, true, cs, ds)

	if instr.Src0() != bytecode.AddrStck {
		instr = bytecode.New(bytecode.PUSH) | instr
		*cs = append(*cs, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, switchAddr-len(*cs))
	*cs = append(*cs, instr)

	// patch jump
	(*cs)[jmpAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cs)-jmpAddr)

	// patch ccont
	(*cs)[ccontAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, assignAddr-ccontAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IndexAt) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	ary := i.Ary.byteCode(1, inFor, cs, ds)
	at := i.At.byteCode(0, inFor, cs, ds)
	instr := bytecode.New(bytecode.IX1) | ary | at

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IndexFromTo) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	ary := i.Ary.byteCode(2, inFor, cs, ds)
	from := i.From.byteCode(1, inFor, cs, ds)
	to := i.To.byteCode(0, inFor, cs, ds)

	instr := bytecode.New(bytecode.IX2) | ary | from | to

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (r Read) byteCode(srcsel int, _ bool, cs *[]bytecode.Type, _ *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.READ)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (w Write) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.WRITE) | w.Value.byteCode(0, inFor, cs, ds)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (a Aton) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.ATON) | a.Value.byteCode(0, inFor, cs, ds)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (t Toa) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.TOA) | t.Value.byteCode(0, inFor, cs, ds)
	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (e Exit) byteCode(srcsel int, inFor bool, cs *[]bytecode.Type, ds *[]value.Type) bytecode.Type {
	instr := bytecode.New(bytecode.EXIT) | e.Value.byteCode(0, inFor, cs, ds)

	*cs = append(*cs, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

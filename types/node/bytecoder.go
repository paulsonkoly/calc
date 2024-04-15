package node

import (
	"github.com/paulsonkoly/calc/types/bytecode"
	"github.com/paulsonkoly/calc/types/compresult"
	"github.com/paulsonkoly/calc/types/dbginfo"
	flags "github.com/paulsonkoly/calc/types/node/bc"
	"github.com/paulsonkoly/calc/types/value"
)

// At what depth of operator arithmetics do we start using the temp register.
const tempifyDepth = 0

type compResult = compresult.Type

type ByteCoder interface {
	byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type
}

// ByteCode compiles bc and appends the results in cr.
//
// The evaluation result is left on the stack.
func ByteCode(bc ByteCoder, cr compResult) {
	var fl flags.Data

	instr := bc.byteCode(0, fl.Pass(), cr)
	if instr.Src0() != bytecode.AddrStck { // leave the final result on the stack
		instr |= bytecode.New(bytecode.PUSH)
		*cr.CS = append(*cr.CS, instr)
	}
}

// ByteCodeNoStck compiles bc and appends the results in cr.
//
// The evaluation result is lost, code is expected to run for side effects.
func ByteCodeNoStck(bc ByteCoder, cr compResult) {
	var fl flags.Data

	instr := bc.byteCode(0, fl.Pass(flags.WithDiscard(true)), cr)
	if instr.Src0() == bytecode.AddrStck { // don't leave the final result on the stack
		instr = bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
	}
}

func (i Int) byteCode(srcsel int, _ flags.Pass, cr compResult) bytecode.Type {
	v := value.NewInt(int(i))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (b Bool) byteCode(srcsel int, _ flags.Pass, cr compResult) bytecode.Type {
	v := value.NewBool(bool(b))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (f Float) byteCode(srcsel int, _ flags.Pass, cr compResult) bytecode.Type {
	v := value.NewFloat(float64(f))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (s String) byteCode(srcsel int, _ flags.Pass, cr compResult) bytecode.Type {
	v := value.NewString(string(s))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
}

func (l List) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	ary := make([]value.Type, 0, len(l.Elems))
	i := 0

	for ; i < len(l.Elems); i++ {
		v, ok := l.Elems[i].Constant()
		if !ok {
			break
		}
		ary = append(ary, v)
	}

	v := value.NewArray(ary)
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)
	if i >= len(l.Elems) {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrDS, ix)
	}

	instr := l.Elems[i].byteCode(0, fl.Data().Pass(), cr)
	instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.AddrDS, ix)
	*cr.CS = append(*cr.CS, instr)

	for _, t := range l.Elems[i+1:] {
		instr = t.byteCode(0, fl.Data().Pass(), cr)
		instr |= bytecode.New(bytecode.ARR) | bytecode.EncodeSrc(1, bytecode.AddrStck, 0)
		*cr.CS = append(*cr.CS, instr)
	}

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (l Local) byteCode(srcsel int, _ flags.Pass, _ compResult) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.AddrLcl, l.Ix)
}

func (c Closure) byteCode(srcsel int, _ flags.Pass, _ compResult) bytecode.Type {
	return bytecode.EncodeSrc(srcsel, bytecode.AddrCls, c.Ix)
}

func (n Name) byteCode(srcsel int, _ flags.Pass, cr compResult) bytecode.Type {
	v := value.NewString(string(n))
	ix := len(*cr.DS)
	*cr.DS = append(*cr.DS, v)
	return bytecode.EncodeSrc(srcsel, bytecode.AddrGbl, ix)
}

func (f Function) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.JMP)
	*cr.CS = append(*cr.CS, instr)
	jmpAddr := len(*cr.CS) - 1

	subfl := fl.Data().Pass(
		flags.WithInFor(false),
		flags.WithForbidTemp(false),
		flags.WithOpDepth(0),
		flags.WithInFunc(true))

	instr = f.Body.byteCode(0, subfl, cr)
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

func (c Call) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	subfl := fl.Data().Pass(flags.WithOpDepth(0))

	// push the arguments
	for _, arg := range c.Arguments.Elems {
		instr := arg.byteCode(0, subfl, cr)
		if instr.Src0() != bytecode.AddrStck && instr.Src0() != bytecode.AddrInv {
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
	instr := c.Name.byteCode(0, fl.Data().Pass(), cr)

	instr |= bytecode.New(bytecode.CALL) | bytecode.EncodeSrc(1, bytecode.AddrImm, len(c.Arguments.Elems))
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (r Return) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	if fl.Data().InFor {
		instr := bytecode.New(bytecode.RCONT) |
			bytecode.EncodeSrc(0, bytecode.AddrImm, fl.Data().CtxLo) |
			bytecode.EncodeSrc(1, bytecode.AddrImm, fl.Data().CtxHi)
		*cr.CS = append(*cr.CS, instr)
	}
	target := r.Target.byteCode(0, fl.Data().Pass(), cr)
	instr := bytecode.New(bytecode.RET) | target
	*cr.CS = append(*cr.CS, instr)

	// return doesn't leave result - at least in the current lexical scope. It
	// pushes the function return value but that has to be encoded in the scope
	// of the  Call. If there is no function ret must return on the stack in the
	// current scope
	if fl.Data().InFunc {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrInv, 0)
	}
	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (y Yield) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	instr := y.Target.byteCode(0, fl.Data().Pass(), cr)
	instr |= bytecode.New(bytecode.YIELD)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (a Assign) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	vref := a.VarRef
	if plus, ok := a.Value.(BinOp); ok && plus.Op == "+" {
		inc := false

		if one, ok := plus.Right.(Int); ok && int(one) == 1 && plus.Left == vref {
			inc = true
		}

		if one, ok := plus.Left.(Int); ok && int(one) == 1 && plus.Right == vref {
			inc = true
		}

		if inc {
			instr := bytecode.New(bytecode.INC) | vref.byteCode(0, fl.Data().Pass(), cr)
			*cr.CS = append(*cr.CS, instr)

			return bytecode.EncodeSrc(srcsel, instr.Src0(), instr.Src0Addr())
		}
	}

	srcInstr := a.Value.byteCode(0, fl.Data().Pass(), cr)
	instr := srcInstr | vref.byteCode(1, fl.Data().Pass(), cr)
	instr |= bytecode.New(bytecode.MOV)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, instr.Src1(), instr.Src1Addr())
}

func (b BinOp) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	var op bytecode.OpCode
	var left, right bytecode.Type
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

	forbidTemp := fl.Data().ForbidTemp || b.Right.HasCall()
	opDepth := fl.Data().OpDepth
	var tempified = false

	left = b.Left.byteCode(1, fl.Data().Pass(flags.WithForbidTemp(forbidTemp), flags.WithOpDepth(opDepth+1)), cr)
	if left.Src1() == bytecode.AddrTmp {
		tempified = true
	}

	if !forbidTemp && opDepth > tempifyDepth && !tempified {
		instr := bytecode.New(bytecode.MOV) |
			bytecode.EncodeSrc(1, bytecode.AddrTmp, 0) |
			bytecode.EncodeSrc(0, left.Src1(), left.Src1Addr())
		*cr.CS = append(*cr.CS, instr)

		tempified = true
	}

	right = b.Right.byteCode(0, fl.Data().Pass(flags.WithForbidTemp(true)), cr)

	var instr bytecode.Type
	if tempified {
		instr = bytecode.New(op|bytecode.TempFlag) | right
	} else {
		instr = bytecode.New(op) | left | right
	}
	*cr.CS = append(*cr.CS, instr)

	if tempified && opDepth == 0 {
		instr = bytecode.New(bytecode.PUSHTMP)
		*cr.CS = append(*cr.CS, instr)

		tempified = false
	}

	if tempified {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrTmp, 0)
	}

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (u UnOp) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	var op bytecode.OpCode

	forbidTemp := fl.Data().ForbidTemp
	opDepth := fl.Data().OpDepth

	switch u.Op {
	case "-":
		return BinOp{Op: "*", Left: Int(-1), Right: u.Target}.byteCode(srcsel, fl.Data().Pass(), cr)
	case "#":
		op = bytecode.LEN
	case "!":
		op = bytecode.NOT
	case "~":
		op = bytecode.FLIP
	default:
		panic("unexpected op")
	}

	tempified := false
	target := u.Target.byteCode(0, fl.Data().Pass(flags.WithOpDepth(opDepth+1)), cr)
	if target.Src0() == bytecode.AddrTmp {
		tempified = true
	}

	if !forbidTemp && opDepth > tempifyDepth && !tempified {
		instr := bytecode.New(bytecode.MOV) | bytecode.EncodeSrc(1, bytecode.AddrTmp, 0) | target
		*cr.CS = append(*cr.CS, instr)

		tempified = true
	}

	var instr bytecode.Type
	if tempified {
		instr = bytecode.New(op | bytecode.TempFlag)
	} else {
		instr = bytecode.New(op) | target
	}
	*cr.CS = append(*cr.CS, instr)

	if tempified && opDepth == 0 {
		instr := bytecode.New(bytecode.PUSHTMP)
		*cr.CS = append(*cr.CS, instr)

		tempified = false
	}

	if tempified {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrTmp, 0)
	}

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)

}

func (b Block) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	var instr bytecode.Type
	for i, t := range b.Body {
		if i != len(b.Body)-1 { // throw away mid-block results
			instr = t.byteCode(srcsel, fl.Data().Pass(flags.WithDiscard(true)), cr)
			if instr.Src(srcsel) == bytecode.AddrStck {
				instr = bytecode.New(bytecode.POP)
				*cr.CS = append(*cr.CS, instr)
			}
		} else {
			instr = t.byteCode(srcsel, fl.Data().Pass(), cr)
		}
	}
	return instr
}

func (i If) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	//
	// JMPF ..., condition                     --+
	// truecase                                  |
	// if trucase result is not on the stack     |
	//    PUSH truecase result                   |
	// JMP +2                                  --|-+
	// PUSH no result                          <-+ |
	//                                         <---+
	//

	discard := fl.Data().Discard

	condition := i.Condition.byteCode(0, fl.Data().Pass(), cr)
	jmpfAddr := len(*cr.CS)
	instr := bytecode.New(bytecode.JMPF) | condition
	*cr.CS = append(*cr.CS, instr)

	tcSize := len(*cr.CS)
	tcInstr := i.TrueCase.byteCode(0, fl.Data().Pass(), cr)
	tcSize = len(*cr.CS) - tcSize

	if tcSize == 0 && discard {
		// if true case produced no code ie, single immediate like
		// if false 1 and we are discarding, then we don't even need the if
		*cr.CS = (*cr.CS)[:len(*cr.CS)-1]
		return bytecode.EncodeSrc(srcsel, tcInstr.Src0(), tcInstr.Src0Addr())
	}

	dest := bytecode.EncodeSrc(srcsel, tcInstr.Src0(), tcInstr.Src0Addr())
	if tcInstr.Src0() != bytecode.AddrStck && tcInstr.Src0() != bytecode.AddrInv && !discard {
		instr = bytecode.New(bytecode.PUSH) | tcInstr
		*cr.CS = append(*cr.CS, instr)
		dest = bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
	}

	if tcInstr.Src0() == bytecode.AddrStck && discard {
		instr = bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
		dest = bytecode.EncodeSrc(srcsel, bytecode.AddrInv, 0)
	}

	// if discard then there is no noResultAddr, this now points to the end
	noResultAddr := len(*cr.CS)

	if !discard {
		instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, 2)
		*cr.CS = append(*cr.CS, instr)

		noResultAddr = len(*cr.CS)

		ix := len(*cr.DS)
		*cr.DS = append(*cr.DS, value.Nil)
		instr = bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
		*cr.CS = append(*cr.CS, instr)
	}

	// patch the JMPF
	(*cr.CS)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, noResultAddr-jmpfAddr)

	return dest
}

func (i IfElse) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
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
	instr := bytecode.New(bytecode.JMPF) | i.Condition.byteCode(0, fl.Data().Pass(), cr)
	*cr.CS = append(*cr.CS, instr)
	jmpfAddr := len(*cr.CS) - 1

	instr = i.TrueCase.byteCode(0, fl.Data().Pass(), cr)
	if instr.Src0() != bytecode.AddrStck && instr.Src0() != bytecode.AddrInv {
		instr |= bytecode.New(bytecode.PUSH)
		*cr.CS = append(*cr.CS, instr)
	}

	jmpTAddr := len(*cr.CS)
	instr = bytecode.New(bytecode.JMP)
	*cr.CS = append(*cr.CS, instr)

	instr = i.FalseCase.byteCode(0, fl.Data().Pass(), cr)
	if instr.Src0() != bytecode.AddrStck && instr.Src0() != bytecode.AddrInv {
		instr |= bytecode.New(bytecode.PUSH)
		*cr.CS = append(*cr.CS, instr)
	}

	// patch jmpf
	(*cr.CS)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, jmpTAddr-jmpfAddr+1)
	// patch jmp
	(*cr.CS)[jmpTAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cr.CS)-jmpTAddr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (w While) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	// PUSH no result
	// JMPF condition                                <-----+ -+
	// POP                                                 |  |
	// loop body                                           |  |
	// if body didn't leave its result on the stack        |  |
	//    PUSH body result                                 |  |
	// JUMP                                           -----+  |
	//                                                <-------+

	discard := fl.Data().Discard

	if !discard {
		ix := len(*cr.DS)
		*cr.DS = append(*cr.DS, value.Nil)
		instr := bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
		*cr.CS = append(*cr.CS, instr)
	}

	condAddr := len(*cr.CS)
	condition := w.Condition.byteCode(0, fl.Data().Pass(), cr)
	jmpfAddr := len(*cr.CS)
	instr := bytecode.New(bytecode.JMPF) | condition
	*cr.CS = append(*cr.CS, instr)

	if !discard {
		instr = bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
	}

	body := w.Body.byteCode(0, fl.Data().Pass(), cr)

	if body.Src0() != bytecode.AddrStck && body.Src0() != bytecode.AddrInv && !discard {
		instr = bytecode.New(bytecode.PUSH) | body
		*cr.CS = append(*cr.CS, instr)
	}

	if body.Src0() == bytecode.AddrStck && discard {
		instr = bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
	}

	instr = bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, condAddr-len(*cr.CS))
	*cr.CS = append(*cr.CS, instr)

	// patch the JMPF
	(*cr.CS)[jmpfAddr] |= bytecode.EncodeSrc(1, bytecode.AddrImm, len(*cr.CS)-jmpfAddr)

	if discard {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrInv, 0)
	}
	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (f For) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
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

	if len(f.Iterators.Elems) != len(f.VarRefs.Elems) {
		panic("for loop number of variables does not match number of iterators")
	}

	ctxID := fl.Data().CtxID
	discard := fl.Data().Discard

	var assignAddr int

	if !discard {
		ix := len(*cr.DS)
		*cr.DS = append(*cr.DS, value.Nil)
		instr := bytecode.New(bytecode.PUSH) | bytecode.EncodeSrc(0, bytecode.AddrDS, ix)
		*cr.CS = append(*cr.CS, instr)
	}

	ccontAddr := len(*cr.CS)
	jmpAddrs := []int{}
	for i, iter := range f.Iterators.Elems {
		// patch previous CCONT
		if i > 0 {
			(*cr.CS)[ccontAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cr.CS)-ccontAddr)
		}
		ccontAddr = len(*cr.CS)
		instr := bytecode.New(bytecode.CCONT) | bytecode.EncodeSrc(1, bytecode.AddrImm, i+ctxID)
		*cr.CS = append(*cr.CS, instr)

		// reset context
		instr = iter.byteCode(0, fl.Data().Pass(flags.WithCtxID(0)), cr)
		if instr.Src0() == bytecode.AddrStck {
			instr = bytecode.New(bytecode.POP)
			*cr.CS = append(*cr.CS, instr)
		}

		instr = bytecode.New(bytecode.DCONT) |
			bytecode.EncodeSrc(0, bytecode.AddrImm, ctxID) |
			bytecode.EncodeSrc(1, bytecode.AddrImm, ctxID+len(f.Iterators.Elems)-1)
		*cr.CS = append(*cr.CS, instr)

		jmpAddrs = append(jmpAddrs, len(*cr.CS))
		instr = bytecode.New(bytecode.JMP)
		*cr.CS = append(*cr.CS, instr)
	}

	var assignBlobJmp int
	if len(f.VarRefs.Elems) > 1 {
		// assign all iterated values after the CCONTs in one blob
		assignAddr = len(*cr.CS)
		for i := len(f.VarRefs.Elems) - 1; i >= 0; i-- {
			vref := f.VarRefs.Elems[i].byteCode(1, fl.Data().Pass(), cr)
			instr := bytecode.New(bytecode.MOV) | vref | bytecode.EncodeSrc(0, bytecode.AddrStck, 0)
			*cr.CS = append(*cr.CS, instr)
		}

		assignBlobJmp = len(*cr.CS)
		instr := bytecode.New(bytecode.JMP)
		*cr.CS = append(*cr.CS, instr)
	}

	// interleave SCONTs with assigns because otherwise if one iterator finishes
	// early we would have the wrong thing on the stack for the loop result
	switchAddr := len(*cr.CS)
	for i, vRef := range f.VarRefs.Elems {
		instr := bytecode.New(bytecode.SCONT) |
			bytecode.EncodeSrc(0, bytecode.AddrImm, ctxID+i)
		*cr.CS = append(*cr.CS, instr)

		if len(f.VarRefs.Elems) <= 1 {
			assignAddr = len(*cr.CS)
		}

		assignee := vRef.byteCode(1, fl.Data().Pass(), cr)
		instr = bytecode.New(bytecode.MOV) | assignee | bytecode.EncodeSrc(0, bytecode.AddrStck, 0)
		*cr.CS = append(*cr.CS, instr)
	}

	if len(f.VarRefs.Elems) > 1 {
		// patch the assign blob jump to jump here
		(*cr.CS)[assignBlobJmp] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cr.CS)-assignBlobJmp)
	}

	if !discard {
		instr := bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
	}

	body := f.Body.byteCode(0, fl.Data().Pass(
		flags.WithInFor(true),
		flags.WithCtxID(ctxID+len(f.VarRefs.Elems)),
		flags.WithCtxLo(ctxID),
		flags.WithCtxHi(ctxID+len(f.VarRefs.Elems)-1)), cr)

	if body.Src0() != bytecode.AddrStck && body.Src0() != bytecode.AddrInv && !discard {
		instr := bytecode.New(bytecode.PUSH) | body
		*cr.CS = append(*cr.CS, instr)
	}

	if body.Src0() == bytecode.AddrStck && discard {
		instr := bytecode.New(bytecode.POP)
		*cr.CS = append(*cr.CS, instr)
	}

	instr := bytecode.New(bytecode.JMP) | bytecode.EncodeSrc(0, bytecode.AddrImm, switchAddr-len(*cr.CS))
	*cr.CS = append(*cr.CS, instr)

	// patch jumps
	for _, jmpAddr := range jmpAddrs {
		(*cr.CS)[jmpAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, len(*cr.CS)-jmpAddr)
	}

	// patch ccont
	(*cr.CS)[ccontAddr] |= bytecode.EncodeSrc(0, bytecode.AddrImm, assignAddr-ccontAddr)

	if discard {
		return bytecode.EncodeSrc(srcsel, bytecode.AddrInv, 0)
	}
	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IndexAt) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	ary := i.Ary.byteCode(1, fl.Data().Pass(), cr)
	at := i.At.byteCode(0, fl.Data().Pass(), cr)
	instr := bytecode.New(bytecode.IX1) | ary | at

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (i IndexFromTo) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	// nbcd := bcData{inFor: bcd.inFor, forbidTemp: bcd.forbidTemp, opDepth: 0}
	ary := i.Ary.byteCode(2, fl.Data().Pass(flags.WithOpDepth(0)), cr)
	from := i.From.byteCode(1, fl.Data().Pass(flags.WithOpDepth(0)), cr)
	to := i.To.byteCode(0, fl.Data().Pass(flags.WithOpDepth(0)), cr)

	instr := bytecode.New(bytecode.IX2) | ary | from | to

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (r Read) byteCode(srcsel int, _ flags.Pass, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.READ)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (w Write) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.WRITE) | w.Value.byteCode(0, fl.Data().Pass(), cr)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (a Aton) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.ATON) | a.Value.byteCode(0, fl.Data().Pass(), cr)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (t Toa) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.TOA) | t.Value.byteCode(0, fl.Data().Pass(), cr)
	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

func (e Exit) byteCode(srcsel int, fl flags.Pass, cr compResult) bytecode.Type {
	instr := bytecode.New(bytecode.EXIT) | e.Value.byteCode(0, fl.Data().Pass(), cr)

	*cr.CS = append(*cr.CS, instr)

	return bytecode.EncodeSrc(srcsel, bytecode.AddrStck, 0)
}

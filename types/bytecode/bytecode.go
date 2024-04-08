// Package bytecode contains the bytecode instructions.
//
// It's not really a "byte" code atm. but a fixed size 64 bit instruction set.
// This might be optimised in the future.
package bytecode

import "fmt"

// Type is a fixed size 64 bit instruction.
type Type uint64

// Instruction layout.
const (
	OpcodeHi   = 63
	OpcodeLo   = 57
	Src2Hi     = 56
	Src2Lo     = 54
	Src1Hi     = 53
	Src1Lo     = 51
	Src0Hi     = 50
	Src0Lo     = 48
	Src2AddrHi = 47
	Src2AddrLo = 32
	Src1AddrHi = 31
	Src1AddrLo = 16
	Src0AddrHi = 15
	Src0AddrLo = 0
)

const (
	SrcChanWidth      = 16
	SrcChanSignExtend = 0xffffffffffff0000
)

// Source addressing.
const (
	AddrInv  = iota
	AddrImm  // immediate
	AddrGbl  // global variable
	AddrLcl  // local variable
	AddrCls  // closure variable
	AddrStck // stack
	AddrTmp  // temp register
	AddrDS   // data segment
)

// OpCode is the instruction code.
type OpCode uint64

const TempFlag = 1 << (OpcodeHi - OpcodeLo)

// Instruction set.
//
//go:generate go run github.com/dmarkham/enumer -type=OpCode
const (
	NOP = OpCode(iota)

	PUSH // PUSH pushes src0
	POP  // POP pops stack, result goes nowhere
	MOV  // MOV dst, src0

	ADD // ADD pushes src1+src0
	SUB // SUB pushes src1-src0
	MUL // MUL pushes src1*src0
	DIV // DIV pushes src1/src0
	MOD // MOD pushes src1%src0

	INC // INC increments src0

	NOT // NOT pushes !src0 (or type error if not bool)
	AND // AND pushes src1&src0
	OR  // OR pushes src1|src0

	LT // LT pushes src1<src0
	GT // GT pushes src1>src0
	LE // LE pushes src1<=src0
	GE // GE pushes src1>=src0
	EQ // EQ pushes src1==src0
	NE // NE pushes src1!=src0

	LSH  // LSH pushes src1<<src0
	RSH  // RSH pushes src1>>src0
	FLIP // FLIP pushes ~src0

	IX1 // IX1 pushes src1[src0]
	IX2 // IX2 pushes src2[src1:src0]

	LEN // LEN pushes the length of src0
	ARR // ARR pushes the src0 + [src1]

	JMP  // JMP jumps relative to ip + src0
	JMPF // JMPF jumps relative to ip+src1 if src0 is false or ip+src2 if src0 is not bool
	FUNC // FUNC pushes a function value with code address from src0, parameter count src1, local count src2
	CALL // CALL calls src0 with argument cnt src1
	RET  // RET returns from a function call pushing src0 after rolling back the stack

	// CONTFRM pushes a new frame of memory context on the context stack.
	CONTFRM
	// CCONT jumps relative to ip + src0 in current context, appends old context
	// to the current context frame, and switches to a new context that continues
	// from old ip.
	CCONT
	// DCONT restores previous context switches memory, but doesn't switch ip.
	DCONT
	// RCONT deletes last frame of memory contexts.
	RCONT
	// SCONT switches to the context id src0; id is the index of the context in
	// the current frame.
	SCONT
	// YIELD pushes src0 in the current context, swaps the current context with
	// its parent, and pushes src0 in the new context.
	YIELD

	READ  // READ builtin
	WRITE // WRITE builtin
	ATON  // ATON converts src0 to a number and pushes it
	TOA   // TOA converts src0 to a string and pushes it
	EXIT  // EXIT terminates the program

	PUSHTMP = OpCode(TempFlag | PUSH) // PUSHTMP pushes the temp register

	ADDTMP = OpCode(TempFlag | ADD) // ADDTMP adds src0 to the temp register
	SUBTMP = OpCode(TempFlag | SUB) // SUBTMP subtracts src0 from the temp register
	MULTMP = OpCode(TempFlag | MUL) // MULTMP multiplies src0 by the temp register
	DIVTMP = OpCode(TempFlag | DIV) // DITMP divides src0 by the temp register
	MODTMP = OpCode(TempFlag | MOD) // MODTMP calculates temp % src0 in the temp register

	NOTTMP = OpCode(TempFlag | NOT) // NOTTMP calculates !temp in the temp register
	ANDTMP = OpCode(TempFlag | AND) // ANDTMP calculates temp & src0 in the temp register
	ORTMP  = OpCode(TempFlag | OR)  // ORTMP calculates temp | src0 in the temp register

	LTTMP = OpCode(TempFlag | LT) // LTTMP calculates temp < src0 in the temp register
	GTTMP = OpCode(TempFlag | GT) // GTTMP calculates temp > src0 in the temp register
	LETMP = OpCode(TempFlag | LE) // LETMP calculates temp <= src0 in the temp register
	GETMP = OpCode(TempFlag | GE) // GETMP calculates temp >= src0 in the temp register
	EQTMP = OpCode(TempFlag | EQ) // EQTMP calculates temp == src0 in the temp register
	NETMP = OpCode(TempFlag | NE) // NETMP calculates temp!= src0 in the temp register

	LSHTMP  = OpCode(TempFlag | LSH)  // LSHTMP calculates temp << src0 in the temp register
	RSHTMP  = OpCode(TempFlag | RSH)  // RHSTMP calculates temp >> src0 in the temp register
	FLIPTMP = OpCode(TempFlag | FLIP) // FLIPTMP calculates ~temp in the temp register

	LENTMP = OpCode(TempFlag | LEN) // LEN pushes the length of src0
)

// New creates a new instruction.
func New(op OpCode) Type {
	op &= (1 << ((OpcodeHi - OpcodeLo) + 1)) - 1

	return Type(uint64(op) << OpcodeLo)
}

// EncodeSrc encodes an instruction operand.
//
// srcsel specifies which operand is encoded (0/1/2). src specifies where the
// operand is coming from, has to be one of the source addressing values.
// srcAddr specifies the source address, or immediate value for instruction
// encoded integers.
func EncodeSrc(srcsel int, src uint64, srcAddr int) Type {
	if srcAddr <= -(1<<SrcChanWidth) || srcAddr >= (1<<SrcChanWidth) {
		panic("srcAddr out of range")
	}
	addr := uint64(srcAddr)
	switch srcsel {
	case 0:
		src &= (1 << ((Src0Hi - Src0Lo) + 1)) - 1
		addr &= (1 << ((Src0AddrHi - Src0AddrLo) + 1)) - 1

		return Type(src<<Src0Lo | uint64(addr)<<Src0AddrLo)
	case 1:

		src &= (1 << ((Src1Hi - Src1Lo) + 1)) - 1
		addr &= (1 << ((Src1AddrHi - Src1AddrLo) + 1)) - 1

		return Type(src<<Src1Lo | uint64(addr)<<Src1AddrLo)

	case 2:
		src &= (1 << ((Src2Hi - Src2Lo) + 1)) - 1
		addr &= (1 << ((Src2AddrHi - Src2AddrLo) + 1)) - 1

		return Type(src<<Src2Lo | uint64(addr)<<Src2AddrLo)

	default:
		panic("wrong srcsel")
	}
}

// String provides Stringer implementation for Type.
func (b Type) String() string {
	oc := b.OpCode()
	src0 := srcString(b.Src0(), b.Src0Addr())
	src1 := srcString(b.Src1(), b.Src1Addr())
	src2 := srcString(b.Src2(), b.Src2Addr())

	return fmt.Sprintf("%#016X : %v %s%s%s", uint64(b), oc, src2, src1, src0)
}

func srcString(src uint64, addr int) string {
	switch src {
	case AddrDS:
		return fmt.Sprintf("DS[%d] ", addr)
	case AddrCls:
		return fmt.Sprintf("CLS[%d] ", addr)
	case AddrLcl:
		return fmt.Sprintf("LCL[%d] ", addr)
	case AddrGbl:
		return fmt.Sprintf("GBL[%d] ", addr)
	case AddrStck:
		return "STCK "
	case AddrTmp:
		return "TMP "
	case AddrImm:
		return fmt.Sprintf("%d ", addr)
	default:
		return ""
	}
}

// Src returns the part of the instruction that encodes the srcsel operand.
func (b Type) Src(srcsel int) uint64 {
	switch srcsel {
	case 0:
		return b.Src0()
	case 1:
		return b.Src1()
	default:
		panic("wrong srcsel")
	}
}

// OpCode returns the opcode of the instruction.
func (b Type) OpCode() OpCode {
	return OpCode(b>>OpcodeLo) & ((1 << (OpcodeHi - OpcodeLo + 1)) - 1)
}

// Src0 returns the part of the instruction that encodes the src0 operand.
func (b Type) Src0() uint64 {
	return uint64((b >> Src0Lo) & ((1 << (Src0Hi - Src0Lo + 1)) - 1))
}

// Src1 returns the part of the instruction that encodes the src1 operand.
func (b Type) Src1() uint64 {
	return uint64((b >> Src1Lo) & ((1 << (Src1Hi - Src1Lo + 1)) - 1))
}

// Src2 returns the part of the instruction that encodes the src2 operand.
func (b Type) Src2() uint64 {
	return uint64((b >> Src2Lo) & ((1 << (Src2Hi - Src2Lo + 1)) - 1))
}

// Src0Addr returns the src0 address or the immediate value of src0.
func (b Type) Src0Addr() int {
	return convImm((b >> Src0AddrLo) & ((1 << (Src0AddrHi - Src0AddrLo + 1)) - 1))
}

// Src1Addr returns the src1 address or the immediate value of src1.
func (b Type) Src1Addr() int {
	return convImm((b >> Src1AddrLo) & ((1 << (Src1AddrHi - Src1AddrLo + 1)) - 1))
}

// Src2Addr returns the src2 address or the immediate value of src2.
func (b Type) Src2Addr() int {
	return convImm((b >> Src2AddrLo) & ((1 << (Src2AddrHi - Src2AddrLo + 1)) - 1))
}

func convImm(n Type) int {
	if n&(1<<(SrcChanWidth-1)) != 0 {
		n |= SrcChanSignExtend
	}
	return int(n)
}

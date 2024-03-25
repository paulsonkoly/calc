package bytecode

import "fmt"

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
	AddrDS   // data segment
)

type OpCode uint64

// Instruction set.
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

	// CCONT jumps relative to ip + src0 in current context, saves context, and
	// switches to a new context that continues from old ip.
	CCONT
	// DCONT restores previous context switches memory, but doesn't switch ip.
	DCONT
	// RCONT removes last memory context saved.
	RCONT
	// SCONT swaps the current context with the last saved.
	SCONT
	// YIELD pushes src0 in the current context, swaps the current context with
	// the saved, and pushes src0 in the new context.
	YIELD

	READ  // READ builtin
	WRITE // WRITE builtin
	ATON  // ATON converts src0 to a number and pushes it
	TOA   // TOA converts src0 to a string and pushes it
	ERROR // ERROR pushes an error message
	EXIT  // EXIT terminates the program
)

func New(op OpCode) Type {
	if op < NOP || op > EXIT {
		panic("op out of range")
	}
	op &= (1 << ((OpcodeHi - OpcodeLo) + 1)) - 1

	return Type(uint64(op) << OpcodeLo)
}

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
	case AddrImm:
		return fmt.Sprintf("%d ", addr)
	default:
		return ""
	}
}

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

func (b Type) OpCode() OpCode {
	return OpCode(b>>OpcodeLo) & ((1 << (OpcodeHi - OpcodeLo + 1)) - 1)
}

func (b Type) Src0() uint64 {
	return uint64((b >> Src0Lo) & ((1 << (Src0Hi - Src0Lo + 1)) - 1))
}

func (b Type) Src1() uint64 {
	return uint64((b >> Src1Lo) & ((1 << (Src1Hi - Src1Lo + 1)) - 1))
}

func (b Type) Src2() uint64 {
	return uint64((b >> Src2Lo) & ((1 << (Src2Hi - Src2Lo + 1)) - 1))
}

func (b Type) Src0Addr() int {
	return convImm((b >> Src0AddrLo) & ((1 << (Src0AddrHi - Src0AddrLo + 1)) - 1))
}

func (b Type) Src1Addr() int {
	return convImm((b >> Src1AddrLo) & ((1 << (Src1AddrHi - Src1AddrLo + 1)) - 1))
}

func (b Type) Src2Addr() int {
	return convImm((b >> Src2AddrLo) & ((1 << (Src2AddrHi - Src2AddrLo + 1)) - 1))
}

func convImm(n Type) int {
	if n&(1<<(SrcChanWidth-1)) != 0 {
		n |= SrcChanSignExtend
	}
	return int(n)
}

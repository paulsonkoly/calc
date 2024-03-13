package bytecode

import "fmt"

type Type uint64

// Instruction layout
const (
	OPCODE_HI    = 63
	OPCODE_LO    = 57
	SRC2_HI      = 56
	SRC2_LO      = 54
	SRC1_HI      = 53
	SRC1_LO      = 51
	SRC0_HI      = 50
	SRC0_LO      = 48
	SRC2_ADDR_HI = 47
	SRC2_ADDR_LO = 32
	SRC1_ADDR_HI = 31
	SRC1_ADDR_LO = 16
	SRC0_ADDR_HI = 15
	SRC0_ADDR_LO = 0
)

const (
	SRC_CHAN_WIDTH       = 16
	SRC_CHAN_SIGN_EXTEND = 0xffffffffffff0000
)

// Source addressing
const (
	ADDR_INV  = iota
	ADDR_IMM  // immediate
	ADDR_GBL  // global variable
	ADDR_LCL  // local variable
	ADDR_CLS  // closure variable
	ADDR_STCK // stack
	ADDR_DS   // data segment
)

type OpCode uint64

// Instruction set
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

	IX1 // IX1 pushes src1[src0]
	IX2 // IX2 pushes src2[src1:src0]

	LEN // LEN pushes the length of src0
	ARR // ARR pushes the src0 + [src1]

	JMP  // JMP jumps relative to ip + src0
	JMPF // JMPF jumps relative to ip+src1 if src0 is false or ip+src2 if src0 is not bool
	FUNC // FUNC pushes a function value with code address from src0, parameter count src1, local count src2
	CALL // CALL calls src0 with argument cnt src1
	RET  // RET returns from a function call pushing src0 after rolling back the stack

  // CCONT creates a new execution context cloning the current one (only
  // last closure and local frames + global) and pushes it to the context stack
	CCONT

  // PCONT pops the last execution context and does not switch
  // master co-routine return path
	PCONT 

  // DCONT pops the last execution context and switches with current but does not jump
  // slave co-routine return path
	DCONT 

  // SCONT switches the current context with the top of the context stack
  // setting IP to the other context IP (jumps). If src0 is valid it is read
  // before the switch and pushed on the stack after the switch.
  // co-routine yield
	SCONT 

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
	op &= (1 << ((OPCODE_HI - OPCODE_LO) + 1)) - 1

	return Type(uint64(op) << OPCODE_LO)
}

func EncodeSrc(srcsel int, src uint64, srcAddr int) Type {
	if srcAddr <= -(1<<SRC_CHAN_WIDTH) || srcAddr >= (1<<SRC_CHAN_WIDTH) {
		panic("srcAddr out of range")
	}
	addr := uint64(srcAddr)
	switch srcsel {
	case 0:
		src &= (1 << ((SRC0_HI - SRC0_LO) + 1)) - 1
		addr &= (1 << ((SRC0_ADDR_HI - SRC0_ADDR_LO) + 1)) - 1

		return Type(src<<SRC0_LO | uint64(addr)<<SRC0_ADDR_LO)
	case 1:

		src &= (1 << ((SRC1_HI - SRC1_LO) + 1)) - 1
		addr &= (1 << ((SRC1_ADDR_HI - SRC1_ADDR_LO) + 1)) - 1

		return Type(src<<SRC1_LO | uint64(addr)<<SRC1_ADDR_LO)

	case 2:
		src &= (1 << ((SRC2_HI - SRC2_LO) + 1)) - 1
		addr &= (1 << ((SRC2_ADDR_HI - SRC2_ADDR_LO) + 1)) - 1

		return Type(src<<SRC2_LO | uint64(addr)<<SRC2_ADDR_LO)

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
	case ADDR_DS:
		return fmt.Sprintf("DS[%d] ", addr)
	case ADDR_CLS:
		return fmt.Sprintf("CLS[%d] ", addr)
	case ADDR_LCL:
		return fmt.Sprintf("LCL[%d] ", addr)
	case ADDR_GBL:
		return fmt.Sprintf("GBL[%d] ", addr)
	case ADDR_STCK:
		return "STCK "
	case ADDR_IMM:
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
	return OpCode(b>>OPCODE_LO) & ((1 << (OPCODE_HI - OPCODE_LO + 1)) - 1)
}

func (b Type) Src0() uint64 {
	return uint64((b >> SRC0_LO) & ((1 << (SRC0_HI - SRC0_LO + 1)) - 1))
}

func (b Type) Src1() uint64 {
	return uint64((b >> SRC1_LO) & ((1 << (SRC1_HI - SRC1_LO + 1)) - 1))
}

func (b Type) Src2() uint64 {
	return uint64((b >> SRC2_LO) & ((1 << (SRC2_HI - SRC2_LO + 1)) - 1))
}

func (b Type) Src0Addr() int {
	return convImm((b >> SRC0_ADDR_LO) & ((1 << (SRC0_ADDR_HI - SRC0_ADDR_LO + 1)) - 1))
}

func (b Type) Src1Addr() int {
	return convImm((b >> SRC1_ADDR_LO) & ((1 << (SRC1_ADDR_HI - SRC1_ADDR_LO + 1)) - 1))
}

func (b Type) Src2Addr() int {
	return convImm((b >> SRC2_ADDR_LO) & ((1 << (SRC2_ADDR_HI - SRC2_ADDR_LO + 1)) - 1))
}

func convImm(n Type) int {
	if n&(1<<(SRC_CHAN_WIDTH-1)) != 0 {
		n |= SRC_CHAN_SIGN_EXTEND
	}
	return int(n)
}

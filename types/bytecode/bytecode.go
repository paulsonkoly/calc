package bytecode

import "fmt"

type Type uint64

// Instruction layout
const (
	OPCODE_HI    = 63
	OPCODE_LO    = 56
	SRC1_HI      = 53
	SRC1_LO      = 51
	SRC0_HI      = 50
	SRC0_LO      = 48
	SRC1_ADDR_HI = 47
	SRC1_ADDR_LO = 24
	SRC0_ADDR_HI = 23
	SRC0_ADDR_LO = 0
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

	PUSH
	POP
	MOV

	ADD
	SUB
	MUL
	DIV
	MOD

	NOT
	AND
	OR

	LT
	GT
	LE
	GE
	EQ
	NE

	IX1
	IX2

	LEN
  ARR

	JMP
	JMPF
  FUNC

	READ
	WRITE
	ATON
	TOA
	ERROR
	EXIT
)

func NewByteCode(op OpCode, src1 uint64, src1addr int, src0 uint64, src0addr int) Type {
	// TODO mask overflows
	op &= (1 << ((OPCODE_HI - OPCODE_LO) + 1)) - 1
	src1 &= (1 << ((SRC1_HI - SRC1_LO) + 1)) - 1
	src0 &= (1 << ((SRC0_HI - SRC0_LO) + 1)) - 1
	src1addr &= (1 << ((SRC1_ADDR_HI - SRC1_ADDR_LO) + 1)) - 1
	src0addr &= (1 << ((SRC0_ADDR_HI - SRC0_ADDR_LO) + 1)) - 1

	return Type(uint64(op)<<OPCODE_LO |
    src1<<SRC1_LO |
    src0<<SRC0_LO |
    uint64(src1addr)<<SRC1_ADDR_LO |
    uint64(src0addr)<<SRC0_ADDR_LO)
}

func (b Type) String() string {
	oc := b.OpCode()
	src0 := srcString(b.Src0(), b.Src0Addr())
	src1 := srcString(b.Src1(), b.Src1Addr())

	return fmt.Sprintf("%#016X : %v %s %s", uint64(b), oc, src1, src0)
}

func srcString(src uint64, addr int) string {
	switch src {
	case ADDR_DS:
		return fmt.Sprintf("DS[%d]", addr)
	case ADDR_CLS:
		return fmt.Sprintf("CLS[%d]", addr)
	case ADDR_LCL:
		return fmt.Sprintf("LCL[%d]", addr)
	case ADDR_GBL:
		return fmt.Sprintf("GBL[%d]", addr)
	case ADDR_STCK:
		return "STCK"
	case ADDR_IMM:
		return fmt.Sprintf("%d", addr)
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

func (b Type) Src1() uint64 {
	return uint64((b >> SRC1_LO) & ((1 << (SRC1_HI - SRC1_LO + 1)) - 1))
}

func (b Type) Src0() uint64 {
	return uint64((b >> SRC0_LO) & ((1 << (SRC0_HI - SRC0_LO + 1)) - 1))
}

func (b Type) Src0Addr() int {
  return convImm((b >> SRC0_ADDR_LO) & ((1 << (SRC0_ADDR_HI - SRC0_ADDR_LO + 1)) - 1))
}

func (b Type) Src1Addr() int {
  return convImm((b >> SRC1_ADDR_LO) & ((1 << (SRC1_ADDR_HI - SRC1_ADDR_LO + 1)) - 1))
}

func convImm(n Type) int {
  if n & (1 << 23) != 0 { // TODO constants
    n |= (0xffffffffff000000)
  }
  return int(n)
}

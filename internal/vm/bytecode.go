package vm

import "compiler/internal/lex"

type OpCode uint8

const (
	OpPush OpCode = iota
	OpPop
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg
	OpCallPrintln
	OpCallReadInt
	OpCallAbs
	OpCallMin
	OpCallMax
)

type Instruction struct {
	Op  OpCode
	Arg int64
	Pos lex.SrcPos
}

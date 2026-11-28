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
	OpLoadGlobal
	OpStoreGlobal
	OpLoadLocal
	OpStoreLocal
	OpCall
	OpReturn
	OpReturnVal
	OpHalt
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

type FuncInfo struct {
	Name      string
	Addr      int
	Arity     int
	NumLocals int
}

type Program struct {
	Code       []Instruction
	Functions  []FuncInfo
	NumGlobals int
}

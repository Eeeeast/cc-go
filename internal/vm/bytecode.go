package vm

import "compiler/internal/lex"

type OpCode uint8

const (
	OpPush        OpCode = iota // arg: int64
	OpPop                       // сброс результата выражения-инструкции
	OpAdd                       // +
	OpSub                       // -
	OpMul                       // *
	OpDiv                       // /
	OpMod                       // %
	OpNeg                       // унарный -
	OpCallPrintln               // fmt.Println(x) -> кладёт Void
	OpCallReadInt               // fmt.ReadInt()  -> кладёт Int
	OpCallAbs                   // abs(x)         -> кладёт Int
	OpCallMin                   // min(a, b)      -> кладёт Int
	OpCallMax                   // max(a, b)      -> кладёт Int
)

type Instruction struct {
	Op  OpCode
	Arg int64
	Pos lex.SrcPos
}

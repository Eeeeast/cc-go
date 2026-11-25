package vm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"compiler/internal/lex"
)

type RuntimeError struct {
	Pos lex.SrcPos
	Msg string
}

func (e *RuntimeError) Error() string {
	if e.Pos.SrcName == "" {
		return e.Msg
	}
	if e.Pos.LineNo == 0 {
		return fmt.Sprintf("%s: %s", e.Pos.SrcName, e.Msg)
	}
	if e.Pos.ColNo == 0 {
		return fmt.Sprintf("%s:%d: %s", e.Pos.SrcName, e.Pos.LineNo, e.Msg)
	}
	return fmt.Sprintf("%s:%d:%d: %s", e.Pos.SrcName, e.Pos.LineNo, e.Pos.ColNo, e.Msg)
}

func newRuntimeError(pos lex.SrcPos, format string, args ...any) error {
	return &RuntimeError{
		Pos: pos,
		Msg: fmt.Sprintf(format, args...),
	}
}

type valueKind uint8

const (
	valInt valueKind = iota
	valUnit
)

type value struct {
	kind valueKind
	num  int64
}

func intVal(n int64) value { return value{kind: valInt, num: n} }
func unitVal() value       { return value{kind: valUnit} }

func (k valueKind) String() string {
	switch k {
	case valInt:
		return "int"
	case valUnit:
		return "()"
	default:
		return fmt.Sprintf("valueKind(%d)", k)
	}
}

type VM struct {
	in    *bufio.Reader
	out   io.Writer
	stack []value
}

func New(in io.Reader, out io.Writer) *VM {
	return &VM{
		in:  bufio.NewReader(in),
		out: out,
	}
}

func (v *VM) Run(code []Instruction) error {
	v.stack = v.stack[:0]

	for _, ins := range code {
		switch ins.Op {
		case OpPush:
			v.push(intVal(ins.Arg))

		case OpPop:
			_, err := v.pop(ins.Pos)
			if err != nil {
				return err
			}

		case OpAdd:
			b, a, err := v.popTwoInts(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedAdd(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(res))

		case OpSub:
			b, a, err := v.popTwoInts(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedSub(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(res))

		case OpMul:
			b, a, err := v.popTwoInts(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedMul(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(res))

		case OpDiv:
			b, a, err := v.popTwoInts(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedDiv(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(res))

		case OpMod:
			b, a, err := v.popTwoInts(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedMod(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(res))

		case OpNeg:
			a, err := v.popInt(ins.Pos)
			if err != nil {
				return err
			}
			if a == math.MinInt64 {
				return newRuntimeError(ins.Pos, "integer overflow in unary negation")
			}
			v.push(intVal(-a))

		case OpCallPrintln:
			x, err := v.popInt(ins.Pos)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(v.out, "%d\n", x)
			if err != nil {
				return newRuntimeError(ins.Pos, "output error: %v", err)
			}
			v.push(unitVal())

		case OpCallReadInt:
			x, err := v.readInt(ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(x))

		case OpCallAbs:
			x, err := v.popInt(ins.Pos)
			if err != nil {
				return err
			}
			if x == math.MinInt64 {
				return newRuntimeError(ins.Pos, "integer overflow in abs")
			}
			if x < 0 {
				x = -x
			}
			v.push(intVal(x))

		case OpCallMin:
			b, a, err := v.popTwoInts(ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(min(a, b)))

		case OpCallMax:
			b, a, err := v.popTwoInts(ins.Pos)
			if err != nil {
				return err
			}
			v.push(intVal(max(a, b)))

		default:
			return newRuntimeError(ins.Pos, "unknown opcode %d", ins.Op)
		}
	}

	return nil
}

func (v *VM) push(val value) {
	v.stack = append(v.stack, val)
}

func (v *VM) pop(pos lex.SrcPos) (value, error) {
	if len(v.stack) == 0 {
		return value{}, newRuntimeError(pos, "stack underflow")
	}
	n := len(v.stack) - 1
	val := v.stack[n]
	v.stack = v.stack[:n]
	return val, nil
}

func (v *VM) popInt(pos lex.SrcPos) (int64, error) {
	val, err := v.pop(pos)
	if err != nil {
		return 0, err
	}
	if val.kind != valInt {
		return 0, newRuntimeError(pos, "epected int, got %s", val.kind)
	}
	return val.num, nil
}

// popTwoInts снимает со стека сначала правый `b`, затем левый `a` операнд.
func (v *VM) popTwoInts(pos lex.SrcPos) (b int64, a int64, err error) {
	b, err = v.popInt(pos)
	if err != nil {
		return 0, 0, err
	}
	a, err = v.popInt(pos)
	if err != nil {
		return 0, 0, err
	}
	return b, a, nil
}

func (v *VM) readInt(pos lex.SrcPos) (int64, error) {
	line, err := v.in.ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || len(line) == 0) {
		return 0, newRuntimeError(pos, "input error: unexpected end of input")
	}

	trimmed := strings.Trim(line, " \t\r\n")
	if trimmed == "" {
		return 0, newRuntimeError(pos, "input error: expected integer, got empty line")
	}

	n, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, newRuntimeError(pos, "input error: invalid 64-bit integer %q", trimmed)
	}
	return n, nil
}

func checkedAdd(a, b int64, pos lex.SrcPos) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, newRuntimeError(pos, "integer overflow in addition")
	}
	return a + b, nil
}

func checkedSub(a, b int64, pos lex.SrcPos) (int64, error) {
	if (b < 0 && a > math.MaxInt64+b) || (b > 0 && a < math.MinInt64+b) {
		return 0, newRuntimeError(pos, "integer overflow in subtraction")
	}
	return a - b, nil
}

func checkedMul(a, b int64, pos lex.SrcPos) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if (a == math.MinInt64 && b == -1) || (b == math.MinInt64 && a == -1) {
		return 0, newRuntimeError(pos, "integer overflow in multiplication")
	}
	res := a * b
	if res/a != b {
		return 0, newRuntimeError(pos, "integer overflow in multiplication")
	}
	return res, nil
}

func checkedDiv(a, b int64, pos lex.SrcPos) (int64, error) {
	if b == 0 {
		return 0, newRuntimeError(pos, "division by zero")
	}
	if a == math.MinInt64 && b == -1 {
		return 0, newRuntimeError(pos, "integer overflow in division")
	}
	return a / b, nil
}

func checkedMod(a, b int64, pos lex.SrcPos) (int64, error) {
	if b == 0 {
		return 0, newRuntimeError(pos, "modulo by zero")
	}
	if a == math.MinInt64 && b == -1 {
		return 0, nil
	}
	return a % b, nil
}

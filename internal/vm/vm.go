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

const maxCallDepth = 10000

type RuntimeError struct {
	Pos lex.SrcPos
	Msg string
}

func (e *RuntimeError) Error() string {
	return e.Pos.FormatError(e.Msg)
}

func newRuntimeError(pos lex.SrcPos, format string, args ...any) error {
	return &RuntimeError{
		Pos: pos,
		Msg: fmt.Sprintf(format, args...),
	}
}

type callFrame struct {
	retIP int
	base  int
}

type VM struct {
	in      *bufio.Reader
	out     io.Writer
	globals []int64
	stack   []int64
	frames  []callFrame
	ip      int
}

func New(in io.Reader, out io.Writer) *VM {
	return &VM{
		in:  bufio.NewReader(in),
		out: out,
	}
}

func (v *VM) Run(prog *Program) error {
	v.globals = make([]int64, prog.NumGlobals)
	v.stack = v.stack[:0]
	v.frames = v.frames[:0]
	v.ip = 0

	code := prog.Code
	for v.ip < len(code) {
		ins := code[v.ip]
		v.ip++

		switch ins.Op {
		case OpPush:
			v.push(ins.Arg)

		case OpPop:
			if _, err := v.pop(ins.Pos); err != nil {
				return err
			}

		case OpAdd:
			b, a, err := v.popTwo(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedAdd(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(res)

		case OpSub:
			b, a, err := v.popTwo(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedSub(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(res)

		case OpMul:
			b, a, err := v.popTwo(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedMul(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(res)

		case OpDiv:
			b, a, err := v.popTwo(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedDiv(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(res)

		case OpMod:
			b, a, err := v.popTwo(ins.Pos)
			if err != nil {
				return err
			}
			res, err := checkedMod(a, b, ins.Pos)
			if err != nil {
				return err
			}
			v.push(res)

		case OpNeg:
			a, err := v.pop(ins.Pos)
			if err != nil {
				return err
			}
			if a == math.MinInt64 {
				return newRuntimeError(ins.Pos, "integer overflow in unary negation")
			}
			v.push(-a)

		case OpLoadGlobal:
			idx := int(ins.Arg)
			if idx < 0 || idx >= len(v.globals) {
				return newRuntimeError(ins.Pos, "invalid global slot %d", idx)
			}
			v.push(v.globals[idx])

		case OpStoreGlobal:
			val, err := v.pop(ins.Pos)
			if err != nil {
				return err
			}
			idx := int(ins.Arg)
			if idx < 0 || idx >= len(v.globals) {
				return newRuntimeError(ins.Pos, "invalid global slot %d", idx)
			}
			v.globals[idx] = val

		case OpLoadLocal:
			if len(v.frames) == 0 {
				return newRuntimeError(ins.Pos, "local access outside function")
			}
			idx := v.frames[len(v.frames)-1].base + int(ins.Arg)
			v.push(v.stack[idx])

		case OpStoreLocal:
			if len(v.frames) == 0 {
				return newRuntimeError(ins.Pos, "local store outside function")
			}
			val, err := v.pop(ins.Pos)
			if err != nil {
				return err
			}
			idx := v.frames[len(v.frames)-1].base + int(ins.Arg)
			v.stack[idx] = val

		case OpCall:
			funcIdx := int(ins.Arg)
			if funcIdx < 0 || funcIdx >= len(prog.Functions) {
				return newRuntimeError(ins.Pos, "invalid function index %d", funcIdx)
			}
			if len(v.frames) >= maxCallDepth {
				return newRuntimeError(ins.Pos, "stack overflow: max call depth exceeded")
			}
			fn := prog.Functions[funcIdx]
			if len(v.stack) < fn.Arity {
				return newRuntimeError(ins.Pos, "stack underflow on call to %s", fn.Name)
			}
			base := len(v.stack) - fn.Arity
			for i := fn.Arity; i < fn.NumLocals; i++ {
				v.stack = append(v.stack, 0)
			}
			v.frames = append(v.frames, callFrame{
				retIP: v.ip,
				base:  base,
			})
			v.ip = fn.Addr

		case OpReturn:
			if len(v.frames) == 0 {
				return newRuntimeError(ins.Pos, "return outside function")
			}
			frame := v.frames[len(v.frames)-1]
			v.frames = v.frames[:len(v.frames)-1]
			v.stack = v.stack[:frame.base]
			v.ip = frame.retIP

		case OpReturnVal:
			if len(v.frames) == 0 {
				return newRuntimeError(ins.Pos, "return outside function")
			}
			retVal, err := v.pop(ins.Pos)
			if err != nil {
				return err
			}
			frame := v.frames[len(v.frames)-1]
			v.frames = v.frames[:len(v.frames)-1]
			v.stack = v.stack[:frame.base]
			v.push(retVal)
			v.ip = frame.retIP

		case OpHalt:
			return nil

		case OpCallPrintln:
			x, err := v.pop(ins.Pos)
			if err != nil {
				return err
			}
			if _, err = fmt.Fprintf(v.out, "%d\n", x); err != nil {
				return newRuntimeError(ins.Pos, "output error: %v", err)
			}

		case OpCallReadInt:
			x, err := v.readInt(ins.Pos)
			if err != nil {
				return err
			}
			v.push(x)

		case OpCallAbs:
			x, err := v.pop(ins.Pos)
			if err != nil {
				return err
			}
			if x == math.MinInt64 {
				return newRuntimeError(ins.Pos, "integer overflow in abs")
			}
			if x < 0 {
				x = -x
			}
			v.push(x)

		case OpCallMin:
			res, err := v.popVariadicFold(int(ins.Arg), ins.Pos, func(a, b int64) int64 {
				return min(a, b)
			})
			if err != nil {
				return err
			}
			v.push(res)

		case OpCallMax:
			res, err := v.popVariadicFold(int(ins.Arg), ins.Pos, func(a, b int64) int64 {
				return max(a, b)
			})
			if err != nil {
				return err
			}
			v.push(res)

		default:
			return newRuntimeError(ins.Pos, "unknown opcode %d", ins.Op)
		}
	}

	return nil
}

func (v *VM) push(val int64) {
	v.stack = append(v.stack, val)
}

func (v *VM) pop(pos lex.SrcPos) (int64, error) {
	if len(v.stack) == 0 {
		return 0, newRuntimeError(pos, "stack underflow")
	}
	n := len(v.stack) - 1
	val := v.stack[n]
	v.stack = v.stack[:n]
	return val, nil
}

func (v *VM) popTwo(pos lex.SrcPos) (b int64, a int64, err error) {
	b, err = v.pop(pos)
	if err != nil {
		return 0, 0, err
	}
	a, err = v.pop(pos)
	if err != nil {
		return 0, 0, err
	}
	return b, a, nil
}

func (v *VM) popVariadicFold(argc int, pos lex.SrcPos, fn func(a, b int64) int64) (int64, error) {
	if argc < 1 {
		return 0, newRuntimeError(pos, "expected at least 1 argument")
	}
	acc, err := v.pop(pos)
	if err != nil {
		return 0, err
	}
	for i := 1; i < argc; i++ {
		val, err := v.pop(pos)
		if err != nil {
			return 0, err
		}
		acc = fn(acc, val)
	}
	return acc, nil
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

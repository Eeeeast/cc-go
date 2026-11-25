package vm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"compiler/internal/lex"
)

func p() lex.SrcPos { return lex.NewSrcPos("test", 1, 1) }

func ins(op OpCode, arg int64) Instruction {
	return Instruction{Op: op, Arg: arg, Pos: p()}
}

func run(t *testing.T, code []Instruction, input string) (string, error) {
	t.Helper()
	var out strings.Builder
	err := New(strings.NewReader(input), &out).Run(code)
	return out.String(), err
}

func TestPushPop(t *testing.T) {
	_, err := run(t, []Instruction{ins(OpPush, 42), ins(OpPop, 0)}, "")
	require.NoError(t, err)
}

func TestPopUnderflow(t *testing.T) {
	_, err := run(t, []Instruction{ins(OpPop, 0)}, "")
	require.ErrorContains(t, err, "stack underflow")
}

func TestArithmetic(t *testing.T) {
	cases := []struct {
		name string
		op   OpCode
		a, b int64
		want string
	}{
		{"add", OpAdd, 2, 3, "5\n"},
		{"sub", OpSub, 7, 3, "4\n"},
		{"sub neg", OpSub, 3, 7, "-4\n"},
		{"mul", OpMul, 4, 5, "20\n"},
		{"div", OpDiv, 10, 3, "3\n"},
		{"div neg", OpDiv, -7, 2, "-3\n"},
		{"mod", OpMod, 10, 3, "1\n"},
		{"mod neg", OpMod, -7, 2, "-1\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code := []Instruction{
				ins(OpPush, c.a),
				ins(OpPush, c.b),
				ins(c.op, 0),
				ins(OpCallPrintln, 0),
				ins(OpPop, 0),
			}
			out, err := run(t, code, "")
			require.NoError(t, err)
			require.Equal(t, c.want, out)
		})
	}
}

func TestArithmeticErrors(t *testing.T) {
	cases := []struct {
		name string
		op   OpCode
		a, b int64
		msg  string
	}{
		{"div by zero", OpDiv, 1, 0, "division by zero"},
		{"mod by zero", OpMod, 1, 0, "modulo by zero"},
		{"add overflow", OpAdd, 9223372036854775807, 1, "integer overflow in addition"},
		{"sub overflow", OpSub, -9223372036854775808, 1, "integer overflow in subtraction"},
		{"mul overflow", OpMul, 9223372036854775807, 2, "integer overflow in multiplication"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code := []Instruction{
				ins(OpPush, c.a),
				ins(OpPush, c.b),
				ins(c.op, 0),
			}
			_, err := run(t, code, "")
			require.ErrorContains(t, err, c.msg)
		})
	}
}

func TestNeg(t *testing.T) {
	code := []Instruction{ins(OpPush, 5), ins(OpNeg, 0), ins(OpCallPrintln, 0), ins(OpPop, 0)}
	out, err := run(t, code, "")
	require.NoError(t, err)
	require.Equal(t, "-5\n", out)
}

func TestNegOverflow(t *testing.T) {
	code := []Instruction{ins(OpPush, -9223372036854775808), ins(OpNeg, 0)}
	_, err := run(t, code, "")
	require.ErrorContains(t, err, "integer overflow in unary negation")
}

func TestPrintln(t *testing.T) {
	code := []Instruction{
		ins(OpPush, 1), ins(OpCallPrintln, 0), ins(OpPop, 0),
		ins(OpPush, 2), ins(OpCallPrintln, 0), ins(OpPop, 0),
	}
	out, err := run(t, code, "")
	require.NoError(t, err)
	require.Equal(t, "1\n2\n", out)
}

func TestPrintlnUnitInArith(t *testing.T) {
	code := []Instruction{
		ins(OpPush, 1), ins(OpPush, 2), ins(OpCallPrintln, 0), ins(OpAdd, 0),
	}
	_, err := run(t, code, "")
	require.ErrorContains(t, err, "epected int, got ()")
}

func TestReadInt(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
		err   bool
	}{
		{"plain", "42\n", "42\n", false},
		{"negative", "-7\n", "-7\n", false},
		{"spaces", "  123  \n", "123\n", false},
		{"empty line", "\n", "", true},
		{"garbage", "abc\n", "", true},
		{"eof", "", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code := []Instruction{ins(OpCallReadInt, 0), ins(OpCallPrintln, 0), ins(OpPop, 0)}
			out, err := run(t, code, c.input)
			if c.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.want, out)
		})
	}
}

func TestAbs(t *testing.T) {
	cases := []struct{ in, want int64 }{{0, 0}, {1, 1}, {-1, 1}, {5, 5}, {-5, 5}}
	for _, c := range cases {
		code := []Instruction{
			ins(OpPush, c.in), ins(OpCallAbs, 0), ins(OpCallPrintln, 0), ins(OpPop, 0),
		}
		out, err := run(t, code, "")
		require.NoError(t, err)
		require.NotEmpty(t, out)
	}
}

func TestAbsOverflow(t *testing.T) {
	code := []Instruction{ins(OpPush, -9223372036854775808), ins(OpCallAbs, 0)}
	_, err := run(t, code, "")
	require.ErrorContains(t, err, "integer overflow in abs")
}

func TestMinMax(t *testing.T) {
	cases := []struct {
		name string
		op   OpCode
		a, b int64
		want string
	}{
		{"min a<b", OpCallMin, 1, 2, "1\n"},
		{"min a>b", OpCallMin, 5, 3, "3\n"},
		{"min eq", OpCallMin, 4, 4, "4\n"},
		{"max a<b", OpCallMax, 1, 2, "2\n"},
		{"max a>b", OpCallMax, 5, 3, "5\n"},
		{"max eq", OpCallMax, 4, 4, "4\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code := []Instruction{
				ins(OpPush, c.a), ins(OpPush, c.b), ins(c.op, 0),
				ins(OpCallPrintln, 0), ins(OpPop, 0),
			}
			out, err := run(t, code, "")
			require.NoError(t, err)
			require.Equal(t, c.want, out)
		})
	}
}

func TestUnknownOpcode(t *testing.T) {
	_, err := run(t, []Instruction{{Op: OpCode(255), Pos: p()}}, "")
	require.ErrorContains(t, err, "unknown opcode")
}

package parser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"compiler/internal/lex"
	"compiler/internal/vm"
)

func parse(t *testing.T, src string) []vm.Instruction {
	t.Helper()
	p, err := New(lex.NewLexer(src, "test"))
	require.NoError(t, err)
	code, err := p.ParseProgram()
	require.NoError(t, err)
	return code
}

func parseErr(t *testing.T, src string) {
	t.Helper()
	p, err := New(lex.NewLexer(src, "test"))
	if err != nil {
		return
	}
	_, err = p.ParseProgram()
	require.Error(t, err)
}

func TestParseProgram(t *testing.T) {
	parse(t, "func main() {}")
	parse(t, "func main() { 1; }")
	parse(t, "func main() { 1; 2; }")

	parseErr(t, "")
	parseErr(t, "main() {}")
	parseErr(t, "func f() {}")
	parseErr(t, "func main) {}")
	parseErr(t, "func main( {}")
	parseErr(t, "func main() {")
	parseErr(t, "func main() {} extra")
}

func TestParseStatement(t *testing.T) {
	parse(t, "func main() { 42; }")
	parseErr(t, "func main() { 42 }")
}

func TestParseAdditive(t *testing.T) {
	parse(t, "func main() { 1 + 2; }")
	parse(t, "func main() { 1 - 2; }")
	parse(t, "func main() { 1 + 2 + 3; }")
}

func TestParseMultiplicative(t *testing.T) {
	parse(t, "func main() { 2 * 3; }")
	parse(t, "func main() { 6 / 2; }")
	parse(t, "func main() { 7 % 2; }")
	parse(t, "func main() { 1 + 2 * 3; }")
}

func TestParseUnary(t *testing.T) {
	parse(t, "func main() { -5; }")
	parse(t, "func main() { --5; }")
	parse(t, "func main() { -(1 + 2); }")
}

func TestParsePrimary(t *testing.T) {
	parse(t, "func main() { 0; }")
	parse(t, "func main() { (5); }")
	parse(t, "func main() { (1 + 2) * 3; }")

	parseErr(t, "func main() { 9223372036854775808; }")
	parseErr(t, "func main() { (1; }")
	parseErr(t, "func main() { (); }")
}

func TestParseCall(t *testing.T) {
	parse(t, "func main() { fmt.Println(1); }")
	parse(t, "func main() { fmt.ReadInt(); }")
	parse(t, "func main() { abs(1); }")
	parse(t, "func main() { min(1, 2); }")
	parse(t, "func main() { max(1, 2); }")
	parse(t, "func main() { abs(min(1, 2)); }")

	parseErr(t, "func main() { foo(1); }")
	parseErr(t, "func main() { fmt.Foo(1); }")
	parseErr(t, "func main() { abs(); }")
	parseErr(t, "func main() { abs(1, 2); }")
	parseErr(t, "func main() { min(1); }")
	parseErr(t, "func main() { fmt.Println(); }")
	parseErr(t, "func main() { fmt.ReadInt(1); }")
	parseErr(t, "func main() { x; }")
}

package driver

import (
	"io"

	"compiler/internal/codegen"
	"compiler/internal/lex"
	"compiler/internal/parser"
	"compiler/internal/vm"
)

func Interpret(stdin io.Reader, stdout io.Writer, path string) error {
	lx, err := lex.FromFile(path)
	if err != nil {
		return err
	}
	p, err := parser.New(lx)
	if err != nil {
		return err
	}
	prog, err := p.ParseProgram()
	if err != nil {
		return err
	}
	code, err := codegen.Generate(prog)
	if err != nil {
		return err
	}
	return vm.New(stdin, stdout).Run(code)
}

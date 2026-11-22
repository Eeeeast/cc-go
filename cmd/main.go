package main

import (
	"fmt"
	"os"

	"compiler/internal/lex"
	"compiler/internal/parser"
	"compiler/internal/vm"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: minigo <source-file>")
		os.Exit(1)
	}

	lexer, err := lex.FromFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p, err := parser.New(lexer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code, err := p.ParseProgram()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	machine := vm.New(os.Stdin, os.Stdout)
	if err := machine.Run(code); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

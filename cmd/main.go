package main

import (
	"fmt"
	"os"

	"compiler/internal/driver"
)

const NumberOfArguments = 2

func main() {
	if len(os.Args) != NumberOfArguments {
		fmt.Fprintln(os.Stderr, "usage: minigo <source-file>")
		os.Exit(1)
	}

	err := driver.Interpret(os.Stdin, os.Stdout, os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

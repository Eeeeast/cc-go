package main

import (
	"fmt"
	"os"

	"compiler/internal/driver"
)

const numberOfArguments = 2

func main() {
	if len(os.Args) != numberOfArguments {
		fmt.Fprintln(os.Stderr, "usage: minigo <source-file>")
		os.Exit(1)
	}

	err := driver.Interpret(os.Stdin, os.Stdout, os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"compiler/internal/driver"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: minigo <source-file>")
		os.Exit(1)
	}

	if err := driver.Interpret(os.Stdin, os.Stdout, os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

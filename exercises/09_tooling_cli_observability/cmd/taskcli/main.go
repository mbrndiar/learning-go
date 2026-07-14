// Command taskcli is the thin entry point around package cliapp: it parses
// os.Args, delegates all real work to cliapp.Run, and turns a returned error
// into a non-zero exit code. Keeping main this small is what makes the rest
// of the logic unit-testable without spawning a subprocess.
package main

import (
	"fmt"
	"os"

	"github.com/mbrndiar/learning-go/exercises/09_tooling_cli_observability"
)

func main() {
	if err := cliapp.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "taskcli:", err)
		os.Exit(1)
	}
}

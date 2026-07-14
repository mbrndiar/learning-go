// Command taskcli is the thin entry point around package solution: it parses
// os.Args, delegates all real work to solution.Run, and turns a returned
// error into a non-zero exit code.
package main

import (
	"fmt"
	"os"

	"github.com/mbrndiar/learning-go/exercises/09_tooling_cli_observability/solution"
)

func main() {
	if err := solution.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "taskcli:", err)
		os.Exit(1)
	}
}

// Command greet is a tiny CLI built with the standard library's flag
// package. It demonstrates string, int, and bool flags, a custom flag.Usage
// message, and reading positional arguments after the flags.
//
// Try it:
//
//	go run ./lessons/09_tooling_cli_observability/01_cli_flags -name Ada -times 2 -shout
//	go run ./lessons/09_tooling_cli_observability/01_cli_flags -h
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run contains the program logic and takes its dependencies as parameters
// (arguments, stdout, stderr) instead of reading globals directly. That
// makes it straightforward to call from a test with fake arguments and an
// in-memory buffer, without needing to fork a real process.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("greet", flag.ContinueOnError)
	fs.SetOutput(stderr)

	name := fs.String("name", "World", "name of the person to greet")
	times := fs.Int("times", 1, "number of times to repeat the greeting")
	shout := fs.Bool("shout", false, "print the greeting in upper case")

	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: greet [flags] [extra words...]")
		fmt.Fprintln(stderr, "Flags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		// flag.ErrHelp is returned for -h/-help; fs.Parse has already
		// printed the usage message via fs.Usage in that case.
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if *times < 1 {
		fmt.Fprintln(stderr, "error: -times must be at least 1")
		return 2
	}

	// fs.Args returns whatever is left after flag parsing: positional
	// arguments that are not themselves flags.
	extra := strings.Join(fs.Args(), " ")

	message := fmt.Sprintf("Hello, %s!", *name)
	if extra != "" {
		message = fmt.Sprintf("%s (%s)", message, extra)
	}
	if *shout {
		message = strings.ToUpper(message)
	}

	for i := 0; i < *times; i++ {
		fmt.Fprintln(stdout, message)
	}
	return 0
}

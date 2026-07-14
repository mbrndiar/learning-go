// Package cliapp implements the core, testable logic behind a tiny greeting
// CLI. The pattern is deliberate: main() should stay a few lines that parse
// os.Args, call into this package, and translate an error into an exit code.
// Everything that can be unit tested — flag parsing, logging configuration,
// business logic, diagnostics — lives here instead, taking explicit
// io.Writer/argument parameters instead of touching os.Args, os.Stdout, or
// flag.CommandLine directly.
package cliapp

import (
	"io"
	"log/slog"
)

// Config holds the parsed command-line configuration for the CLI.
type Config struct {
	Name      string // greeted name, e.g. "World"
	Count     int    // number of times to print the greeting, must be >= 1
	Verbose   bool   // enable debug-level logging
	LogFormat string // "text" or "json"
}

// ParseArgs parses args (typically os.Args[1:]) into a Config using a
// dedicated flag.FlagSet rather than the global flag.CommandLine, so it can
// be called repeatedly and concurrently in tests without shared state.
//
// Supported flags:
//
//	-name string        name to greet (default "World")
//	-count int          number of greetings to print (default 1, must be >= 1)
//	-verbose            enable debug logging (default false)
//	-log-format string  "text" or "json" (default "text")
//
// If args requests help (-h/-help), ParseArgs returns flag.ErrHelp as the
// error so callers can distinguish "print usage and exit 0" from a genuine
// parse failure. Any other flag.FlagSet parse error, an out-of-range -count,
// or an unrecognized -log-format must be returned as a descriptive error.
//
// TODO(task 1): implement ParseArgs.
func ParseArgs(args []string) (Config, error) {
	panic("not implemented")
}

// NewLogger returns an *slog.Logger that writes to w using a JSON handler
// when cfg.LogFormat == "json" and a text handler otherwise, defaulting to
// slog.LevelDebug when cfg.Verbose is true and slog.LevelInfo otherwise.
//
// TODO(task 2): implement NewLogger.
func NewLogger(cfg Config, w io.Writer) *slog.Logger {
	panic("not implemented")
}

// Greeting returns cfg.Count lines, each greeting cfg.Name, in the form
// "Hello, <name>!". Callers are expected to have already validated
// cfg.Count >= 1 (ParseArgs enforces this); Greeting itself simply returns
// an empty slice for cfg.Count <= 0.
//
// TODO(task 3): implement Greeting.
func Greeting(cfg Config) []string {
	panic("not implemented")
}

// Diagnostics returns a small set of runtime facts useful in a bug report:
// the keys "go_version", "os", "arch", "num_cpu", and "num_goroutine", built
// from the runtime package (runtime.Version, runtime.GOOS, runtime.GOARCH,
// runtime.NumCPU, runtime.NumGoroutine). Numeric values must be formatted as
// base-10 strings (e.g. with strconv.Itoa).
//
// TODO(task 4): implement Diagnostics.
func Diagnostics() map[string]string {
	panic("not implemented")
}

// Run parses args, wires up a logger writing to stderr, prints the greeting
// to stdout, and logs a debug line with the parsed config plus an info line
// with diagnostics. It returns an error instead of calling os.Exit so it
// remains fully testable; -h/-help prints usage to stdout and returns nil
// (not an error).
//
// TODO(task 5): implement Run using ParseArgs, NewLogger, Greeting, and
// Diagnostics. Use errors.Is(err, flag.ErrHelp) to detect the help case.
func Run(args []string, stdout, stderr io.Writer) error {
	panic("not implemented")
}

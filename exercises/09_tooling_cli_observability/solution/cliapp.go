// Package solution is the reference implementation for
// exercises/09_tooling_cli_observability.
package solution

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
)

// Config holds the parsed command-line configuration for the CLI.
type Config struct {
	Name      string
	Count     int
	Verbose   bool
	LogFormat string
}

// ParseArgs parses args using a dedicated flag.FlagSet rather than the
// global flag.CommandLine, so it can be called repeatedly and concurrently
// in tests without shared state.
func ParseArgs(args []string) (Config, error) {
	fs := flag.NewFlagSet("taskcli", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := Config{}
	fs.StringVar(&cfg.Name, "name", "World", "name to greet")
	fs.IntVar(&cfg.Count, "count", 1, "number of greetings to print")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "enable debug logging")
	fs.StringVar(&cfg.LogFormat, "log-format", "text", `log format: "text" or "json"`)

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, flag.ErrHelp
		}
		return Config{}, fmt.Errorf("parsing flags: %w", err)
	}

	if cfg.Count < 1 {
		return Config{}, fmt.Errorf("-count must be >= 1, got %d", cfg.Count)
	}
	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		return Config{}, fmt.Errorf("-log-format must be %q or %q, got %q", "text", "json", cfg.LogFormat)
	}

	return cfg, nil
}

// NewLogger returns an *slog.Logger writing to w, using JSON or text output
// based on cfg.LogFormat and slog.LevelDebug/slog.LevelInfo based on
// cfg.Verbose.
func NewLogger(cfg Config, w io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if cfg.Verbose {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}
	return slog.New(handler)
}

// Greeting returns cfg.Count lines greeting cfg.Name.
func Greeting(cfg Config) []string {
	lines := make([]string, 0, cfg.Count)
	for i := 0; i < cfg.Count; i++ {
		lines = append(lines, fmt.Sprintf("Hello, %s!", cfg.Name))
	}
	return lines
}

// Diagnostics returns a small set of runtime facts useful in a bug report.
func Diagnostics() map[string]string {
	return map[string]string{
		"go_version":    runtime.Version(),
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"num_cpu":       strconv.Itoa(runtime.NumCPU()),
		"num_goroutine": strconv.Itoa(runtime.NumGoroutine()),
	}
}

// Run parses args, wires a logger writing to stderr, prints the greeting to
// stdout, and logs the parsed config plus diagnostics. -h/-help prints usage
// to stdout and returns nil.
func Run(args []string, stdout, stderr io.Writer) error {
	cfg, err := ParseArgs(args)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(stdout, "usage: taskcli [-name NAME] [-count N] [-verbose] [-log-format text|json]")
		return nil
	}
	if err != nil {
		return err
	}

	logger := NewLogger(cfg, stderr)
	logger.Debug("parsed config", "name", cfg.Name, "count", cfg.Count, "log_format", cfg.LogFormat)

	for _, line := range Greeting(cfg) {
		fmt.Fprintln(stdout, line)
	}

	diag := Diagnostics()
	logger.Info("diagnostics",
		"go_version", diag["go_version"],
		"os", diag["os"],
		"arch", diag["arch"],
		"num_cpu", diag["num_cpu"],
		"num_goroutine", diag["num_goroutine"],
	)

	return nil
}

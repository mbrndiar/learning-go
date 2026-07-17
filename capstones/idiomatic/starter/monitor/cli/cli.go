// Package cli owns the testable monitor command boundary.
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/probe"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/scheduler"
)

const (
	// ExitOK indicates normal command completion.
	ExitOK = 0
	// ExitUsage indicates invalid command-line usage.
	ExitUsage = 2
	// ExitConfig indicates invalid or unsupported configuration.
	ExitConfig = 3
	// ExitConfigIO indicates a configuration file I/O error.
	ExitConfigIO = 4
	// ExitInternal indicates monitor or server startup/internal failure.
	ExitInternal = 5
	// ExitCancelled indicates a cancelled one-shot check.
	ExitCancelled = 130
)

// Dependencies supplies deterministic application and server seams.
type Dependencies struct {
	Client          *http.Client
	Prober          probe.Prober
	Trigger         scheduler.Trigger
	Listen          func(network, address string) (net.Listener, error)
	Now             func() time.Time
	Logger          *slog.Logger
	ShutdownTimeout time.Duration
}

// Run is the stable check/serve process boundary.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return RunWithDependencies(ctx, args, stdout, stderr, Dependencies{})
}

// RunWithDependencies runs a command with deterministic test seams.
func RunWithDependencies(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	dependencies Dependencies,
) int {
	writePlaceholder(stderr)
	return 1
}

func writePlaceholder(stderr io.Writer) {
	fmt.Fprintln(stderr, "monitor: not implemented")
}

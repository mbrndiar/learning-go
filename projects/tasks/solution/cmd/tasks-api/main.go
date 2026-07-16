// Command tasks-api runs a selected Task HTTP server and persistence backend.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/server"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	config := server.DefaultConfig()
	flags := flag.NewFlagSet("tasks-api", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&config.Server, "server", config.Server, "HTTP server (nethttp)")
	flags.StringVar(&config.Backend, "backend", config.Backend, "storage backend (sqlite or markdown)")
	flags.StringVar(&config.Data, "data", config.Data, "storage path")
	flags.StringVar(&config.Host, "host", config.Host, "listen host")
	flags.IntVar(&config.Port, "port", config.Port, "listen port")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if len(flags.Args()) != 0 {
		return 2
	}
	if _, err := config.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Run(ctx, config, slog.Default()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

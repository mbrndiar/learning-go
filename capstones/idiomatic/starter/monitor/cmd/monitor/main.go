// Command monitor runs the idiomatic service health monitor.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

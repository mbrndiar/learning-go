// Command monitor runs the idiomatic service health monitor.
package main

import (
	"context"
	"os"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/cli"
)

func main() {
	os.Exit(cli.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

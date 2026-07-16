// Command tasks-api runs a selected Task HTTP server and persistence backend.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/server"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	config := server.DefaultConfig()
	flags := flag.NewFlagSet("tasks-api", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&config.Server, "server", config.Server, "HTTP server")
	flags.StringVar(&config.Backend, "backend", config.Backend, "storage backend")
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
	if config.Server != "nethttp" && config.Server != "chi" && config.Server != "gin" {
		fmt.Fprintf(os.Stderr, "task server: invalid configuration: server %q is not implemented\n", config.Server)
		return 2
	}
	fmt.Fprintln(os.Stderr, task.ErrNotImplemented)
	return 1
}

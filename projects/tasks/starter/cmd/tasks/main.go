// Command tasks runs the Task command-line client with a selected transport.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/cli"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	clientName := "nethttp"
	remaining := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		if args[index] == "--client" {
			if index+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "usage: tasks --client nethttp|resty [options] command")
				return cli.ExitUsage
			}
			clientName = args[index+1]
			index++
			continue
		}
		if strings.HasPrefix(args[index], "--client=") {
			clientName = strings.TrimPrefix(args[index], "--client=")
			continue
		}
		remaining = append(remaining, args[index])
	}
	if clientName != "nethttp" && clientName != "resty" {
		fmt.Fprintf(os.Stderr, "configuration: client %q is not implemented\n", clientName)
		return cli.ExitUsage
	}
	factory := func(client.Config) (client.Transport, error) {
		return nil, task.ErrNotImplemented
	}
	return cli.Run(remaining, factory, os.Stdout, os.Stderr)
}

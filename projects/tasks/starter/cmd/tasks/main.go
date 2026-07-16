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
	remaining := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		if args[index] == "--client" {
			if index+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "usage: tasks --client nethttp [options] command")
				return cli.ExitUsage
			}
			index++
			continue
		}
		if strings.HasPrefix(args[index], "--client=") {
			continue
		}
		remaining = append(remaining, args[index])
	}
	factory := func(client.Config) (client.Transport, error) {
		return nil, task.ErrNotImplemented
	}
	return cli.Run(remaining, factory, os.Stdout, os.Stderr)
}

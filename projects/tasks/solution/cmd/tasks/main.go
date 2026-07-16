// Command tasks runs the Task command-line client with a selected transport.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/cli"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	clientnethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/client/nethttp"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	clientName, remaining, err := selectClient(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitUsage
	}
	if clientName != "nethttp" {
		fmt.Fprintf(os.Stderr, "configuration: client %q is not implemented\n", clientName)
		return cli.ExitUsage
	}
	factory := func(config client.Config) (client.Transport, error) {
		return clientnethttp.New(config)
	}
	return cli.Run(remaining, factory, os.Stdout, os.Stderr)
}

func selectClient(args []string) (string, []string, error) {
	selected := "nethttp"
	remaining := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--client":
			if index+1 >= len(args) {
				return "", nil, fmt.Errorf("usage: tasks --client nethttp [options] command")
			}
			selected = args[index+1]
			index++
		default:
			if strings.HasPrefix(args[index], "--client=") {
				selected = strings.TrimPrefix(args[index], "--client=")
			} else {
				remaining = append(remaining, args[index])
			}
		}
	}
	return selected, remaining, nil
}

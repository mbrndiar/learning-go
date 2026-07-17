package cli

import (
	"io"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// Run parses args, builds a Transport with factory, executes the requested
// command, and writes output to stdout or stderr, returning the process exit
// code.
func Run(args []string, factory Factory, stdout, stderr io.Writer) int {
	request, err := ParseRequest(args)
	if err != nil {
		_, _ = io.WriteString(stderr, "usage: tasks [--base-url URL] [--timeout SECONDS] <add|list|show|update|complete|remove> ...\n")
		return ExitUsage
	}
	if request.Help {
		_, _ = io.WriteString(stdout, "usage: tasks [--base-url URL] [--timeout SECONDS] <add|list|show|update|complete|remove> ...\n")
		return ExitSuccess
	}
	_, _ = io.WriteString(stderr, task.ErrNotImplemented.Error()+"\n")
	return 1
}

// Main is the cli package's entry point for command binaries.
func Main(args []string, factory Factory, stdout, stderr io.Writer) int {
	return Run(args, factory, stdout, stderr)
}

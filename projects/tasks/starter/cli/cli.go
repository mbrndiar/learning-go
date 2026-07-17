// Package cli owns shared Task command parsing, output, and exit-code policy.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// Exit codes form the stable process contract for scripts invoking tasks.
const (
	// ExitSuccess reports a completed command.
	ExitSuccess = 0
	// ExitUsage reports invalid command syntax or configuration.
	ExitUsage = 2
	// ExitAPI reports a valid error returned by the Task API.
	ExitAPI = 3
	// ExitUnexpectedResponse reports a response outside the wire contract.
	ExitUnexpectedResponse = 4
	// ExitConnection reports transport setup, connection, timeout, or output failure.
	ExitConnection = 5
)

// Factory builds the Transport used to execute a parsed Request.
type Factory func(client.Config) (client.Transport, error)

// Settings holds the parsed connection options shared by every command.
type Settings struct {
	BaseURL string
	Timeout time.Duration
}

// Request is a fully parsed invocation: connection Settings plus exactly one
// command value, unless Help is set.
type Request struct {
	Settings Settings
	Command  any
	Help     bool
}

// AddCommand requests creation of a task.
type AddCommand struct{ Title string }

// ListCommand requests tasks matching an optional completion filter.
type ListCommand struct{ Completed *bool }

// ShowCommand requests one task by ID.
type ShowCommand struct{ ID int64 }

// UpdateCommand requests changes to the fields whose pointers are non-nil.
type UpdateCommand struct {
	ID        int64
	Title     *string
	Completed *bool
}

// CompleteCommand marks one task complete.
type CompleteCommand struct{ ID int64 }

// RemoveCommand deletes one task.
type RemoveCommand struct{ ID int64 }

// ParseRequest parses args into a Request, validating shared flags and
// dispatching to the selected command's own flags.
func ParseRequest(args []string) (Request, error) {
	settings := Settings{BaseURL: "http://127.0.0.1:8000", Timeout: client.DefaultTimeout}
	flags := flag.NewFlagSet("tasks", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&settings.BaseURL, "base-url", settings.BaseURL, "")
	timeout := "5"
	flags.StringVar(&timeout, "timeout", timeout, "")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Request{Help: true}, nil
		}
		return Request{}, err
	}
	seconds, err := strconv.ParseFloat(timeout, 64)
	if err != nil || seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return Request{}, errors.New("timeout must be positive and finite")
	}
	settings.Timeout = time.Duration(seconds * float64(time.Second))
	parsed, err := url.Parse(strings.TrimSpace(settings.BaseURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return Request{}, errors.New("base URL must be an absolute HTTP URL")
	}
	parsed.Scheme, parsed.Host = strings.ToLower(parsed.Scheme), strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	settings.BaseURL = parsed.String()
	remaining := flags.Args()
	if len(remaining) == 0 {
		return Request{}, errors.New("a command is required")
	}
	commandArgs := remaining[1:]
	var command any
	switch remaining[0] {
	case "add":
		if len(commandArgs) != 1 {
			return Request{}, errors.New("add requires exactly one title")
		}
		command = AddCommand{Title: commandArgs[0]}
	case "list":
		commandFlags := flag.NewFlagSet("list", flag.ContinueOnError)
		commandFlags.SetOutput(io.Discard)
		var completed optionalBool
		commandFlags.Var(&completed, "completed", "")
		if err := commandFlags.Parse(commandArgs); err != nil || len(commandFlags.Args()) != 0 {
			return Request{}, errors.New("list accepts only --completed true|false")
		}
		command = ListCommand{Completed: completed.pointer()}
	case "show", "complete", "remove":
		id, err := singleID(commandArgs)
		if err != nil {
			return Request{}, err
		}
		switch remaining[0] {
		case "show":
			command = ShowCommand{ID: id}
		case "complete":
			command = CompleteCommand{ID: id}
		case "remove":
			command = RemoveCommand{ID: id}
		}
	case "update":
		if len(commandArgs) == 0 {
			return Request{}, errors.New("update requires a task ID")
		}
		id, err := positiveID(commandArgs[0])
		if err != nil {
			return Request{}, err
		}
		commandFlags := flag.NewFlagSet("update", flag.ContinueOnError)
		commandFlags.SetOutput(io.Discard)
		var title optionalString
		var completed optionalBool
		commandFlags.Var(&title, "title", "")
		commandFlags.Var(&completed, "completed", "")
		if err := commandFlags.Parse(commandArgs[1:]); err != nil || len(commandFlags.Args()) != 0 ||
			(!title.set && !completed.set) {
			return Request{}, errors.New("update requires --title or --completed")
		}
		command = UpdateCommand{ID: id, Title: title.pointer(), Completed: completed.pointer()}
	default:
		return Request{}, fmt.Errorf("unknown command: %s", remaining[0])
	}
	return Request{Settings: settings, Command: command}, nil
}

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

func singleID(args []string) (int64, error) {
	if len(args) != 1 {
		return 0, errors.New("command requires exactly one task ID")
	}
	return positiveID(args[0])
}

func positiveID(raw string) (int64, error) {
	for _, value := range []byte(raw) {
		if value < '0' || value > '9' {
			return 0, errors.New("task ID must be a positive integer")
		}
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("task ID must be a positive integer")
	}
	return id, nil
}

type optionalBool struct {
	set, value bool
}

func (value *optionalBool) String() string {
	if !value.set {
		return ""
	}
	return strconv.FormatBool(value.value)
}

func (value *optionalBool) Set(raw string) error {
	if raw != "true" && raw != "false" {
		return errors.New("must be true or false")
	}
	value.set, value.value = true, raw == "true"
	return nil
}

func (value *optionalBool) pointer() *bool {
	if !value.set {
		return nil
	}
	result := value.value
	return &result
}

type optionalString struct {
	set   bool
	value string
}

func (value *optionalString) String() string { return value.value }
func (value *optionalString) Set(raw string) error {
	value.set, value.value = true, raw
	return nil
}
func (value *optionalString) pointer() *string {
	if !value.set {
		return nil
	}
	result := value.value
	return &result
}

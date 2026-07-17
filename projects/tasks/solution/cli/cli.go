// Package cli owns shared Task command parsing, output, and exit-code policy.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
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

// Factory constructs the selected client transport for one command invocation.
type Factory func(client.Config) (client.Transport, error)

// Settings contains transport-independent CLI connection settings.
type Settings struct {
	BaseURL string
	Timeout time.Duration
}

// Request is the parsed CLI configuration and concrete command.
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

// ParseRequest validates global options and parses one Task subcommand.
func ParseRequest(args []string) (Request, error) {
	settings := Settings{BaseURL: "http://127.0.0.1:8000", Timeout: client.DefaultTimeout}
	global := flag.NewFlagSet("tasks", flag.ContinueOnError)
	global.SetOutput(io.Discard)
	global.StringVar(&settings.BaseURL, "base-url", settings.BaseURL, "")
	timeout := "5"
	global.StringVar(&timeout, "timeout", timeout, "")
	if err := global.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Request{Help: true}, nil
		}
		return Request{}, err
	}
	if err := parseTimeout(timeout, &settings.Timeout); err != nil {
		return Request{}, err
	}
	normalized, err := client.NormalizeBaseURL(settings.BaseURL)
	if err != nil {
		return Request{}, err
	}
	settings.BaseURL = normalized
	remaining := global.Args()
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
		flags := flag.NewFlagSet("list", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		var completed optionalBool
		flags.Var(&completed, "completed", "")
		if err := flags.Parse(commandArgs); err != nil || len(flags.Args()) != 0 {
			return Request{}, errors.New("list accepts only --completed true|false")
		}
		command = ListCommand{Completed: completed.pointer()}
	case "show":
		id, parseErr := parseSingleID(commandArgs)
		if parseErr != nil {
			return Request{}, parseErr
		}
		command = ShowCommand{ID: id}
	case "update":
		if len(commandArgs) == 0 {
			return Request{}, errors.New("update requires a task ID")
		}
		id, parseErr := parseID(commandArgs[0])
		if parseErr != nil {
			return Request{}, parseErr
		}
		flags := flag.NewFlagSet("update", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		var title optionalString
		var completed optionalBool
		flags.Var(&title, "title", "")
		flags.Var(&completed, "completed", "")
		if err := flags.Parse(commandArgs[1:]); err != nil || len(flags.Args()) != 0 {
			return Request{}, errors.New("update accepts --title and --completed")
		}
		if !title.set && !completed.set {
			return Request{}, errors.New("update requires --title or --completed")
		}
		command = UpdateCommand{ID: id, Title: title.pointer(), Completed: completed.pointer()}
	case "complete":
		id, parseErr := parseSingleID(commandArgs)
		if parseErr != nil {
			return Request{}, parseErr
		}
		command = CompleteCommand{ID: id}
	case "remove":
		id, parseErr := parseSingleID(commandArgs)
		if parseErr != nil {
			return Request{}, parseErr
		}
		command = RemoveCommand{ID: id}
	default:
		return Request{}, fmt.Errorf("unknown command: %s", remaining[0])
	}
	return Request{Settings: settings, Command: command}, nil
}

// Run executes one parsed command and returns the stable process exit code.
func Run(args []string, factory Factory, stdout, stderr io.Writer) int {
	request, err := ParseRequest(args)
	if err != nil {
		writeUsage(stderr)
		return ExitUsage
	}
	if request.Help {
		writeUsage(stdout)
		return ExitSuccess
	}
	transport, err := factory(client.Config{
		BaseURL: request.Settings.BaseURL,
		Timeout: request.Settings.Timeout,
	})
	if err != nil {
		fmt.Fprintln(stderr, "transport: request failed")
		return ExitConnection
	}

	ctx, cancel := context.WithTimeout(context.Background(), request.Settings.Timeout)
	defer cancel()
	var output any
	switch command := request.Command.(type) {
	case AddCommand:
		output, err = transport.Create(ctx, task.CreateInput{Title: command.Title})
	case ListCommand:
		output, err = transport.List(ctx, task.ListFilter{Completed: command.Completed})
	case ShowCommand:
		output, err = transport.Get(ctx, command.ID)
	case UpdateCommand:
		output, err = transport.Update(ctx, command.ID, task.UpdateInput{
			Title: command.Title, Completed: command.Completed,
		})
	case CompleteCommand:
		completed := true
		output, err = transport.Update(ctx, command.ID, task.UpdateInput{Completed: &completed})
	case RemoveCommand:
		err = transport.Delete(ctx, command.ID)
		output = map[string]int64{"deleted": command.ID}
	default:
		err = errors.New("unknown parsed command")
	}
	// The CLI owns the transport produced by factory, so it always closes it
	// on both successful and failed requests. A command error takes
	// precedence over a cleanup failure: the command's own error, response
	// validation, or connection failure is what the caller needs to see. Only
	// when the command itself succeeded does a Close failure surface as
	// ExitConnection, and it does so before anything is written to stdout.
	var closeErr error
	if closer, ok := transport.(interface{ Close() error }); ok {
		closeErr = closer.Close()
	}
	if err != nil {
		return renderError(err, stderr)
	}
	if closeErr != nil {
		fmt.Fprintln(stderr, "transport: request failed")
		return ExitConnection
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(jsonOutput(output)); err != nil {
		fmt.Fprintln(stderr, "transport: request failed")
		return ExitConnection
	}
	return ExitSuccess
}

type taskOutput struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func jsonOutput(output any) any {
	switch value := output.(type) {
	case task.Task:
		return taskOutput{ID: value.ID, Title: value.Title, Completed: value.Completed}
	case []task.Task:
		result := make([]taskOutput, len(value))
		for index, item := range value {
			result[index] = taskOutput{ID: item.ID, Title: item.Title, Completed: item.Completed}
		}
		return result
	default:
		return output
	}
}

// Main is the reusable command entry point for thin executable wrappers.
func Main(args []string, factory Factory, stdout, stderr io.Writer) int {
	return Run(args, factory, stdout, stderr)
}

func renderError(err error, stderr io.Writer) int {
	var apiError *client.APIError
	if errors.As(err, &apiError) {
		fmt.Fprintf(stderr, "api: %d %s: %s\n", apiError.Status, apiError.Code, apiError.Message)
		return ExitAPI
	}
	if errors.Is(err, client.ErrUnexpectedResponse) {
		var responseError *client.ResponseError
		if errors.As(err, &responseError) {
			fmt.Fprintf(stderr, "malformed-response: %s\n", responseError.Message)
		} else {
			fmt.Fprintln(stderr, "malformed-response: invalid server response")
		}
		return ExitUnexpectedResponse
	}
	if errors.Is(err, client.ErrConnection) {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintln(stderr, "connection: timeout: deadline exceeded")
		} else {
			var connectionError *client.ConnectionError
			if errors.As(err, &connectionError) && connectionError.Err != nil {
				fmt.Fprintf(stderr, "connection: %s\n", connectionError.Err)
			} else {
				fmt.Fprintln(stderr, "connection: request failed")
			}
		}
		return ExitConnection
	}
	fmt.Fprintln(stderr, "transport: request failed")
	return ExitConnection
}

func writeUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: tasks [--base-url URL] [--timeout SECONDS] <add|list|show|update|complete|remove> ...")
}

func parseTimeout(raw string, target *time.Duration) error {
	seconds, err := strconv.ParseFloat(raw, 64)
	if err == nil {
		if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
			return errors.New("timeout must be positive and finite")
		}
		duration := time.Duration(seconds * float64(time.Second))
		if duration <= 0 {
			return errors.New("timeout must be positive and finite")
		}
		*target = duration
		return nil
	}
	duration, durationErr := time.ParseDuration(raw)
	if durationErr != nil || duration <= 0 {
		return errors.New("timeout must be positive and finite")
	}
	*target = duration
	return nil
}

func parseSingleID(args []string) (int64, error) {
	if len(args) != 1 {
		return 0, errors.New("command requires exactly one task ID")
	}
	return parseID(args[0])
}

func parseID(raw string) (int64, error) {
	if raw == "" {
		return 0, errors.New("task ID must be a positive integer")
	}
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
	set   bool
	value bool
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
	value.set = true
	value.value = raw == "true"
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

func (value *optionalString) String() string {
	return value.value
}

func (value *optionalString) Set(raw string) error {
	value.set = true
	value.value = raw
	return nil
}

func (value *optionalString) pointer() *string {
	if !value.set {
		return nil
	}
	result := strings.Clone(value.value)
	return &result
}

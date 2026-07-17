package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

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

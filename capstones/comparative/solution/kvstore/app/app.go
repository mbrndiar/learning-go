// Package app owns the testable comparative command boundary.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/mbrndiar/learning-go/capstones/comparative/solution/kvstore/domain"
	"github.com/mbrndiar/learning-go/capstones/comparative/solution/kvstore/storage"
)

// Run is the stable process boundary used by the thin command.
func Run(ctx context.Context, args []string, stdout io.Writer, _ io.Writer) int {
	result, err := run(ctx, args, storage.NewSQLiteOpener())
	if err != nil {
		return writeError(stdout, err)
	}
	if err := writeEnvelope(stdout, map[string]any{"ok": true, "result": result}); err != nil {
		return 5
	}
	return 0
}

type parsedCommand struct {
	database    string
	name        string
	key         string
	value       domain.Value
	expectation domain.Expectation
	expectText  string
	hasExpect   bool
}

func run(ctx context.Context, args []string, opener storage.Opener) (any, error) {
	command, err := parseCommand(args)
	if err != nil {
		return nil, err
	}
	store, err := opener.Open(ctx, command.database)
	if err != nil {
		return nil, err
	}

	var result any
	switch command.name {
	case "set":
		result, err = store.Set(ctx, command.key, command.value, command.expectation)
	case "get":
		result, err = store.Get(ctx, command.key)
	case "delete":
		result, err = store.Delete(ctx, command.key, command.expectation)
	case "list":
		result, err = store.List(ctx)
	default:
		panic("validated command has unknown name")
	}
	closeErr := store.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		operation := "read"
		if command.name == "set" || command.name == "delete" {
			operation = "commit"
		}
		return nil, &domain.Error{
			Category: "storage_error",
			Details: map[string]any{
				"operation": operation,
				"reason":    "storage_failure",
			},
			Cause: closeErr,
		}
	}
	return result, nil
}

func parseCommand(args []string) (parsedCommand, error) {
	if len(args) < 3 || args[0] != "--db" {
		return parsedCommand{}, usageError()
	}
	command := parsedCommand{database: args[1], name: args[2]}
	switch {
	case command.name == "list" && len(args) == 3:
	case command.name == "get" && len(args) == 4:
		command.key = args[3]
	case command.name == "delete" && len(args) == 4:
		command.key = args[3]
		command.expectation.Kind = domain.ExpectAny
	case command.name == "delete" &&
		len(args) == 6 &&
		args[4] == "--expect":
		command.key = args[3]
		command.expectText = args[5]
		command.hasExpect = true
	case command.name == "set" &&
		len(args) == 6 &&
		args[4] == "--value-json":
		command.key = args[3]
		command.expectation.Kind = domain.ExpectAny
	case command.name == "set" &&
		len(args) == 8 &&
		args[4] == "--value-json" &&
		args[6] == "--expect":
		command.key = args[3]
		command.expectText = args[7]
		command.hasExpect = true
	default:
		return parsedCommand{}, usageError()
	}

	if command.database == "" {
		return parsedCommand{}, &domain.Error{
			Category: "invalid_argument",
			Details:  map[string]any{"field": "db", "reason": "empty"},
		}
	}
	if command.database == ":memory:" || strings.HasPrefix(command.database, "file:") {
		return parsedCommand{}, &domain.Error{
			Category: "invalid_argument",
			Details:  map[string]any{"field": "db", "reason": "unsupported_form"},
		}
	}
	if command.name == "list" {
		return command, nil
	}
	key, err := domain.ParseKey(command.key)
	if err != nil {
		return parsedCommand{}, err
	}
	command.key = key

	if command.hasExpect {
		expectation, err := domain.ParseExpectation(command.expectText, command.name == "set")
		if err != nil {
			return parsedCommand{}, err
		}
		command.expectation = expectation
	}
	if command.name == "set" {
		value, err := domain.ParseValue(json.RawMessage(args[5]))
		if err != nil {
			return parsedCommand{}, err
		}
		command.value = value
	}
	return command, nil
}

func usageError() error {
	return &domain.Error{
		Category: "usage",
		Details:  map[string]any{"reason": "invalid_cli"},
	}
}

func writeError(stdout io.Writer, err error) int {
	var contractError *domain.Error
	if !errors.As(err, &contractError) {
		contractError = &domain.Error{
			Category: "storage_error",
			Details: map[string]any{
				"operation": "write",
				"reason":    "storage_failure",
			},
			Cause: err,
		}
	}
	_ = writeEnvelope(stdout, map[string]any{
		"ok": false,
		"error": map[string]any{
			"category": contractError.Category,
			"details":  contractError.Details,
		},
	})
	return contractError.ExitCode()
}

func writeEnvelope(stdout io.Writer, envelope any) error {
	encoded, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	_, err = stdout.Write(encoded)
	return err
}

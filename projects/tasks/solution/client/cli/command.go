// Package cli owns shared Task command parsing, output, and exit-code policy.
package cli

import (
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
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

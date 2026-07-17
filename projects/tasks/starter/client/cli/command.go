// Package cli owns shared Task command parsing, output, and exit-code policy.
package cli

import (
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
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

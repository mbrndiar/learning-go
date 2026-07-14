package taskmanager

import (
	"context"
	"errors"
)

// ErrTaskNotFound reports that a requested task does not exist. Every Storage
// implementation must return an error satisfying errors.Is(err, ErrTaskNotFound)
// for missing tasks so callers can branch uniformly.
var ErrTaskNotFound = errors.New("task not found")

// Storage abstracts task persistence. It is intentionally small and defined
// where it is consumed (by Manager) so backends stay decoupled from the domain
// package. All methods take a context so callers control cancellation and
// timeouts.
type Storage interface {
	// List returns every stored task.
	List(ctx context.Context) ([]Task, error)
	// Get returns the task with the given identifier, or ErrTaskNotFound.
	Get(ctx context.Context, id int) (Task, error)
	// Add stores a new task with the given title and returns it with an
	// assigned identifier.
	Add(ctx context.Context, title string) (Task, error)
	// Complete marks the task with the given identifier as done and returns
	// the updated task, or ErrTaskNotFound.
	Complete(ctx context.Context, id int) (Task, error)
	// Remove deletes the task with the given identifier, or ErrTaskNotFound.
	Remove(ctx context.Context, id int) error
}

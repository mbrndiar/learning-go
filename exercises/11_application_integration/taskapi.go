// Package taskapi implements a small task-tracking HTTP service, covering
// JSON validation, HTTP handlers and clients tested with httptest, context
// timeouts, middleware, and a database/sql repository backed by an in-memory
// fake driver instead of a real database.
package taskapi

import (
	"context"
	"errors"
	"time"
)

// maxTitleLen is the longest a Task's Title may be.
const maxTitleLen = 200

// ErrNotFound is returned by TaskStore implementations when a requested task
// does not exist.
var ErrNotFound = errors.New("taskapi: task not found")

// Task is a single to-do item exchanged over the wire as JSON and persisted
// through TaskStore. DueDate is optional.
type Task struct {
	ID      int64      `json:"id"`
	Title   string     `json:"title"`
	Done    bool       `json:"done"`
	DueDate *time.Time `json:"due_date,omitempty"`
}

// Validate reports whether t is acceptable input, given the current time
// now (passed explicitly rather than read from time.Now() so validation
// stays deterministic and testable). Title must be non-empty after trimming
// whitespace and at most maxTitleLen characters; DueDate, if set, must not
// be before now.
//
// TODO(task 1): implement Validate.
func (t Task) Validate(now time.Time) error {
	panic("not implemented")
}

// TaskStore is the persistence boundary the HTTP layer depends on. It is
// implemented by SQLTaskStore (backed by database/sql) and, in tests, by
// lightweight fakes -- the HTTP handlers never import database/sql
// themselves.
type TaskStore interface {
	// Create persists t, ignoring any t.ID, and returns t with its
	// assigned ID.
	Create(ctx context.Context, t Task) (Task, error)
	// Get returns the task with the given id, or ErrNotFound if none
	// exists.
	Get(ctx context.Context, id int64) (Task, error)
	// List returns every task, ordered by ID ascending.
	List(ctx context.Context) ([]Task, error)
}

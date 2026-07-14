// Package solution is the reference implementation for
// exercises/11_application_integration.
package solution

import (
	"context"
	"errors"
	"strings"
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
// now.
func (t Task) Validate(now time.Time) error {
	title := strings.TrimSpace(t.Title)
	if title == "" {
		return errors.New("title must not be empty")
	}
	if len(title) > maxTitleLen {
		return errors.New("title must be at most 200 characters")
	}
	if t.DueDate != nil && t.DueDate.Before(now) {
		return errors.New("due date must not be in the past")
	}
	return nil
}

// TaskStore is the persistence boundary the HTTP layer depends on.
type TaskStore interface {
	Create(ctx context.Context, t Task) (Task, error)
	Get(ctx context.Context, id int64) (Task, error)
	List(ctx context.Context) ([]Task, error)
}

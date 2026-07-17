// Package api defines transport-neutral HTTP boundary policy for Task routers.
// The wire-level source of truth remains projects/tasks/docs/SPEC.md and
// projects/tasks/docs/openapi.yaml.
package api

import (
	"context"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// Service is the application capability every router adapter must call to
// fulfill a routed request; it is satisfied by task.Service.
type Service interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context, task.ListFilter) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Delete(context.Context, int64) error
}

// Task is the wire representation of a task.Task in success responses.
type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Error is the machine-readable body carried inside ErrorEnvelope.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorEnvelope is the JSON shape every documented error response uses.
type ErrorEnvelope struct {
	Error Error `json:"error"`
}

// TaskDTO converts one domain task.Task into its wire representation.
func TaskDTO(value task.Task) Task {
	return Task{ID: value.ID, Title: value.Title, Completed: value.Completed}
}

// TaskDTOs converts a slice of domain tasks into their wire representation,
// preserving order.
func TaskDTOs(values []task.Task) []Task {
	return nil
}

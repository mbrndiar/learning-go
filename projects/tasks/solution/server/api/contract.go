// Package api defines transport-neutral HTTP boundary policy for Task routers.
package api

import (
	"context"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// Service is the transport-neutral boundary every adapter (nethttp, chi,
// gin, ...) depends on. Adapters use it instead of depending on a concrete
// task.Service or repository implementation.
type Service interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context, task.ListFilter) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Delete(context.Context, int64) error
}

// Task is the wire shape of a task, produced from the domain value by
// TaskDTO/TaskDTOs so adapters never serialize task.Task directly.
type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Error is the wire shape of one error inside an ErrorEnvelope.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorEnvelope is the single JSON error shape returned by every adapter.
type ErrorEnvelope struct {
	Error Error `json:"error"`
}

// TaskDTO converts one domain task to its wire representation.
func TaskDTO(value task.Task) Task {
	return Task{ID: value.ID, Title: value.Title, Completed: value.Completed}
}

// TaskDTOs converts a slice of domain tasks to their wire representation,
// preserving order.
func TaskDTOs(values []task.Task) []Task {
	result := make([]Task, len(values))
	for index, value := range values {
		result[index] = TaskDTO(value)
	}
	return result
}

package markdown

import (
	"context"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// Repository will store tasks in one Markdown checklist.
type Repository struct {
	placeholder struct{}
}

var _ task.Repository = (*Repository)(nil)

// Open is an exercise placeholder. It should become a compatibility wrapper
// around OpenContext using context.Background().
func Open(path string) (*Repository, error) {
	return OpenContext(context.Background(), path)
}

// OpenContext is an exercise placeholder.
func OpenContext(ctx context.Context, path string) (*Repository, error) {
	return nil, task.ErrNotImplemented
}

// Create is an exercise placeholder.
func (r *Repository) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

// List is an exercise placeholder.
func (r *Repository) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	return nil, task.ErrNotImplemented
}

// Get is an exercise placeholder.
func (r *Repository) Get(ctx context.Context, id int64) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

// Update is an exercise placeholder.
func (r *Repository) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

// Delete is an exercise placeholder.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return task.ErrNotImplemented
}

package client

import (
	"context"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// Transport is the remote Task capability consumed by command policy.
type Transport interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context, task.ListFilter) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Delete(context.Context, int64) error
}

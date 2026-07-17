package task

import "context"

// Repository is the persistence capability consumed by Service.
type Repository interface {
	Create(context.Context, CreateInput) (Task, error)
	List(context.Context, ListFilter) ([]Task, error)
	Get(context.Context, int64) (Task, error)
	Update(context.Context, int64, UpdateInput) (Task, error)
	Delete(context.Context, int64) error
}

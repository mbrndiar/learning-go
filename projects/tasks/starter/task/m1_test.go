package task_test

import (
	"context"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m1"
)

func TestMilestone1StarterIsExplicitlyIncomplete(t *testing.T) {
	repository := &countingRepository{}
	service := task.NewService(repository)
	title := "title"
	completed := false

	m1.AssertStarterExplicit(t, task.ErrNotImplemented, func() int {
		return repository.calls
	}, []func() error{
		func() error {
			_, err := service.Create(context.Background(), task.CreateInput{Title: title})
			return err
		},
		func() error {
			_, err := service.List(context.Background(), task.ListFilter{Completed: &completed})
			return err
		},
		func() error {
			_, err := service.Get(context.Background(), 1)
			return err
		},
		func() error {
			_, err := service.Update(context.Background(), 1, task.UpdateInput{Title: &title})
			return err
		},
		func() error {
			return service.Delete(context.Background(), 1)
		},
		func() error {
			_, err := task.NormalizeTitle(title)
			return err
		},
		func() error {
			return task.ValidateTitle(title)
		},
		func() error {
			return task.ValidateID(1)
		},
		func() error {
			_, err := task.NormalizeUpdate(task.UpdateInput{Title: &title})
			return err
		},
		func() error {
			return task.ValidateUpdate(task.UpdateInput{Title: &title})
		},
		func() error {
			_, err := task.NormalizeListFilter(task.ListFilter{})
			return err
		},
		func() error {
			return task.ValidateListFilter(task.ListFilter{})
		},
		func() error {
			return task.ValidateTask(task.Task{ID: 1, Title: title})
		},
	})
}

type countingRepository struct {
	calls int
}

func (r *countingRepository) Create(context.Context, task.CreateInput) (task.Task, error) {
	r.calls++
	return task.Task{}, nil
}

func (r *countingRepository) List(context.Context, task.ListFilter) ([]task.Task, error) {
	r.calls++
	return nil, nil
}

func (r *countingRepository) Get(context.Context, int64) (task.Task, error) {
	r.calls++
	return task.Task{}, nil
}

func (r *countingRepository) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	r.calls++
	return task.Task{}, nil
}

func (r *countingRepository) Delete(context.Context, int64) error {
	r.calls++
	return nil
}

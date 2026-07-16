package task_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m1"
)

func TestMilestone1(t *testing.T) {
	m1.AssertSolutionTask(t, m1.TaskHarness{
		MaxTitleLength:      task.MaxTitleLength,
		NormalizeTitle:      task.NormalizeTitle,
		ValidateTitle:       task.ValidateTitle,
		ValidateID:          task.ValidateID,
		NormalizeUpdate:     normalizeUpdate,
		NormalizeListFilter: normalizeFilter,
		ValidateTask: func(value m1.Task) error {
			return task.ValidateTask(toTask(value))
		},
		NewService: func(repository m1.Repository) m1.Service {
			return serviceAdapter{service: task.NewService(repositoryAdapter{repository: repository})}
		},
		IsValidation: func(err error) bool {
			return errors.Is(err, task.ErrValidation)
		},
		ValidationDetails: func(err error) (string, string, bool) {
			var target *task.ValidationError
			if !errors.As(err, &target) {
				return "", "", false
			}
			return target.Field, target.Message, true
		},
		NewNotFound: func(id int64) error {
			return task.NewNotFoundError(id)
		},
		IsNotFound: func(err error) bool {
			return errors.Is(err, task.ErrNotFound)
		},
		NotFoundID: func(err error) (int64, bool) {
			var target *task.NotFoundError
			if !errors.As(err, &target) {
				return 0, false
			}
			return target.ID, true
		},
		WrapStorage: task.WrapStorage,
		IsStorage: func(err error) bool {
			return errors.Is(err, task.ErrStorage)
		},
	})
}

func TestExportedValidationHelpers(t *testing.T) {
	title := "Learn adapters"
	completed := false
	if err := task.ValidateUpdate(task.UpdateInput{Title: &title, Completed: &completed}); err != nil {
		t.Fatalf("ValidateUpdate(valid) = %v", err)
	}
	padded := " Learn adapters "
	if err := task.ValidateUpdate(task.UpdateInput{Title: &padded}); !errors.Is(err, task.ErrValidation) {
		t.Fatalf("ValidateUpdate(padded) = %v", err)
	}
	if err := task.ValidateUpdate(task.UpdateInput{}); !errors.Is(err, task.ErrValidation) {
		t.Fatalf("ValidateUpdate(empty) = %v", err)
	}
	if err := task.ValidateListFilter(task.ListFilter{}); err != nil {
		t.Fatalf("ValidateListFilter(empty) = %v", err)
	}
	if err := task.ValidateListFilter(task.ListFilter{Completed: &completed}); err != nil {
		t.Fatalf("ValidateListFilter(false) = %v", err)
	}
}

type repositoryAdapter struct {
	repository m1.Repository
}

func (r repositoryAdapter) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	value, err := r.repository.Create(ctx, m1.CreateInput{Title: input.Title})
	return toTask(value), err
}

func (r repositoryAdapter) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	values, err := r.repository.List(ctx, m1.ListFilter{Completed: filter.Completed})
	tasks := make([]task.Task, len(values))
	for index, value := range values {
		tasks[index] = toTask(value)
	}
	return tasks, err
}

func (r repositoryAdapter) Get(ctx context.Context, id int64) (task.Task, error) {
	value, err := r.repository.Get(ctx, id)
	return toTask(value), err
}

func (r repositoryAdapter) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	value, err := r.repository.Update(ctx, id, m1.UpdateInput{
		Title:     input.Title,
		Completed: input.Completed,
	})
	return toTask(value), err
}

func (r repositoryAdapter) Delete(ctx context.Context, id int64) error {
	return r.repository.Delete(ctx, id)
}

type serviceAdapter struct {
	service *task.Service
}

func (s serviceAdapter) Create(ctx context.Context, input m1.CreateInput) (m1.Task, error) {
	value, err := s.service.Create(ctx, task.CreateInput{Title: input.Title})
	return fromTask(value), err
}

func (s serviceAdapter) List(ctx context.Context, filter m1.ListFilter) ([]m1.Task, error) {
	values, err := s.service.List(ctx, task.ListFilter{Completed: filter.Completed})
	tasks := make([]m1.Task, len(values))
	for index, value := range values {
		tasks[index] = fromTask(value)
	}
	return tasks, err
}

func (s serviceAdapter) Get(ctx context.Context, id int64) (m1.Task, error) {
	value, err := s.service.Get(ctx, id)
	return fromTask(value), err
}

func (s serviceAdapter) Update(ctx context.Context, id int64, input m1.UpdateInput) (m1.Task, error) {
	value, err := s.service.Update(ctx, id, task.UpdateInput{
		Title:     input.Title,
		Completed: input.Completed,
	})
	return fromTask(value), err
}

func (s serviceAdapter) Delete(ctx context.Context, id int64) error {
	return s.service.Delete(ctx, id)
}

func normalizeUpdate(input m1.UpdateInput) (m1.UpdateInput, error) {
	value, err := task.NormalizeUpdate(task.UpdateInput{
		Title:     input.Title,
		Completed: input.Completed,
	})
	return m1.UpdateInput{Title: value.Title, Completed: value.Completed}, err
}

func normalizeFilter(filter m1.ListFilter) (m1.ListFilter, error) {
	value, err := task.NormalizeListFilter(task.ListFilter{Completed: filter.Completed})
	return m1.ListFilter{Completed: value.Completed}, err
}

func toTask(value m1.Task) task.Task {
	return task.Task{ID: value.ID, Title: value.Title, Completed: value.Completed}
}

func fromTask(value task.Task) m1.Task {
	return m1.Task{ID: value.ID, Title: value.Title, Completed: value.Completed}
}

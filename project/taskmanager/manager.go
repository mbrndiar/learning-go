package taskmanager

import (
	"context"
	"errors"
	"fmt"
)

// Manager applies domain rules and delegates persistence to a Storage. It is
// the single entry point CLIs and other callers use, keeping validation in one
// place regardless of which backend is configured.
type Manager struct {
	storage Storage
}

// NewManager returns a Manager backed by the given storage. The storage must
// not be nil.
func NewManager(storage Storage) (*Manager, error) {
	if storage == nil {
		return nil, errors.New("taskmanager: storage must not be nil")
	}
	return &Manager{storage: storage}, nil
}

// List returns every task.
func (m *Manager) List(ctx context.Context) ([]Task, error) {
	tasks, err := m.storage.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("taskmanager: list tasks: %w", err)
	}
	return tasks, nil
}

// Get returns a single task by identifier.
func (m *Manager) Get(ctx context.Context, id int) (Task, error) {
	if id <= 0 {
		return Task{}, fmt.Errorf("%w: got %d", ErrInvalidID, id)
	}
	task, err := m.storage.Get(ctx, id)
	if err != nil {
		return Task{}, fmt.Errorf("taskmanager: get task %d: %w", id, err)
	}
	return task, nil
}

// Add validates the title and stores a new task.
func (m *Manager) Add(ctx context.Context, title string) (Task, error) {
	normalized, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, fmt.Errorf("taskmanager: add task: %w", err)
	}
	task, err := m.storage.Add(ctx, normalized)
	if err != nil {
		return Task{}, fmt.Errorf("taskmanager: add task: %w", err)
	}
	return task, nil
}

// Complete marks a task as done.
func (m *Manager) Complete(ctx context.Context, id int) (Task, error) {
	if id <= 0 {
		return Task{}, fmt.Errorf("%w: got %d", ErrInvalidID, id)
	}
	task, err := m.storage.Complete(ctx, id)
	if err != nil {
		return Task{}, fmt.Errorf("taskmanager: complete task %d: %w", id, err)
	}
	return task, nil
}

// Remove deletes a task.
func (m *Manager) Remove(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("%w: got %d", ErrInvalidID, id)
	}
	if err := m.storage.Remove(ctx, id); err != nil {
		return fmt.Errorf("taskmanager: remove task %d: %w", id, err)
	}
	return nil
}

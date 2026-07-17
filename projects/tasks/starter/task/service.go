package task

import "context"

// Service coordinates validation and repository operations.
type Service struct {
	repository Repository
}

// NewService creates a task service.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// Create validates input, persists a task through repository, and returns
// the stored Task.
func (s *Service) Create(ctx context.Context, input CreateInput) (Task, error) {
	return Task{}, ErrNotImplemented
}

// List validates filter and returns matching tasks ordered by ID.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]Task, error) {
	return nil, ErrNotImplemented
}

// Get validates id and returns the matching task, or a NotFoundError.
func (s *Service) Get(ctx context.Context, id int64) (Task, error) {
	return Task{}, ErrNotImplemented
}

// Update validates id and input, applies the present fields, and returns the
// updated task.
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (Task, error) {
	return Task{}, ErrNotImplemented
}

// Delete validates id and removes the matching task.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return ErrNotImplemented
}

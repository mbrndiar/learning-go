package task

import (
	"context"
	"fmt"
)

// Service coordinates validation and repository operations.
type Service struct {
	repository Repository
}

// NewService creates a task service.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// Create validates and normalizes input before creating a task.
func (s *Service) Create(ctx context.Context, input CreateInput) (Task, error) {
	title, err := NormalizeTitle(input.Title)
	if err != nil {
		return Task{}, err
	}
	if s == nil || s.repository == nil {
		return Task{}, fmt.Errorf("%w: repository is required", ErrStorage)
	}
	return s.repository.Create(ctx, CreateInput{Title: title})
}

// List returns tasks matching an optional completion filter.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]Task, error) {
	normalized, err := NormalizeListFilter(filter)
	if err != nil {
		return nil, err
	}
	if s == nil || s.repository == nil {
		return nil, fmt.Errorf("%w: repository is required", ErrStorage)
	}
	return s.repository.List(ctx, normalized)
}

// Get returns one task after validating its ID.
func (s *Service) Get(ctx context.Context, id int64) (Task, error) {
	if err := ValidateID(id); err != nil {
		return Task{}, err
	}
	if s == nil || s.repository == nil {
		return Task{}, fmt.Errorf("%w: repository is required", ErrStorage)
	}
	return s.repository.Get(ctx, id)
}

// Update applies a validated partial update.
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (Task, error) {
	if err := ValidateID(id); err != nil {
		return Task{}, err
	}
	normalized, err := NormalizeUpdate(input)
	if err != nil {
		return Task{}, err
	}
	if s == nil || s.repository == nil {
		return Task{}, fmt.Errorf("%w: repository is required", ErrStorage)
	}
	return s.repository.Update(ctx, id, normalized)
}

// Delete removes one task after validating its ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := ValidateID(id); err != nil {
		return err
	}
	if s == nil || s.repository == nil {
		return fmt.Errorf("%w: repository is required", ErrStorage)
	}
	return s.repository.Delete(ctx, id)
}

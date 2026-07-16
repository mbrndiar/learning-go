// Package task defines the storage- and transport-independent project boundary.
package task

import (
	"context"
	"errors"
	"fmt"
)

// Implemented reports whether the reference implementation is selected.
const Implemented = false

// MaxTitleLength is the maximum number of Unicode characters in a title.
const MaxTitleLength = 120

var (
	// ErrNotImplemented marks an intentional learner placeholder.
	ErrNotImplemented = errors.New("tasks project: not implemented")
	// ErrValidation classifies invalid domain input.
	ErrValidation = errors.New("tasks project: validation error")
	// ErrNotFound classifies a missing task.
	ErrNotFound = errors.New("tasks project: task not found")
	// ErrStorage classifies persistence failures.
	ErrStorage = errors.New("tasks project: storage error")
)

// Task is the transport- and storage-independent task value.
type Task struct {
	ID        int64
	Title     string
	Completed bool
}

// CreateInput contains values accepted when creating a task.
type CreateInput struct {
	Title string
}

// UpdateInput is a partial task update. Nil fields are omitted.
type UpdateInput struct {
	Title     *string
	Completed *bool
}

// ListFilter optionally limits tasks by completion state.
type ListFilter struct {
	Completed *bool
}

// ValidationError describes one invalid field.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements error.
func (e *ValidationError) Error() string {
	if e == nil {
		return ErrValidation.Error()
	}
	return e.Message
}

// Unwrap classifies the error as ErrValidation.
func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

// NotFoundError identifies the task that was not found.
type NotFoundError struct {
	ID int64
}

// Error implements error.
func (e *NotFoundError) Error() string {
	if e == nil {
		return ErrNotFound.Error()
	}
	return fmt.Sprintf("task %d was not found", e.ID)
}

// Unwrap classifies the error as ErrNotFound.
func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// StorageError preserves a persistence operation and its underlying failure.
type StorageError struct {
	Operation string
	Err       error
}

// Error implements error.
func (e *StorageError) Error() string {
	if e == nil {
		return ErrStorage.Error()
	}
	if e.Err == nil {
		return fmt.Sprintf("task storage %s failed", e.Operation)
	}
	return fmt.Sprintf("task storage %s failed: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying persistence failure.
func (e *StorageError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is classifies every StorageError as ErrStorage.
func (e *StorageError) Is(target error) bool {
	return target == ErrStorage
}

// NewNotFoundError constructs the canonical missing-task error.
func NewNotFoundError(id int64) *NotFoundError {
	return &NotFoundError{ID: id}
}

// WrapStorage adds storage classification and operation context.
func WrapStorage(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &StorageError{Operation: operation, Err: err}
}

// Repository is the persistence capability consumed by Service.
type Repository interface {
	Create(context.Context, CreateInput) (Task, error)
	List(context.Context, ListFilter) ([]Task, error)
	Get(context.Context, int64) (Task, error)
	Update(context.Context, int64, UpdateInput) (Task, error)
	Delete(context.Context, int64) error
}

// Service coordinates validation and repository operations.
type Service struct {
	repository Repository
}

// NewService creates a task service.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// Create is an exercise placeholder.
func (s *Service) Create(ctx context.Context, input CreateInput) (Task, error) {
	return Task{}, ErrNotImplemented
}

// List is an exercise placeholder.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]Task, error) {
	return nil, ErrNotImplemented
}

// Get is an exercise placeholder.
func (s *Service) Get(ctx context.Context, id int64) (Task, error) {
	return Task{}, ErrNotImplemented
}

// Update is an exercise placeholder.
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (Task, error) {
	return Task{}, ErrNotImplemented
}

// Delete is an exercise placeholder.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return ErrNotImplemented
}

// NormalizeTitle is an exercise placeholder.
func NormalizeTitle(title string) (string, error) {
	return "", ErrNotImplemented
}

// ValidateTitle is an exercise placeholder.
func ValidateTitle(title string) error {
	return ErrNotImplemented
}

// ValidateID is an exercise placeholder.
func ValidateID(id int64) error {
	return ErrNotImplemented
}

// NormalizeUpdate is an exercise placeholder.
func NormalizeUpdate(input UpdateInput) (UpdateInput, error) {
	return UpdateInput{}, ErrNotImplemented
}

// ValidateUpdate is an exercise placeholder.
func ValidateUpdate(input UpdateInput) error {
	return ErrNotImplemented
}

// NormalizeListFilter is an exercise placeholder.
func NormalizeListFilter(filter ListFilter) (ListFilter, error) {
	return ListFilter{}, ErrNotImplemented
}

// ValidateListFilter is an exercise placeholder.
func ValidateListFilter(filter ListFilter) error {
	return ErrNotImplemented
}

// ValidateTask is an exercise placeholder.
func ValidateTask(value Task) error {
	return ErrNotImplemented
}

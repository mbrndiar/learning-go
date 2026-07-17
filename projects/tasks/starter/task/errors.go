package task

import (
	"errors"
	"fmt"
)

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

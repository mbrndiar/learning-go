// Package task defines the storage- and transport-independent project boundary.
package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Implemented reports whether the reference implementation is selected.
const Implemented = true

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

// NormalizeTitle trims and validates a title.
func NormalizeTitle(title string) (string, error) {
	if !utf8.ValidString(title) {
		return "", validationError("title", "title must contain valid UTF-8")
	}
	for _, r := range title {
		if r == '\n' || r == '\r' || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return "", validationError("title", "title must occupy one physical line")
		}
		if unicode.IsControl(r) {
			return "", validationError("title", "title must not contain control characters")
		}
	}
	title = strings.TrimSpace(title)
	count := utf8.RuneCountInString(title)
	if count < 1 || count > MaxTitleLength {
		return "", validationError("title", "title must contain between 1 and 120 characters")
	}
	return title, nil
}

// ValidateTitle reports whether a title is already normalized and valid.
func ValidateTitle(title string) error {
	normalized, err := NormalizeTitle(title)
	if err != nil {
		return err
	}
	if normalized != title {
		return validationError("title", "title must not have leading or trailing whitespace")
	}
	return nil
}

// ValidateID requires a positive task ID.
func ValidateID(id int64) error {
	if id <= 0 {
		return validationError("id", "id must be a positive integer")
	}
	return nil
}

// NormalizeUpdate validates a partial update and returns normalized copies.
func NormalizeUpdate(input UpdateInput) (UpdateInput, error) {
	if input.Title == nil && input.Completed == nil {
		return UpdateInput{}, validationError("body", "update must include title or completed")
	}

	var normalized UpdateInput
	if input.Title != nil {
		title, err := NormalizeTitle(*input.Title)
		if err != nil {
			return UpdateInput{}, err
		}
		normalized.Title = &title
	}
	if input.Completed != nil {
		completed := *input.Completed
		normalized.Completed = &completed
	}
	return normalized, nil
}

// ValidateUpdate reports whether a partial update is already normalized.
func ValidateUpdate(input UpdateInput) error {
	normalized, err := NormalizeUpdate(input)
	if err != nil {
		return err
	}
	if input.Title != nil && *normalized.Title != *input.Title {
		return validationError("title", "title must not have leading or trailing whitespace")
	}
	return nil
}

// NormalizeListFilter copies an optional filter without losing explicit false.
func NormalizeListFilter(filter ListFilter) (ListFilter, error) {
	if filter.Completed == nil {
		return ListFilter{}, nil
	}
	completed := *filter.Completed
	return ListFilter{Completed: &completed}, nil
}

// ValidateListFilter validates a completion filter.
func ValidateListFilter(filter ListFilter) error {
	_, err := NormalizeListFilter(filter)
	return err
}

// ValidateTask validates a task value returned by storage or a remote server.
func ValidateTask(value Task) error {
	if err := ValidateID(value.ID); err != nil {
		return err
	}
	return ValidateTitle(value.Title)
}

func validationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

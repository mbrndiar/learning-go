// Package taskmanager owns the task domain model and coordinates storage.
//
// It defines a validated Task, a small consumer-owned Storage interface, and a
// Manager that applies domain rules before delegating persistence. Two Storage
// implementations ship with the package: an atomic JSON FileStorage for local
// use and a RESTStorage that talks to the remote task API through taskclient.
package taskmanager

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// MaxTitleLength bounds a task title, matching the API-side limit so local and
// remote storage agree on what a valid task looks like.
const MaxTitleLength = 256

// Validation errors are exported as sentinels so callers can branch with
// errors.Is instead of matching message text.
var (
	// ErrEmptyTitle reports that a title was empty after trimming whitespace.
	ErrEmptyTitle = errors.New("task title must not be empty")
	// ErrTitleTooLong reports that a title exceeded MaxTitleLength runes.
	ErrTitleTooLong = errors.New("task title is too long")
	// ErrInvalidTitle reports that a title contained invalid UTF-8.
	ErrInvalidTitle = errors.New("task title is not valid UTF-8")
	// ErrInvalidID reports that a task identifier was not positive.
	ErrInvalidID = errors.New("task id must be positive")
)

// Task is the domain model shared across storage backends.
type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Validate reports whether the task satisfies the domain invariants. A stored
// task must have a positive identifier and a valid, non-empty title.
func (t Task) Validate() error {
	if t.ID <= 0 {
		return fmt.Errorf("%w: got %d", ErrInvalidID, t.ID)
	}
	if _, err := NormalizeTitle(t.Title); err != nil {
		return err
	}
	return nil
}

// NormalizeTitle trims surrounding whitespace and validates a candidate title,
// returning the cleaned value. It is used before persisting new tasks so the
// same rules apply regardless of storage backend.
func NormalizeTitle(title string) (string, error) {
	if !utf8.ValidString(title) {
		return "", ErrInvalidTitle
	}
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", ErrEmptyTitle
	}
	if utf8.RuneCountInString(trimmed) > MaxTitleLength {
		return "", fmt.Errorf("%w: %d runes", ErrTitleTooLong, utf8.RuneCountInString(trimmed))
	}
	return trimmed, nil
}

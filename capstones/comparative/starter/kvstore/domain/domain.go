// Package domain defines the comparative capstone's storage-independent API.
package domain

import (
	"encoding/json"
	"errors"
)

// ErrNotImplemented marks the incomplete boundary used by the starter harness.
// The complete solution retains the symbol for API parity but never returns it.
var ErrNotImplemented = errors.New("comparative kvstore: not implemented")

// Implemented reports whether the harness placeholders have been replaced.
const Implemented = false

// Revision is a global successful-mutation sequence number.
type Revision int64

// MaxRevision is the largest revision permitted by the shared specification.
const MaxRevision Revision = 9007199254740991

// ExpectationKind identifies a conditional mutation rule.
type ExpectationKind string

const (
	ExpectAny    ExpectationKind = "any"
	ExpectAbsent ExpectationKind = "absent"
	ExpectExact  ExpectationKind = "exact"
)

// Expectation is the parsed condition attached to a set or delete.
type Expectation struct {
	Kind     ExpectationKind
	Revision Revision
}

// Value is one normalized restricted-JSON value.
type Value = any

// Entry is the observable representation of a stored key.
type Entry struct {
	Key      string   `json:"key"`
	Value    Value    `json:"value"`
	Revision Revision `json:"revision"`
}

// SetResult is returned after a successful set.
type SetResult struct {
	Key      string   `json:"key"`
	Value    Value    `json:"value"`
	Revision Revision `json:"revision"`
	Created  bool     `json:"created"`
}

// DeleteResult is returned after a successful delete.
type DeleteResult struct {
	Key             string   `json:"key"`
	DeletedRevision Revision `json:"deleted_revision"`
	Revision        Revision `json:"revision"`
}

// ListResult is returned by list.
type ListResult struct {
	Entries        []Entry  `json:"entries"`
	GlobalRevision Revision `json:"global_revision"`
}

// Error is a shared-contract error with an optional wrapped cause.
type Error struct {
	Category string         `json:"category"`
	Details  map[string]any `json:"details"`
	Cause    error          `json:"-"`
}

func (e *Error) Error() string {
	return e.Category
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// ExitCode returns the normative process exit code for the error.
func (e *Error) ExitCode() int {
	return 1
}

// ParseKey validates and returns a shared-contract key.
func ParseKey(value string) (string, error) {
	return "", ErrNotImplemented
}

// ParseExpectation parses a set or delete expectation.
func ParseExpectation(value string, allowAbsent bool) (Expectation, error) {
	return Expectation{}, ErrNotImplemented
}

// ParseValue parses and normalizes one restricted JSON value.
func ParseValue(input json.RawMessage) (Value, error) {
	return nil, ErrNotImplemented
}

// ParseStoredValue parses a value already persisted in normalized form.
func ParseStoredValue(input string) (Value, error) {
	return nil, ErrNotImplemented
}

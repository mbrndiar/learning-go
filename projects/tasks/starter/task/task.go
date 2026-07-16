// Package task defines the storage- and transport-independent project boundary.
package task

import "errors"

// ErrNotImplemented marks an intentional learner placeholder.
var ErrNotImplemented = errors.New("tasks project: not implemented")

// Implemented reports whether the reference implementation is selected.
const Implemented = false

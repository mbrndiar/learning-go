// Package task defines the storage- and transport-independent project boundary.
// Learner placeholders expose the required contracts; detailed invariants live
// in projects/tasks/docs/SPEC.md.
package task

// Implemented reports whether the reference implementation is selected.
const Implemented = false

// MaxTitleLength is the maximum number of Unicode characters in a title.
const MaxTitleLength = 120

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

// UpdateInput is a validated partial update. Nil means absent; boundary
// adapters must reject explicit null before constructing this value.
type UpdateInput struct {
	Title     *string
	Completed *bool
}

// ListFilter optionally limits tasks by completion state.
type ListFilter struct {
	Completed *bool
}

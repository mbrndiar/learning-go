// Package collections provides small generic collection types and
// functions to practice type parameters, constraints, and designing a clear
// exported API boundary around unexported internal state.
//
// Implement every function and method below. Replace each
// panic("not implemented") with working code; do not change any signature.
package collections

// Number is a constraint satisfied by any built-in integer or floating
// point type, including named types with an underlying numeric type.
//
// TODO(task 1): declare the Number interface constraint so that Sum
// compiles for int, int64, and float64 values.
type Number interface {
	// TODO(task 1): list the permitted underlying types, e.g. ~int | ~int64 | ~float64.
}

// Sum returns the total of every value in values. It returns the zero value
// of T for an empty or nil slice.
//
// TODO(task 1): implement Sum.
func Sum[T Number](values []T) T {
	panic("not implemented")
}

// Map applies fn to every element of values and returns the resulting
// slice, preserving order. It returns an empty, non-nil slice for an empty
// input.
//
// TODO(task 2): implement Map.
func Map[T, U any](values []T, fn func(T) U) []U {
	panic("not implemented")
}

// Filter returns a new slice containing only the elements of values for
// which predicate returns true, preserving order. It returns an empty,
// non-nil slice if nothing matches.
//
// TODO(task 3): implement Filter.
func Filter[T any](values []T, predicate func(T) bool) []T {
	panic("not implemented")
}

// Reduce folds values into a single accumulated result, starting from
// initial and calling fn(accumulator, element) for each element in order.
//
// TODO(task 4): implement Reduce.
func Reduce[T, U any](values []T, initial U, fn func(U, T) U) U {
	panic("not implemented")
}

// Stack is a generic last-in-first-out collection. The zero value is an
// empty, ready-to-use Stack. Its internal storage is unexported so callers
// can only observe it through the exported methods below.
type Stack[T any] struct {
	items []T
}

// Push adds v to the top of the stack.
//
// TODO(task 5): implement Push for Stack.
func (s *Stack[T]) Push(v T) {
	panic("not implemented")
}

// Pop removes and returns the item at the top of the stack. It returns the
// zero value of T and ok=false if the stack is empty.
//
// TODO(task 5): implement Pop for Stack.
func (s *Stack[T]) Pop() (v T, ok bool) {
	panic("not implemented")
}

// Peek returns the item at the top of the stack without removing it. It
// returns the zero value of T and ok=false if the stack is empty.
//
// TODO(task 5): implement Peek for Stack.
func (s *Stack[T]) Peek() (v T, ok bool) {
	panic("not implemented")
}

// Len returns the number of items currently on the stack.
//
// TODO(task 5): implement Len for Stack.
func (s *Stack[T]) Len() int {
	panic("not implemented")
}

// Queue is a generic first-in-first-out collection. The zero value is an
// empty, ready-to-use Queue. Its internal storage is unexported so callers
// can only observe it through the exported methods below.
type Queue[T any] struct {
	items []T
}

// Enqueue adds v to the back of the queue.
//
// TODO(task 6): implement Enqueue for Queue.
func (q *Queue[T]) Enqueue(v T) {
	panic("not implemented")
}

// Dequeue removes and returns the item at the front of the queue. It
// returns the zero value of T and ok=false if the queue is empty.
//
// TODO(task 6): implement Dequeue for Queue.
func (q *Queue[T]) Dequeue() (v T, ok bool) {
	panic("not implemented")
}

// Len returns the number of items currently in the queue.
//
// TODO(task 6): implement Len for Queue.
func (q *Queue[T]) Len() int {
	panic("not implemented")
}

// Package collections is the reference implementation for the packages and
// generics exercise. See ../collections.go for the task descriptions.
package collections

// Number is a constraint satisfied by any built-in integer or floating
// point type, including named types with an underlying numeric type.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Sum returns the total of every value in values. It returns the zero value
// of T for an empty or nil slice.
func Sum[T Number](values []T) T {
	var total T
	for _, v := range values {
		total += v
	}
	return total
}

// Map applies fn to every element of values and returns the resulting
// slice, preserving order. It returns an empty, non-nil slice for an empty
// input.
func Map[T, U any](values []T, fn func(T) U) []U {
	result := make([]U, 0, len(values))
	for _, v := range values {
		result = append(result, fn(v))
	}
	return result
}

// Filter returns a new slice containing only the elements of values for
// which predicate returns true, preserving order. It returns an empty,
// non-nil slice if nothing matches.
func Filter[T any](values []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range values {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce folds values into a single accumulated result, starting from
// initial and calling fn(accumulator, element) for each element in order.
func Reduce[T, U any](values []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range values {
		acc = fn(acc, v)
	}
	return acc
}

// Stack is a generic last-in-first-out collection. The zero value is an
// empty, ready-to-use Stack. Its internal storage is unexported so callers
// can only observe it through the exported methods below.
type Stack[T any] struct {
	items []T
}

// Push adds v to the top of the stack.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes and returns the item at the top of the stack. It returns the
// zero value of T and ok=false if the stack is empty.
func (s *Stack[T]) Pop() (v T, ok bool) {
	if len(s.items) == 0 {
		return v, false
	}
	last := len(s.items) - 1
	v = s.items[last]
	s.items = s.items[:last]
	return v, true
}

// Peek returns the item at the top of the stack without removing it. It
// returns the zero value of T and ok=false if the stack is empty.
func (s *Stack[T]) Peek() (v T, ok bool) {
	if len(s.items) == 0 {
		return v, false
	}
	return s.items[len(s.items)-1], true
}

// Len returns the number of items currently on the stack.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// Queue is a generic first-in-first-out collection. The zero value is an
// empty, ready-to-use Queue. Its internal storage is unexported so callers
// can only observe it through the exported methods below.
type Queue[T any] struct {
	items []T
}

// Enqueue adds v to the back of the queue.
func (q *Queue[T]) Enqueue(v T) {
	q.items = append(q.items, v)
}

// Dequeue removes and returns the item at the front of the queue. It
// returns the zero value of T and ok=false if the queue is empty.
func (q *Queue[T]) Dequeue() (v T, ok bool) {
	if len(q.items) == 0 {
		return v, false
	}
	v = q.items[0]
	q.items = q.items[1:]
	return v, true
}

// Len returns the number of items currently in the queue.
func (q *Queue[T]) Len() int {
	return len(q.items)
}

// Package main covers iota-based enum-like types and the classic nil
// interface pitfall: a nil pointer stored in an interface is not a nil
// interface.
package main

import "fmt"

// Status is a distinct named type based on int, not a plain int. This stops
// arbitrary integers and values from other enum-like types from being passed
// where a Status is expected without an explicit conversion.
type Status int

// iota starts at 0 in each const block and increments by one per line,
// giving sequential values without writing them out by hand. Reordering
// these lines changes the numeric values, so treat the order as part of the
// type's contract once it is used elsewhere (for example, persisted to a
// file or database).
const (
	StatusPending Status = iota
	StatusActive
	StatusDone
	StatusCancelled
)

// String implements fmt.Stringer so Status prints as a readable label
// instead of a bare integer. This is the standard pattern for giving an
// iota-based type a human-readable form.
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusActive:
		return "active"
	case StatusDone:
		return "done"
	case StatusCancelled:
		return "cancelled"
	default:
		return fmt.Sprintf("Status(%d)", int(s))
	}
}

// NotFoundError is a typed error used to demonstrate the nil-interface
// pitfall below.
type NotFoundError struct {
	ID int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("item %d not found", e.ID)
}

// lookup returns *NotFoundError through a plain error return type. This
// function contains the bug the lesson warns about: it declares a typed nil
// pointer and returns it unconditionally, instead of returning the untyped
// nil literal when there is no error.
func lookup(found bool) error {
	var problem *NotFoundError
	if !found {
		problem = &NotFoundError{ID: 42}
	}
	// BUG: problem's static type is *NotFoundError. Even when problem is a
	// nil pointer, wrapping it in the error interface produces a non-nil
	// interface value, because an interface is nil only when both its type
	// and value are unset.
	return problem
}

// lookupFixed shows the correct pattern: only return a non-nil error value
// when there truly is an error; otherwise return the bare nil literal.
func lookupFixed(found bool) error {
	if !found {
		return &NotFoundError{ID: 42}
	}
	return nil
}

func main() {
	fmt.Println("status values:")
	for _, status := range []Status{StatusPending, StatusActive, StatusDone, StatusCancelled} {
		fmt.Printf("  %d -> %s\n", int(status), status)
	}

	// --- The nil interface pitfall ---

	buggy := lookup(true) // found = true, so no real error occurred
	//lint:ignore SA4023 This comparison intentionally demonstrates the typed-nil interface pitfall.
	if buggy != nil {
		// This branch runs even though nothing went wrong, because buggy
		// holds a non-nil interface (type *NotFoundError, value nil).
		fmt.Println("buggy result: unexpectedly non-nil error:", buggy)
	} else {
		fmt.Println("buggy result: nil as expected")
	}

	fixed := lookupFixed(true)
	if fixed != nil {
		fmt.Println("fixed result: unexpectedly non-nil error:", fixed)
	} else {
		fmt.Println("fixed result: nil as expected")
	}

	// Directly comparing a typed nil pointer with the interface it is
	// assigned to makes the rule concrete.
	var typedNilPointer *NotFoundError
	var asInterface error = typedNilPointer
	fmt.Println("typedNilPointer == nil:", typedNilPointer == nil)
	//lint:ignore SA4023 This comparison intentionally demonstrates that an interface holding a typed nil is non-nil.
	fmt.Println("asInterface == nil:    ", asInterface == nil)
}

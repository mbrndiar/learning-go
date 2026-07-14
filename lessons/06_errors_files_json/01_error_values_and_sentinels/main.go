// Package main introduces errors as ordinary values: sentinel errors and
// custom typed errors.
package main

import (
	"errors"
	"fmt"
)

// ErrNotFound is a sentinel error: a package-level value that callers can
// compare against. Exporting it lets other packages recognize this specific
// failure without parsing an error message.
var ErrNotFound = errors.New("item not found")

// ValidationError is a typed error: a custom type that carries structured
// information (which field, and what was wrong with it) instead of only a
// string. Any type with an Error() string method satisfies the built-in
// error interface, implicitly, just like any other interface in Go.
type ValidationError struct {
	Field   string
	Problem string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s %s", e.Field, e.Problem)
}

// catalog is a tiny in-memory lookup table used to produce both kinds of
// error deterministically.
var catalog = map[int]string{
	1: "keyboard",
	2: "monitor",
}

// find returns a catalog entry, or one of the two error kinds above. Errors
// are ordinary return values in Go: there are no exceptions, so a function
// that can fail returns an error as its last result and callers are expected
// to check it immediately.
func find(id int) (string, error) {
	if id < 0 {
		return "", &ValidationError{Field: "id", Problem: "must not be negative"}
	}
	name, ok := catalog[id]
	if !ok {
		return "", ErrNotFound
	}
	return name, nil
}

func main() {
	for _, id := range []int{1, 2, 99, -1} {
		name, err := find(id)
		if err != nil {
			// A sentinel error can be compared directly with == (or, more
			// robustly, errors.Is - covered in the next lesson) as long as
			// it has not been wrapped.
			if errors.Is(err, ErrNotFound) {
				fmt.Printf("id %d: not found\n", id)
				continue
			}

			// A typed error's extra fields are only reachable after a type
			// assertion or errors.As (also covered next). A plain
			// fmt.Println still works because Error() satisfies fmt's
			// formatting.
			var validationErr *ValidationError
			if errors.As(err, &validationErr) {
				fmt.Printf("id %d: %s (field=%s)\n", id, validationErr.Error(), validationErr.Field)
				continue
			}

			fmt.Printf("id %d: unexpected error: %v\n", id, err)
			continue
		}
		fmt.Printf("id %d: %s\n", id, name)
	}

	// Two sentinel errors with the same message are NOT equal: errors.New
	// always allocates a distinct value, so identity - not message text -
	// is what makes a sentinel comparable and reliable.
	twin := errors.New("item not found")
	fmt.Println("ErrNotFound == twin:", ErrNotFound == twin)
	fmt.Println("same message text:  ", ErrNotFound.Error() == twin.Error())
}

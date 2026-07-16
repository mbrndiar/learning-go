// Package m1 contains shared Milestone 1 contract assertions.
package m1

import (
	"errors"
	"testing"
)

// RequireErrorKind checks a sentinel category without coupling to an implementation tree.
func RequireErrorKind(t *testing.T, err, sentinel error) {
	t.Helper()
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want errors.Is(_, %v)", err, sentinel)
	}
}

// Package m2 contains shared Milestone 2 contract assertions.
package m2

import "testing"

// RequireResult checks implementation-neutral probe classification fields.
func RequireResult(t *testing.T, status, wantStatus, code, wantCode string) {
	t.Helper()
	if status != wantStatus || code != wantCode {
		t.Fatalf("probe = status %q code %q, want status %q code %q", status, code, wantStatus, wantCode)
	}
}

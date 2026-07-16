// Package m3 contains shared Milestone 3 contract assertions.
package m3

import (
	"context"
	"testing"
)

// RequireSignal waits deterministically for an owned scheduler event.
func RequireSignal(t *testing.T, ctx context.Context, signal <-chan struct{}, name string) {
	t.Helper()
	select {
	case <-signal:
	case <-ctx.Done():
		t.Fatalf("waiting for %s: %v", name, ctx.Err())
	}
}

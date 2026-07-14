package main

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"
)

func TestRunWithLeakPreventionCompletes(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	work := make(chan struct{}, 1)
	work <- struct{}{}

	if err := runWithLeakPrevention(ctx, work); err != nil {
		t.Fatalf("runWithLeakPrevention() error = %v, want nil", err)
	}
}

func TestRunWithLeakPreventionTimesOut(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	work := make(chan struct{}) // never sent on

	err := runWithLeakPrevention(ctx, work)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("runWithLeakPrevention() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestRunWithLeakPreventionDoesNotLeakGoroutines(t *testing.T) {
	// Not run in parallel with other subtests, so the goroutine count
	// comparison is not disturbed by unrelated concurrent tests.
	before := goroutineCountStable()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	work := make(chan struct{})
	if err := runWithLeakPrevention(ctx, work); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("runWithLeakPrevention() error = %v, want context.DeadlineExceeded", err)
	}

	after := goroutineCountStable()
	if after > before {
		t.Fatalf("goroutine count grew from %d to %d; suspected leak", before, after)
	}
}

func TestGoroutineCountStableConverges(t *testing.T) {
	t.Parallel()

	before := goroutineCountStable()
	// Start and let a goroutine finish before measuring again.
	done := make(chan struct{})
	go func() { close(done) }()
	<-done
	runtime.Gosched()

	after := goroutineCountStable()
	if after > before+1 {
		t.Fatalf("unexpected sustained goroutine growth: before=%d after=%d", before, after)
	}
}

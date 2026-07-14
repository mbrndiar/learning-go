package main

import (
	"context"
	"reflect"
	"testing"
)

func TestCountToCompletesWithoutCancellation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	out := make(chan int)
	finished := make(chan struct{})

	go countTo(ctx, 4, out, finished)

	var got []int
	for v := range out {
		got = append(got, v)
	}
	<-finished

	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("countTo() produced %v, want %v", got, want)
	}
}

func TestCollectUntilCanceledStopsGoroutine(t *testing.T) {
	t.Parallel()

	got := collectUntilCanceled(100, 3)
	want := []int{1, 2, 3}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectUntilCanceled() = %v, want %v", got, want)
	}
}

func TestCountToStopsOnAlreadyCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before countTo ever runs

	out := make(chan int)
	finished := make(chan struct{})
	go countTo(ctx, 5, out, finished)

	// countTo must close out (possibly with zero values sent) and signal
	// finished even though the context was already canceled.
	for range out {
	}
	<-finished

	if err := ctx.Err(); err != context.Canceled {
		t.Fatalf("ctx.Err() = %v, want %v", err, context.Canceled)
	}
}

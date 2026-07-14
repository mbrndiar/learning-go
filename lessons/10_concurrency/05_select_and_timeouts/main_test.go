package main

import (
	"errors"
	"testing"
	"time"
)

func TestFirstOf(t *testing.T) {
	t.Parallel()

	a := make(chan int, 1)
	b := make(chan int, 1)
	a <- 42

	// Only a has a value, so the result is deterministic even though
	// select would choose pseudo-randomly if both were ready.
	if got, want := firstOf(a, b), 42; got != want {
		t.Fatalf("firstOf() = %d, want %d", got, want)
	}
}

func TestTryReceiveNonBlockingWithValue(t *testing.T) {
	t.Parallel()

	ch := make(chan int, 1)
	ch <- 7

	value, ok := tryReceiveNonBlocking(ch)
	if !ok || value != 7 {
		t.Fatalf("tryReceiveNonBlocking() = (%d, %v), want (7, true)", value, ok)
	}
}

func TestTryReceiveNonBlockingEmpty(t *testing.T) {
	t.Parallel()

	ch := make(chan int)
	value, ok := tryReceiveNonBlocking(ch)
	if ok || value != 0 {
		t.Fatalf("tryReceiveNonBlocking() = (%d, %v), want (0, false)", value, ok)
	}
}

func TestAwaitWithTimeoutSucceeds(t *testing.T) {
	t.Parallel()

	ch := make(chan int, 1)
	ch <- 9

	value, err := awaitWithTimeout(ch, time.Second)
	if err != nil {
		t.Fatalf("awaitWithTimeout() error = %v, want nil", err)
	}
	if value != 9 {
		t.Fatalf("awaitWithTimeout() value = %d, want 9", value)
	}
}

func TestAwaitWithTimeoutExpires(t *testing.T) {
	t.Parallel()

	never := make(chan int)
	_, err := awaitWithTimeout(never, 10*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("awaitWithTimeout() error = %v, want ErrTimeout", err)
	}
}

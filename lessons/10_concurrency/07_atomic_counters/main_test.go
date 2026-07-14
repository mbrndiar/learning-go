package main

import "testing"

func TestIncrementAtomically(t *testing.T) {
	t.Parallel()

	c := &atomicCounter{}
	const goroutines = 200

	incrementAtomically(c, goroutines)

	if got, want := c.Value(), int64(goroutines); got != want {
		t.Fatalf("c.Value() = %d, want %d", got, want)
	}
}

func TestAtomicCounterZeroValue(t *testing.T) {
	t.Parallel()

	c := &atomicCounter{}
	if got, want := c.Value(), int64(0); got != want {
		t.Fatalf("new atomicCounter Value() = %d, want %d", got, want)
	}
}

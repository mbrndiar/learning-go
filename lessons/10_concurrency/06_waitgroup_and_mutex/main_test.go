package main

import "testing"

func TestIncrementConcurrently(t *testing.T) {
	t.Parallel()

	c := &counter{}
	const goroutines = 200

	incrementConcurrently(c, goroutines)

	if got, want := c.Value(), goroutines; got != want {
		t.Fatalf("c.Value() = %d, want %d", got, want)
	}
}

func TestCounterZeroValue(t *testing.T) {
	t.Parallel()

	c := &counter{}
	if got, want := c.Value(), 0; got != want {
		t.Fatalf("new counter Value() = %d, want %d", got, want)
	}
}

package main

import (
	"reflect"
	"testing"
)

func TestFillThenDrainWithinCapacity(t *testing.T) {
	t.Parallel()

	got := fillThenDrain(3, []string{"a", "b", "c"})
	want := []string{"a", "b", "c"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fillThenDrain() = %v, want %v", got, want)
	}
}

func TestFillThenDrainBeyondCapacity(t *testing.T) {
	t.Parallel()

	got := fillThenDrain(2, []string{"a", "b", "c", "d", "e"})
	want := []string{"a", "b", "c", "d", "e"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fillThenDrain() = %v, want %v", got, want)
	}
}

func TestBufferedChannelLenCap(t *testing.T) {
	t.Parallel()

	queue := make(chan string, 2)
	queue <- "first"

	if got, want := len(queue), 1; got != want {
		t.Fatalf("len(queue) = %d, want %d", got, want)
	}
	if got, want := cap(queue), 2; got != want {
		t.Fatalf("cap(queue) = %d, want %d", got, want)
	}
}

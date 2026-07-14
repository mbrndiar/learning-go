package main

import (
	"reflect"
	"testing"
)

func TestSquare(t *testing.T) {
	t.Parallel()

	if got := square(5); got != 25 {
		t.Fatalf("square(5) = %d, want 25", got)
	}
	if got := square(0); got != 0 {
		t.Fatalf("square(0) = %d, want 0", got)
	}
}

func TestSquareAll(t *testing.T) {
	t.Parallel()

	got := squareAll([]int{5, 3, 4, 1, 2})
	want := []int{1, 4, 9, 16, 25}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("squareAll() = %v, want %v", got, want)
	}
}

func TestSquareAllEmpty(t *testing.T) {
	t.Parallel()

	got := squareAll(nil)
	if len(got) != 0 {
		t.Fatalf("squareAll(nil) = %v, want empty slice", got)
	}
}

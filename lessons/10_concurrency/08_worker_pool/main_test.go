package main

import (
	"reflect"
	"testing"
)

func TestRunWorkerPool(t *testing.T) {
	t.Parallel()

	inputs := []int{1, 2, 3, 4, 5}
	got := runWorkerPool(inputs, 2)

	want := []result{
		{id: 0, value: 1},
		{id: 1, value: 4},
		{id: 2, value: 9},
		{id: 3, value: 16},
		{id: 4, value: 25},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("runWorkerPool() = %v, want %v", got, want)
	}
}

func TestRunWorkerPoolMoreWorkersThanJobs(t *testing.T) {
	t.Parallel()

	got := runWorkerPool([]int{2, 3}, 8)
	want := []result{
		{id: 0, value: 4},
		{id: 1, value: 9},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("runWorkerPool() = %v, want %v", got, want)
	}
}

func TestRunWorkerPoolEmptyInput(t *testing.T) {
	t.Parallel()

	got := runWorkerPool(nil, 4)
	if len(got) != 0 {
		t.Fatalf("runWorkerPool(nil) = %v, want empty slice", got)
	}
}

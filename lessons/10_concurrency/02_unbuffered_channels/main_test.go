package main

import (
	"reflect"
	"testing"
)

func TestProduceAndCollect(t *testing.T) {
	t.Parallel()

	numbers := make(chan int)
	go produce([]int{10, 20, 30}, numbers)

	got := collect(numbers)
	want := []int{10, 20, 30}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collect() = %v, want %v", got, want)
	}
}

func TestCollectOnClosedEmptyChannel(t *testing.T) {
	t.Parallel()

	empty := make(chan int)
	close(empty)

	got := collect(empty)
	if len(got) != 0 {
		t.Fatalf("collect(closed empty) = %v, want empty slice", got)
	}
}

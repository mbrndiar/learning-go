package main

import "testing"

func TestSumGenerate(t *testing.T) {
	t.Parallel()

	numbers := make(chan int)
	go generate(4, numbers)

	if got, want := sum(numbers), 10; got != want {
		t.Fatalf("sum(generate(4)) = %d, want %d", got, want)
	}
}

func TestTryReceiveFromClosedChannel(t *testing.T) {
	t.Parallel()

	closed := make(chan int)
	close(closed)

	value, ok := tryReceive(closed)
	if ok {
		t.Fatalf("tryReceive(closed) ok = %v, want false", ok)
	}
	if value != 0 {
		t.Fatalf("tryReceive(closed) value = %d, want zero value 0", value)
	}
}

func TestSendOnClosedChannelPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic when sending on a closed channel")
		}
	}()

	closed := make(chan int)
	close(closed)
	closed <- 1 // must panic; recover above turns it into a passing assertion
}

func TestDoubleCloseChannelPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic when closing a channel twice")
		}
	}()

	ch := make(chan int)
	close(ch)
	close(ch) // must panic
}

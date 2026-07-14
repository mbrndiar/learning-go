// Command 04_channel_direction_and_closing shows directional channel types
// in function signatures and the ownership rule for closing channels: only
// the sender should close a channel, and only when it will never send again.
package main

import "fmt"

// generate owns the out channel: it is the only goroutine that sends on it,
// so it is the only one allowed to close it. The chan<- type means the
// compiler will reject any attempt to receive from out inside this
// function, catching direction mistakes at build time.
func generate(count int, out chan<- int) {
	defer close(out)
	for i := 1; i <= count; i++ {
		out <- i
	}
}

// sum only ever reads from in, enforced by the <-chan type. It never closes
// in, because it does not own it; closing a channel you only read from is a
// common bug this signature prevents.
func sum(in <-chan int) int {
	total := 0
	for v := range in {
		total += v
	}
	return total
}

// tryReceive demonstrates the comma-ok idiom for detecting a closed
// channel without a range loop: ok is false once the channel is both
// closed and drained.
func tryReceive(in <-chan int) (value int, ok bool) {
	value, ok = <-in
	return value, ok
}

func main() {
	numbers := make(chan int)
	go generate(3, numbers)
	fmt.Println("sum:", sum(numbers))

	closed := make(chan int)
	close(closed)
	value, ok := tryReceive(closed)
	fmt.Printf("receive from closed empty channel: value=%d ok=%v\n", value, ok)

	// Common mistakes, shown here only in comments because running them
	// would panic and crash the program:
	//
	//   close(closed) // panic: close of closed channel
	//   closed <- 1   // panic: send on closed channel
	//
	// Only the single owning sender should ever call close, and only after
	// it has sent its last value.
}

// Command 02_unbuffered_channels shows that an unbuffered channel makes the
// sender and receiver rendezvous: a send blocks until another goroutine is
// ready to receive, and vice versa.
package main

import "fmt"

// produce sends every value in order on out, then closes it. Closing is the
// sender's responsibility because only the sender knows when no more values
// are coming; a receiver must never close a channel it merely reads from.
func produce(values []int, out chan<- int) {
	defer close(out)
	for _, v := range values {
		out <- v // blocks here until main below receives it
	}
}

// collect drains in until it is closed, returning every value in the order
// it arrived. Because there is a single producer and a single consumer on
// an unbuffered channel, that order exactly matches the send order.
func collect(in <-chan int) []int {
	var values []int
	for v := range in { // range on a channel exits automatically on close
		values = append(values, v)
	}
	return values
}

func main() {
	numbers := make(chan int) // unbuffered: capacity 0

	go produce([]int{10, 20, 30}, numbers)

	for n := range numbers {
		fmt.Println("received", n)
	}
	fmt.Println("channel closed, loop exited cleanly")
}

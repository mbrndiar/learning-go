// Command 03_buffered_channels shows how a channel's capacity lets a sender
// get ahead of a receiver, up to that capacity, and how len/cap describe a
// channel's current buffered contents.
package main

import "fmt"

// fillThenDrain pushes every item into a channel of the given capacity and
// then drains and returns them in order. If len(items) is greater than
// capacity, the extra sends block until the goroutine below starts
// receiving, which is why draining happens concurrently with sending.
func fillThenDrain(capacity int, items []string) []string {
	queue := make(chan string, capacity)

	go func() {
		defer close(queue)
		for _, item := range items {
			queue <- item // blocks only once the buffer is full
		}
	}()

	drained := make([]string, 0, len(items))
	for item := range queue {
		drained = append(drained, item)
	}
	return drained
}

func main() {
	queue := make(chan string, 2) // buffer holds up to 2 values without a receiver

	queue <- "first"
	queue <- "second"
	fmt.Println("queued without blocking, len/cap:", len(queue), "/", cap(queue))

	close(queue) // safe: buffered values already in the channel still drain
	for item := range queue {
		fmt.Println("drained:", item)
	}

	fmt.Println("---")

	// Five items through a channel with room for only two: the producer
	// goroutine must wait for the consumer to keep up once the buffer
	// fills, but the order is still preserved because there is one
	// producer and one consumer.
	items := []string{"a", "b", "c", "d", "e"}
	fmt.Println("drained in order:", fillThenDrain(2, items))
}

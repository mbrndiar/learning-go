// Command 09_context_cancellation shows context.Context as the standard way
// to cancel work: propagate a Context down a call chain, and let every
// long-running goroutine watch ctx.Done() so it can stop promptly instead
// of leaking.
package main

import (
	"context"
	"fmt"
	"time"
)

// countTo sends increasing numbers on out until it reaches limit, or ctx is
// canceled, whichever happens first. It always closes out and always
// signals finished before returning, so a caller can prove the goroutine
// actually stopped instead of just hoping it did.
func countTo(ctx context.Context, limit int, out chan<- int, finished chan<- struct{}) {
	defer close(out)
	defer close(finished)

	for i := 1; i <= limit; i++ {
		select {
		case out <- i:
		case <-ctx.Done():
			return // stop immediately; do not leak this goroutine
		}
	}
}

// collectUntilCanceled cancels the count after receiving stopAfter values
// and reports how many it actually received before the goroutine confirmed
// it had exited.
func collectUntilCanceled(limit, stopAfter int) []int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // always call cancel to release the context's resources

	out := make(chan int)
	finished := make(chan struct{})
	go countTo(ctx, limit, out, finished)

	var received []int
	for v := range out {
		received = append(received, v)
		if len(received) == stopAfter {
			cancel() // ask countTo to stop
		}
	}

	<-finished // block until countTo has actually returned; proves no leak
	return received
}

func main() {
	// A short-lived deadline: the goroutine must give up work it cannot
	// finish in time rather than run forever.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	out := make(chan int)
	finished := make(chan struct{})
	go countTo(ctx, 1_000_000, out, finished)

	count := 0
	for range out {
		count++
	}
	<-finished
	fmt.Println("stopped after", count, "values; ctx.Err():", ctx.Err())

	fmt.Println("---")
	fmt.Println("received before manual cancel:", collectUntilCanceled(100, 3))
}

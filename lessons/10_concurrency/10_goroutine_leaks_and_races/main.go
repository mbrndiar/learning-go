// Command 10_goroutine_leaks_and_races ties together the previous lessons
// to show what a goroutine leak looks like, how to prevent one with
// context cancellation, and how to reason about data races. The race
// detector (`go test -race`) is the tool that catches races; this lesson
// shows code that passes it, plus commented anti-patterns that would not.
package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// BAD (do not copy): a goroutine leak.
//
//	func leaky() <-chan int {
//	    ch := make(chan int)
//	    go func() {
//	        ch <- slowComputation() // if nobody ever receives, this
//	    }()                         // goroutine blocks here forever
//	    return ch
//	}
//
// If the caller of leaky never reads from the returned channel (for
// example because it took an earlier error path and returned early), the
// goroutine above stays blocked on the send for the lifetime of the
// program. It cannot be garbage collected, because the runtime cannot
// prove no one will ever receive from ch. Multiply this by every request a
// server handles and memory/goroutine count grows without bound.

// BAD (do not copy): a data race.
//
//	func race() int {
//	    total := 0
//	    for i := 0; i < 100; i++ {
//	        go func() { total++ }() // unsynchronized write from many
//	    }                            // goroutines: a data race
//	    return total
//	}
//
// Multiple goroutines write total without a Mutex or atomic operation, so
// the result is not reliably determined and `go test -race` reports it as a
// race. Lesson 06 (Mutex) and lesson 07 (atomic) show the two standard fixes.

// slowTask simulates work that respects cancellation: it either finishes,
// or notices ctx is done and returns early. This is the fix for the leaky
// pattern above.
func slowTask(ctx context.Context, work <-chan struct{}) error {
	select {
	case <-work:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runWithLeakPrevention starts slowTask and cancels it if work never
// arrives, then waits for confirmation that the goroutine actually
// returned. Returning only after that confirmation, instead of after a
// fixed sleep, is what makes this function leak-free and deterministic.
func runWithLeakPrevention(ctx context.Context, work <-chan struct{}) error {
	done := make(chan error, 1)
	go func() {
		done <- slowTask(ctx, work)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return <-done // slowTask will also observe ctx.Done() and return promptly
	}
}

// goroutineCountStable polls runtime.NumGoroutine() until it stops climbing or
// a bound on attempts is reached. Goroutine counts can change for unrelated
// runtime work, so this is a best-effort observation for the demo, not a proof.
// The done channel in runWithLeakPrevention is the actual synchronization
// guarantee that the goroutine returned.
func goroutineCountStable() int {
	last := runtime.NumGoroutine()
	for i := 0; i < 50; i++ {
		runtime.Gosched()
		current := runtime.NumGoroutine()
		if current == last {
			return current
		}
		last = current
	}
	return last
}

func main() {
	before := goroutineCountStable()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	work := make(chan struct{}) // nobody ever sends on this: forces the timeout path
	err := runWithLeakPrevention(ctx, work)
	fmt.Println("runWithLeakPrevention error:", err)

	after := goroutineCountStable()
	fmt.Printf("goroutines before=%d after=%d (leak-free: no persistent growth)\n", before, after)
}

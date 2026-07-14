// Command 07_atomic_counters shows sync/atomic as a lock-free alternative
// to a Mutex for simple counters, and compares it with the Mutex-based
// approach from the previous lesson.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// atomicCounter uses atomic.Int64, a type whose methods perform the
// increment and read as a single indivisible hardware operation, so no
// explicit lock is needed.
type atomicCounter struct {
	count atomic.Int64
}

// Inc atomically increments the counter by one.
func (c *atomicCounter) Inc() {
	c.count.Add(1)
}

// Value atomically reads the current count.
func (c *atomicCounter) Value() int64 {
	return c.count.Load()
}

// incrementAtomically starts n goroutines that each call Inc once and waits
// for all of them, exactly like incrementConcurrently in the previous
// lesson, but backed by an atomic counter instead of a mutex-protected one.
func incrementAtomically(c *atomicCounter, n int) {
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}

	wg.Wait()
}

func main() {
	c := &atomicCounter{}
	const goroutines = 100

	incrementAtomically(c, goroutines)
	fmt.Println("final atomic count:", c.Value())

	// Use atomic.Int64/Uint64/Bool/Value for single independent values.
	// Reach for a Mutex as soon as you need to update more than one field
	// together, or run a multi-step check-then-act sequence, because
	// atomics only make a single operation indivisible, not a sequence of
	// them.
}

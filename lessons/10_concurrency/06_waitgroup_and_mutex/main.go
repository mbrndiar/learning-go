// Command 06_waitgroup_and_mutex shows sync.WaitGroup for waiting on a
// known number of goroutines, and sync.Mutex for protecting state that
// multiple goroutines read and write concurrently.
package main

import (
	"fmt"
	"sync"
)

// counter is a plain int guarded by a Mutex. Every access must go through
// Inc or Value; touching count directly from another goroutine would be a
// data race.
type counter struct {
	mu    sync.Mutex
	count int
}

// Inc safely increments the counter from any number of goroutines.
func (c *counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

// Value safely reads the current count.
func (c *counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// incrementConcurrently starts n goroutines that each call Inc once, and
// blocks until all of them are done using a WaitGroup. Without the
// WaitGroup, the function could return, and the caller could read
// c.Value(), before every goroutine finished incrementing.
func incrementConcurrently(c *counter, n int) {
	var wg sync.WaitGroup
	wg.Add(n) // record how many goroutines we are about to wait for

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done() // signal completion, even on an early return
			c.Inc()
		}()
	}

	wg.Wait() // blocks until every Done call has happened
}

func main() {
	c := &counter{}
	const goroutines = 100

	incrementConcurrently(c, goroutines)
	fmt.Println("final count:", c.Value())
}

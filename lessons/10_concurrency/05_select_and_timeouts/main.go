// Command 05_select_and_timeouts shows how select waits on multiple
// channel operations at once, how a default case makes a select
// non-blocking, and how to bound a wait with a timeout.
package main

import (
	"errors"
	"fmt"
	"time"
)

// firstOf returns whichever of a or b produces a value first. If both are
// ready in the same instant, select picks between them pseudo-randomly;
// callers must not depend on which one wins in that case.
func firstOf(a, b <-chan int) int {
	select {
	case v := <-a:
		return v
	case v := <-b:
		return v
	}
}

// tryReceiveNonBlocking reports whether a value was immediately available
// on ch, without ever blocking. The default case runs the instant neither
// channel case is ready.
func tryReceiveNonBlocking(ch <-chan int) (value int, ok bool) {
	select {
	case v := <-ch:
		return v, true
	default:
		return 0, false
	}
}

// ErrTimeout is returned by awaitWithTimeout when the wait deadline elapses
// before ch produces a value.
var ErrTimeout = errors.New("timed out waiting for value")

// awaitWithTimeout waits for a value on ch but gives up after timeout,
// returning ErrTimeout instead of blocking forever. This is the classic
// select-based timeout pattern; lesson 09 shows the equivalent, and
// generally preferred, context.WithTimeout approach.
func awaitWithTimeout(ch <-chan int, timeout time.Duration) (int, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop() // release the timer's resources promptly

	select {
	case v := <-ch:
		return v, nil
	case <-timer.C:
		return 0, ErrTimeout
	}
}

func main() {
	a := make(chan int, 1)
	b := make(chan int, 1)
	a <- 1
	fmt.Println("firstOf result:", firstOf(a, b))

	empty := make(chan int)
	value, ok := tryReceiveNonBlocking(empty)
	fmt.Printf("non-blocking receive on empty channel: value=%d ok=%v\n", value, ok)

	never := make(chan int)
	_, err := awaitWithTimeout(never, 20*time.Millisecond)
	fmt.Println("awaitWithTimeout on a channel nothing ever sends on:", err)
}

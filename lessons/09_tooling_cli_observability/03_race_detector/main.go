// Command racedetector demonstrates an intentional data race, a race-free
// fix, and how the race detector catches the difference. The two code
// paths compute the same thing (a counter incremented concurrently by
// several goroutines) with and without synchronization.
//
// Try it:
//
//	go run ./lessons/09_tooling_cli_observability/03_race_detector                  # safe path (default)
//	go run ./lessons/09_tooling_cli_observability/03_race_detector -mode=race       # unsynchronized, wrong result likely
//	go run -race ./lessons/09_tooling_cli_observability/03_race_detector -mode=race # race detector reports the data race
//
// This lesson previews goroutines, sync.WaitGroup, and sync.Mutex so you can
// use the race detector in a realistic example. Module 10 introduces those
// concurrency primitives from first principles; focus here on the difference
// between an ordinary run and a -race run.
package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

const (
	goroutines             = 50
	incrementsPerGoroutine = 1000
)

func main() {
	mode := flag.String("mode", "safe", "counter implementation: safe or race")
	flag.Parse()

	var got int
	switch *mode {
	case "race":
		got = racyCount()
	case "safe":
		got = safeCount()
	default:
		fmt.Fprintf(os.Stderr, "unknown -mode %q: want \"safe\" or \"race\"\n", *mode)
		os.Exit(2)
	}

	want := goroutines * incrementsPerGoroutine
	fmt.Printf("mode=%s counter=%d want=%d\n", *mode, got, want)
	if got != want {
		fmt.Println("mismatch: concurrent increments were lost without synchronization")
	}
}

// racyCount increments a plain int from many goroutines with no
// synchronization at all. Each "counter++" is really a read, an add, and a
// write; when two goroutines interleave those steps, an increment can be
// lost. Run this with `go run -race ... -mode=race` to see the race
// detector flag the concurrent unsynchronized access - it may also print a
// wrong (too low) total even without -race.
func racyCount() int {
	counter := 0
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				counter++ // unsynchronized read-modify-write: a data race
			}
		}()
	}
	wg.Wait()
	return counter
}

// safeCount increments the same kind of counter, but every access is
// guarded by a sync.Mutex, so the read-modify-write sequence cannot
// interleave across goroutines. This always produces the expected total
// and is race-detector clean.
func safeCount() int {
	counter := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return counter
}

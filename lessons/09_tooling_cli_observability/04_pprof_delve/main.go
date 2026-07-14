// Command pprofdelve demonstrates writing CPU and memory profiles with the
// standard library's runtime/pprof, as orientation before reading
// go tool pprof and Delve documentation. It performs a small CPU-bound
// workload (counting primes) and a small allocation-heavy workload, then
// writes both profiles to a temporary directory it prints on exit.
//
// Try it:
//
//	go run ./lessons/09_tooling_cli_observability/04_pprof_delve
//	go tool pprof -top <printed-path>/cpu.pprof
//	go tool pprof -top <printed-path>/heap.pprof
//
// See this module's README for delve (interactive debugger) orientation;
// it is not exercised at runtime because it is an external tool, not part
// of the standard library.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

func main() {
	dir, err := os.MkdirTemp("", "pprof-delve-lesson-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create profile directory:", err)
		os.Exit(1)
	}
	fmt.Println("profiles written to:", dir)

	if err := writeCPUProfile(filepath.Join(dir, "cpu.pprof")); err != nil {
		fmt.Fprintln(os.Stderr, "cpu profile:", err)
		os.Exit(1)
	}
	if err := writeHeapProfile(filepath.Join(dir, "heap.pprof")); err != nil {
		fmt.Fprintln(os.Stderr, "heap profile:", err)
		os.Exit(1)
	}

	fmt.Println("done - inspect the profiles with `go tool pprof -top <file>`")
}

// writeCPUProfile records CPU samples while running a deliberately
// CPU-bound workload, then writes them in the pprof binary format.
func writeCPUProfile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		return err
	}
	defer pprof.StopCPUProfile()

	_ = countPrimesBelow(200_000)
	return nil
}

// writeHeapProfile forces a garbage collection so the snapshot reflects
// live objects, then writes the current heap profile.
func writeHeapProfile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	allocateGarbage(200_000)
	runtime.GC()
	return pprof.WriteHeapProfile(f)
}

// countPrimesBelow is intentionally unoptimized (trial division) so it
// takes measurable CPU time for the profile above to sample.
func countPrimesBelow(limit int) int {
	count := 0
	for n := 2; n < limit; n++ {
		if isPrime(n) {
			count++
		}
	}
	return count
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// allocateGarbage builds and discards many small slices so the heap
// profile below has something to show.
func allocateGarbage(count int) {
	var sink [][]byte
	for i := 0; i < count; i++ {
		sink = append(sink, make([]byte, 64))
		if len(sink) > 1000 {
			sink = sink[:0] // let earlier allocations become garbage
		}
	}
}

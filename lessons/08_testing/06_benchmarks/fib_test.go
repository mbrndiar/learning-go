package benchmarks

import (
	"fmt"
	"testing"
)

// correctness tests run every time with `go test`; benchmarks only run with
// `go test -bench`, so keep both to avoid a fast benchmark suite silently
// drifting from correct behavior.
func TestFibImplementationsAgree(t *testing.T) {
	for n := 0; n <= 15; n++ {
		want := FibIterative(n)
		if got := FibRecursive(n); got != want {
			t.Errorf("FibRecursive(%d) = %d, want %d (from FibIterative)", n, got, want)
		}
	}
}

// BenchmarkFibRecursive measures the naive recursive implementation.
// Benchmark functions take *testing.B and run the code under test
// b.N times; the testing framework picks b.N to produce a stable
// measurement.
func BenchmarkFibRecursive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FibRecursive(20)
	}
}

// BenchmarkFibIterative measures the O(n) implementation for comparison.
// Run both with:
//
//	go test -bench . -benchmem ./lessons/08_testing/06_benchmarks
func BenchmarkFibIterative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FibIterative(20)
	}
}

// BenchmarkFibIterative_N30 shows a table of input sizes using sub-benchmarks,
// mirroring how subtests work for regular tests.
func BenchmarkFibIterativeSizes(b *testing.B) {
	for _, n := range []int{10, 20, 30, 40} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FibIterative(n)
			}
		})
	}
}

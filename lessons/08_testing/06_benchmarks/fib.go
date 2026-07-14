// Package benchmarks implements two Fibonacci algorithms with very
// different performance characteristics, used to demonstrate go test's
// benchmarking support.
package benchmarks

// FibRecursive computes the nth Fibonacci number using naive recursion.
// It recomputes the same sub-results many times, so it is exponential in n
// and deliberately slow for larger n - useful for showing a benchmark
// "before" number.
func FibRecursive(n int) int {
	if n < 2 {
		return n
	}
	return FibRecursive(n-1) + FibRecursive(n-2)
}

// FibIterative computes the nth Fibonacci number in O(n) time and O(1)
// space by keeping only the last two values.
func FibIterative(n int) int {
	if n < 2 {
		return n
	}
	prev, curr := 0, 1
	for i := 2; i <= n; i++ {
		prev, curr = curr, prev+curr
	}
	return curr
}

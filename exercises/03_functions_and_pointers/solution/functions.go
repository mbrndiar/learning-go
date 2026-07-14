// Package solution is the reference implementation for
// exercises/03_functions_and_pointers.
package solution

import "errors"

// ErrDivideByZero is returned by Divide when b is zero.
var ErrDivideByZero = errors.New("division by zero")

// ErrNoValues is returned by MinMax when called with no arguments.
var ErrNoValues = errors.New("no values provided")

// Divide returns a/b. If b is zero, it returns ErrDivideByZero instead of
// dividing.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}

// MinMax returns the minimum and maximum of nums. If nums is empty, it
// returns ErrNoValues.
func MinMax(nums ...int) (min int, max int, err error) {
	if len(nums) == 0 {
		return 0, 0, ErrNoValues
	}
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return min, max, nil
}

// Sum returns the sum of nums. Sum() with no arguments returns 0.
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Counter returns a function that, on each call, returns the next count
// starting at 1. Each call to Counter produces an independent counter
// because count is captured by the closure, not shared globally.
func Counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// Accumulator returns a function that adds its argument to a running total
// seeded by start, and returns the new total. Each call to Accumulator
// produces an independent running total.
func Accumulator(start int) func(int) int {
	total := start
	return func(n int) int {
		total += n
		return total
	}
}

// Increment adds 1 to the int pointed to by n. It does nothing if n is nil.
func Increment(n *int) {
	if n == nil {
		return
	}
	*n++
}

// SwapInts swaps the values pointed to by a and b. It does nothing if
// either pointer is nil.
func SwapInts(a, b *int) {
	if a == nil || b == nil {
		return
	}
	*a, *b = *b, *a
}

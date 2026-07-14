// Package functions contains starter exercises covering multiple return
// values, variadic parameters, closures, and pointers. Replace every
// panic("TODO: ...") with a working implementation.
package functions

import "errors"

// ErrDivideByZero is returned by Divide when b is zero.
var ErrDivideByZero = errors.New("division by zero")

// ErrNoValues is returned by MinMax when called with no arguments.
var ErrNoValues = errors.New("no values provided")

// Divide returns a/b. If b is zero, it returns ErrDivideByZero instead of
// dividing.
func Divide(a, b float64) (float64, error) {
	panic("TODO: implement Divide")
}

// MinMax returns the minimum and maximum of nums. If nums is empty, it
// returns ErrNoValues.
func MinMax(nums ...int) (min int, max int, err error) {
	panic("TODO: implement MinMax")
}

// Sum returns the sum of nums. Sum() with no arguments returns 0.
func Sum(nums ...int) int {
	panic("TODO: implement Sum")
}

// Counter returns a function that, on each call, returns the next count
// starting at 1. Each call to Counter must produce an independent counter.
func Counter() func() int {
	panic("TODO: implement Counter")
}

// Accumulator returns a function that adds its argument to a running total
// seeded by start, and returns the new total. Each call to Accumulator must
// produce an independent running total.
func Accumulator(start int) func(int) int {
	panic("TODO: implement Accumulator")
}

// Increment adds 1 to the int pointed to by n. It does nothing if n is nil.
func Increment(n *int) {
	panic("TODO: implement Increment")
}

// SwapInts swaps the values pointed to by a and b. It does nothing if
// either pointer is nil.
func SwapInts(a, b *int) {
	panic("TODO: implement SwapInts")
}

// Package tabledriven implements a few small calculator functions used to
// demonstrate table-driven tests: a slice of input/output cases driven
// through one shared test body.
package tabledriven

import "fmt"

// Add returns a + b.
func Add(a, b int) int {
	return a + b
}

// Divide returns a / b as a float64. It returns an error instead of letting
// the division reach the +Inf/NaN floating-point result, so callers cannot
// silently propagate a division by zero.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("divide %g by zero: not allowed", a)
	}
	return a / b, nil
}

// Abs returns the absolute value of n.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

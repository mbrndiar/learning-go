// This lesson covers function declarations: parameters, multiple return
// values, named returns, and why Go encourages returning an error instead
// of throwing an exception.
package main

import (
	"errors"
	"fmt"
)

// add takes two named parameters of the same type. When consecutive
// parameters share a type, Go lets you write the type once at the end.
func add(a, b int) int {
	return a + b
}

// divide returns two values: the result and an error. Returning an error
// as an ordinary value (instead of using exceptions) forces every caller
// to decide, at the call site, how to handle failure. This is one of
// Go's core idioms.
func divide(numerator, denominator float64) (float64, error) {
	if denominator == 0 {
		return 0, errors.New("divide: denominator must not be zero")
	}
	return numerator / denominator, nil
}

// minMax demonstrates named return values. "smallest" and "largest" are
// declared in the function signature and start at their zero values; a
// bare "return" sends back whatever they currently hold. Named returns
// are most useful as documentation for what each return value means, and
// in short functions like this one.
func minMax(numbers []int) (smallest, largest int) {
	if len(numbers) == 0 {
		return // returns the zero values 0, 0
	}
	smallest, largest = numbers[0], numbers[0]
	for _, n := range numbers[1:] {
		if n < smallest {
			smallest = n
		}
		if n > largest {
			largest = n
		}
	}
	return smallest, largest
}

func main() {
	fmt.Println("--- multiple parameters, single return ---")
	fmt.Printf("add(2, 3) = %d\n", add(2, 3))

	fmt.Println("--- multiple return values, including an error ---")
	for _, denominator := range []float64{2, 0} {
		result, err := divide(10, denominator)
		if err != nil {
			// Idiomatic Go checks errors immediately after the call that
			// might produce them, rather than deferring the check.
			fmt.Printf("divide(10, %v) failed: %v\n", denominator, err)
			continue
		}
		fmt.Printf("divide(10, %v) = %v\n", denominator, result)
	}

	fmt.Println("--- ignoring a return value with the blank identifier ---")
	// If you only need one of several return values, assign the others
	// to "_" so the compiler does not complain about unused variables.
	result, _ := divide(9, 3)
	fmt.Printf("result only: %v\n", result)

	fmt.Println("--- named returns document intent ---")
	lo, hi := minMax([]int{4, 1, 7, 3})
	fmt.Printf("smallest=%d largest=%d\n", lo, hi)
	loEmpty, hiEmpty := minMax(nil)
	fmt.Printf("empty slice -> smallest=%d largest=%d (zero values)\n", loEmpty, hiEmpty)
}

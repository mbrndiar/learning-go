// This lesson covers variadic functions, functions as values, and
// closures that capture variables from their surrounding scope.
package main

import "fmt"

// sum is variadic: callers may pass any number of int arguments,
// including zero. Inside the function, "numbers" is an ordinary []int.
func sum(numbers ...int) int {
	total := 0
	for _, n := range numbers {
		total += n
	}
	return total
}

// describe accepts a function value as a parameter. Functions are
// first-class values in Go: they can be stored in variables, passed as
// arguments, and returned from other functions.
func describe(operation string, calc func(int, int) int, a, b int) {
	fmt.Printf("%s(%d, %d) = %d\n", operation, a, b, calc(a, b))
}

// makeCounter returns a closure: an anonymous function that captures the
// variable "count" from its enclosing scope. Each call to makeCounter
// creates a brand new, independent "count" variable, so separate counters
// do not interfere with each other.
func makeCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	fmt.Println("--- variadic functions ---")
	fmt.Printf("sum() = %d\n", sum())
	fmt.Printf("sum(1, 2, 3) = %d\n", sum(1, 2, 3))

	// A slice can be "spread" into a variadic parameter using "...".
	values := []int{10, 20, 30, 40}
	fmt.Printf("sum(values...) = %d\n", sum(values...))

	fmt.Println("--- functions as values ---")
	add := func(a, b int) int { return a + b }
	multiply := func(a, b int) int { return a * b }
	describe("add", add, 3, 4)
	describe("multiply", multiply, 3, 4)

	// Functions can be stored in a slice or map just like any other value.
	operations := map[string]func(int, int) int{
		"add":      add,
		"multiply": multiply,
	}
	for _, name := range []string{"add", "multiply"} { // fixed order for deterministic output
		fmt.Printf("operations[%q](5, 6) = %d\n", name, operations[name](5, 6))
	}

	fmt.Println("--- closures capture variables, not values ---")
	counterA := makeCounter()
	counterB := makeCounter()
	fmt.Printf("counterA: %d, %d, %d\n", counterA(), counterA(), counterA())
	fmt.Printf("counterB: %d (independent from counterA)\n", counterB())

	fmt.Println("--- common mistake: capturing a loop variable by reference ---")
	// Since Go 1.22, each iteration of a for loop creates a NEW copy of
	// the loop variable, so closures created in a loop correctly capture
	// the value from their own iteration. (Older Go versions shared one
	// variable across iterations, which was a frequent bug source.)
	var printers []func()
	for i := 0; i < 3; i++ {
		printers = append(printers, func() {
			fmt.Printf("captured i=%d\n", i)
		})
	}
	for _, print := range printers {
		print()
	}
}

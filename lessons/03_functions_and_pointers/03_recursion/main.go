// This lesson covers recursion: functions that call themselves to solve
// a problem in terms of smaller versions of itself, plus the trade-offs
// compared to an iterative solution.
package main

import "fmt"

// factorial computes n! recursively. Every recursive function needs a
// "base case" that stops the recursion; here it is n <= 1. Without a
// base case, the function would call itself forever until the program
// runs out of stack space.
func factorial(n int) int {
	if n <= 1 {
		return 1 // base case
	}
	return n * factorial(n-1) // recursive case: solve a smaller problem
}

// fibonacci computes the nth Fibonacci number using naive recursion. It
// is intentionally simple to read, but recomputes the same values many
// times (fibonacci(3) is computed once for every call to fibonacci(5),
// for example), so it is exponentially slow for large n.
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// fibonacciIterative computes the same value with a loop, doing a fixed
// amount of work per step. For performance-sensitive code, an iterative
// version (or a recursive version with memoization) is usually
// preferable to naive recursion.
func fibonacciIterative(n int) int {
	if n <= 1 {
		return n
	}
	previous, current := 0, 1
	for i := 2; i <= n; i++ {
		previous, current = current, previous+current
	}
	return current
}

// countDown demonstrates recursion used for its control-flow shape rather
// than for a mathematical formula.
func countDown(n int) {
	if n <= 0 {
		fmt.Println("liftoff!")
		return
	}
	fmt.Println(n)
	countDown(n - 1)
}

func main() {
	fmt.Println("--- factorial ---")
	for _, n := range []int{0, 1, 5} {
		fmt.Printf("factorial(%d) = %d\n", n, factorial(n))
	}

	fmt.Println("--- naive vs iterative fibonacci ---")
	for _, n := range []int{0, 1, 10} {
		fmt.Printf("fibonacci(%d) = %d, fibonacciIterative(%d) = %d\n",
			n, fibonacci(n), n, fibonacciIterative(n))
	}

	fmt.Println("--- recursion for control flow ---")
	countDown(3)

	// Every recursive call adds a new frame to the call stack. A
	// recursive function with no reachable base case, or one that is
	// simply called with values that never reach it, causes a stack
	// overflow at runtime. Always make sure the recursive argument moves
	// toward the base case on every call.
}

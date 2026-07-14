// Command 01_goroutines_basics shows how the go keyword starts a new
// goroutine and why you need a synchronization point before you can trust
// that goroutine's work is finished.
package main

import (
	"fmt"
	"sort"
)

// square is the small unit of "work" each goroutine performs. Keeping it as
// a plain function makes it easy to call directly in tests, with no
// concurrency involved.
func square(n int) int {
	return n * n
}

// squareAll launches one goroutine per input number and collects every
// result through a channel. The channel is buffered to the exact number of
// goroutines, so every send succeeds immediately and no goroutine blocks
// waiting for a receiver. Channels are covered in depth in the next two
// lessons; here it is used only as a safe way to learn when goroutines
// finish.
func squareAll(numbers []int) []int {
	results := make(chan int, len(numbers))

	for _, n := range numbers {
		// Go 1.22+ gives each loop iteration its own copy of n, so it is
		// safe to reference the loop variable directly inside the closure.
		go func() {
			results <- square(n)
		}()
	}

	collected := make([]int, 0, len(numbers))
	for range numbers {
		collected = append(collected, <-results)
	}

	// Goroutines finish in an unpredictable order, so we sort before
	// printing to keep the program's output deterministic.
	sort.Ints(collected)
	return collected
}

func main() {
	numbers := []int{5, 3, 4, 1, 2}

	fmt.Println("input:", numbers)
	fmt.Println("squares:", squareAll(numbers))

	// This goroutine is intentionally NOT waited for. main can return
	// before it ever runs. Go does not print anything when that happens;
	// it simply abandons the goroutine. This is the seed of a goroutine
	// leak, which lesson 10 covers in detail.
	go fmt.Println("this line may never print")
}

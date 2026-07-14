// This lesson covers every shape of Go's single loop keyword, "for", and
// how "range" iterates over strings, slices, and maps.
package main

import "fmt"

func main() {
	fmt.Println("--- classic three-part for ---")
	// init; condition; post — the same shape as C's for loop.
	for i := 0; i < 3; i++ {
		fmt.Printf("classic loop i=%d\n", i)
	}

	fmt.Println("--- condition-only for (while loop) ---")
	// Dropping the init and post clauses gives Go's equivalent of a
	// "while" loop from other languages.
	countdown := 3
	for countdown > 0 {
		fmt.Printf("countdown=%d\n", countdown)
		countdown--
	}

	fmt.Println("--- infinite for with break ---")
	// Dropping all three clauses gives an infinite loop; a "break" (often
	// guarded by an if) is required to stop it. This shape is common when
	// the exit condition is easier to express in the middle of the body.
	attempts := 0
	for {
		attempts++
		if attempts >= 3 {
			fmt.Printf("stopping after %d attempts\n", attempts)
			break
		}
	}

	fmt.Println("--- range over a slice ---")
	// range gives (index, value) pairs. "value" is a COPY of the element,
	// so mutating it does not change the original slice.
	fruits := []string{"apple", "banana", "cherry"}
	for index, fruit := range fruits {
		fmt.Printf("fruits[%d] = %s\n", index, fruit)
	}

	fmt.Println("--- range value is a copy: a common mutation mistake ---")
	numbers := []int{1, 2, 3}
	for _, n := range numbers {
		n *= 10 // this only changes the local copy "n"
		_ = n
	}
	fmt.Printf("numbers unchanged by the loop above: %v\n", numbers)
	// To mutate the underlying slice, index into it directly instead.
	for i := range numbers {
		numbers[i] *= 10
	}
	fmt.Printf("numbers after indexed mutation: %v\n", numbers)

	fmt.Println("--- range over a string yields runes, not bytes ---")
	for index, char := range "Go!" {
		fmt.Printf("index=%d rune=%q\n", index, char)
	}

	fmt.Println("--- range with only the value, or only the index ---")
	for _, fruit := range fruits {
		fmt.Printf("value only: %s\n", fruit)
	}
	for index := range fruits {
		fmt.Printf("index only: %d\n", index)
	}

	fmt.Println("--- range over an integer (Go 1.22+) ---")
	// "range n" for an integer n iterates n times, yielding 0..n-1. It is
	// a concise replacement for "for i := 0; i < n; i++" when you do not
	// need a custom step.
	for i := range 3 {
		fmt.Printf("range over int: i=%d\n", i)
	}
}

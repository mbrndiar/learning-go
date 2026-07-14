// This lesson covers arrays and slices: their differences, how length
// and capacity work, and how append grows a slice.
package main

import "fmt"

func main() {
	fmt.Println("--- arrays have a fixed size that is part of their type ---")
	// [3]int and [4]int are different types. An array's size cannot
	// change; this is rarely used directly in idiomatic Go but underlies
	// how slices work.
	var fixed [3]int
	fmt.Printf("zero-value array: %v\n", fixed)
	fixed[0], fixed[1], fixed[2] = 10, 20, 30
	fmt.Printf("filled array: %v (len=%d)\n", fixed, len(fixed))

	fmt.Println("--- arrays copy on assignment; slices share data ---")
	arrayCopy := fixed
	arrayCopy[0] = 999
	fmt.Printf("original array: %v, copy: %v (independent)\n", fixed, arrayCopy)

	fmt.Println("--- slices: a view over an underlying array ---")
	// A slice literal creates both a backing array and a slice header
	// that describes (pointer, length, capacity) for it.
	numbers := []int{1, 2, 3}
	fmt.Printf("numbers=%v len=%d cap=%d\n", numbers, len(numbers), cap(numbers))

	fmt.Println("--- append grows the slice, reallocating when needed ---")
	// While capacity allows it, append writes into the existing backing
	// array. Once length would exceed capacity, append allocates a NEW,
	// larger backing array and copies the old elements into it. Capacity
	// growth is an implementation detail, but the pattern (roughly
	// doubling for small slices) is worth recognizing.
	grown := make([]int, 0, 2) // length 0, capacity 2
	for i := 1; i <= 5; i++ {
		previousCap := cap(grown)
		grown = append(grown, i)
		if cap(grown) != previousCap {
			fmt.Printf("append %d: len=%d cap=%d (reallocated: %d -> %d)\n",
				i, len(grown), cap(grown), previousCap, cap(grown))
		} else {
			fmt.Printf("append %d: len=%d cap=%d (reused existing array)\n",
				i, len(grown), cap(grown))
		}
	}

	fmt.Println("--- slicing an existing slice ---")
	// low:high selects a half-open range: elements from index low up to,
	// but not including, index high. The result shares the same backing
	// array as the original (see the next lesson for the implications).
	letters := []string{"a", "b", "c", "d", "e"}
	middle := letters[1:3]
	fmt.Printf("letters=%v, letters[1:3]=%v\n", letters, middle)

	fmt.Println("--- make creates a slice with an explicit length ---")
	buffer := make([]int, 3) // length 3, all elements zero-valued
	fmt.Printf("make([]int, 3) = %v\n", buffer)

	fmt.Println("--- a nil slice behaves like an empty slice ---")
	var empty []int
	fmt.Printf("nil slice: %v, len=%d, is nil=%t\n", empty, len(empty), empty == nil)
	// append works fine on a nil slice; it allocates a backing array on
	// first use. This is why "var items []T" is a perfectly idiomatic way
	// to start an empty collection.
	empty = append(empty, 1)
	fmt.Printf("after append to nil slice: %v\n", empty)
}

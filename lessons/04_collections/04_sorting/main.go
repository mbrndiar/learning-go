// This lesson covers sorting slices and maps using the standard library:
// slices.Sort for basic ordering, slices.SortFunc for custom rules, and
// combining the slices and maps packages to get deterministic output
// from a map.
package main

import (
	"fmt"
	"maps"
	"slices"
)

type player struct {
	name  string
	score int
}

func main() {
	fmt.Println("--- slices.Sort for ordered basic types ---")
	// slices.Sort works directly on any slice of an "ordered" type
	// (numbers and strings), sorting it in place, ascending.
	numbers := []int{5, 2, 4, 1, 3}
	slices.Sort(numbers)
	fmt.Printf("sorted numbers: %v\n", numbers)

	words := []string{"banana", "apple", "cherry"}
	slices.Sort(words)
	fmt.Printf("sorted words: %v\n", words)

	fmt.Println("--- slices.Sort does not preserve the original slice ---")
	// Sort mutates its argument in place; if you need the original order
	// too, clone the slice first.
	original := []int{3, 1, 2}
	sortedCopy := slices.Clone(original)
	slices.Sort(sortedCopy)
	fmt.Printf("original=%v sortedCopy=%v\n", original, sortedCopy)

	fmt.Println("--- slices.SortFunc for custom orderings ---")
	// SortFunc takes a comparison function returning a negative number,
	// zero, or a positive number, exactly like strings.Compare. This is
	// required for types with no natural ordering, like a struct, and
	// also for reversing the natural order.
	players := []player{
		{name: "Ada", score: 88},
		{name: "Grace", score: 95},
		{name: "Alan", score: 95},
	}
	slices.SortFunc(players, func(a, b player) int {
		if a.score != b.score {
			return b.score - a.score // higher score first (descending)
		}
		// Tie-break by name so the order is fully deterministic even when
		// scores match, which also makes the sort behave predictably.
		if a.name < b.name {
			return -1
		}
		if a.name > b.name {
			return 1
		}
		return 0
	})
	for _, p := range players {
		fmt.Printf("%s: %d\n", p.name, p.score)
	}

	fmt.Println("--- reversing a sorted slice ---")
	ascending := []int{1, 2, 3}
	slices.Reverse(ascending)
	fmt.Printf("reversed: %v\n", ascending)

	fmt.Println("--- checking whether a slice is already sorted ---")
	fmt.Printf("slices.IsSorted(numbers) = %t\n", slices.IsSorted(numbers))

	fmt.Println("--- sorting map keys for deterministic iteration ---")
	// Map iteration order is randomized (see the maps lesson), so when
	// output needs to be deterministic (logs, tests, snapshots), collect
	// and sort the keys explicitly.
	inventory := map[string]int{"bolts": 120, "screws": 80, "nails": 200}
	sortedNames := slices.Sorted(maps.Keys(inventory))
	for _, name := range sortedNames {
		fmt.Printf("%s: %d\n", name, inventory[name])
	}

	fmt.Println("--- sorting by value instead of key ---")
	type stock struct {
		name  string
		count int
	}
	stocks := make([]stock, 0, len(inventory))
	for name, count := range inventory {
		stocks = append(stocks, stock{name: name, count: count})
	}
	slices.SortFunc(stocks, func(a, b stock) int {
		return b.count - a.count // largest count first
	})
	for _, s := range stocks {
		fmt.Printf("%s has %d in stock\n", s.name, s.count)
	}
}

// This lesson covers break, continue, labeled loops, and the mutation
// cautions that come from combining loops with slices of structs or
// nested loops.
package main

import "fmt"

// Structs are introduced in module 5. Here task is only a compact way to keep
// a name and completion flag together while demonstrating range semantics.
type task struct {
	name string
	done bool
}

func main() {
	fmt.Println("--- continue skips to the next iteration ---")
	for i := 1; i <= 6; i++ {
		if i%2 != 0 {
			continue // skip odd numbers; nothing below runs for them
		}
		fmt.Printf("even number: %d\n", i)
	}

	fmt.Println("--- break stops the innermost loop only ---")
	grid := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	target := 5
	found := false
	for row := range grid {
		for col := range grid[row] {
			if grid[row][col] == target {
				fmt.Printf("found %d at row=%d col=%d\n", target, row, col)
				found = true
				break // this only exits the inner (col) loop
			}
		}
		if found {
			break // the outer loop still needs its own break
		}
	}

	fmt.Println("--- labeled break exits an outer loop directly ---")
	// A label placed before a loop lets break/continue target that exact
	// loop, even from inside a deeply nested loop. This avoids the extra
	// "found" flag used above.
search:
	for row := range grid {
		for col := range grid[row] {
			if grid[row][col] == target {
				fmt.Printf("labeled break found %d at row=%d col=%d\n", target, row, col)
				break search
			}
		}
	}

	fmt.Println("--- labeled continue skips to the next outer iteration ---")
outer:
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			if col == 1 {
				continue outer // move to the next row immediately
			}
			fmt.Printf("row=%d col=%d\n", row, col)
		}
	}

	fmt.Println("--- mutation caution: range gives copies of struct elements ---")
	tasks := []task{{name: "write", done: false}, {name: "review", done: false}}
	for _, t := range tasks {
		t.done = true // mutates only the local copy "t", not the slice
	}
	fmt.Printf("tasks unaffected by copy mutation: %+v\n", tasks)

	// To mutate elements in place, index into the slice, or range over
	// pointers if the slice holds pointers instead of values.
	for i := range tasks {
		tasks[i].done = true
	}
	fmt.Printf("tasks after indexed mutation: %+v\n", tasks)

	fmt.Println("--- mutation caution: appending to a slice while ranging over it ---")
	// range evaluates the slice's length once, before the loop starts, so
	// appending to the SAME slice variable inside the loop does not make
	// the loop visit the newly appended elements. This avoids infinite
	// growth but can still surprise you if you expect the new elements to
	// be visited.
	numbers := []int{1, 2, 3}
	for _, n := range numbers {
		if n == 3 {
			numbers = append(numbers, 4)
		}
	}
	fmt.Printf("numbers after append during range: %v (loop never saw the 4)\n", numbers)
}

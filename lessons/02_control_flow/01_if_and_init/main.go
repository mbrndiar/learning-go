// This lesson covers Go's if statement, including the optional
// initialization clause that scopes a variable to the if/else chain.
// That pattern is idiomatic Go and shows up constantly with functions
// that return a value and an error together.
package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("--- plain if/else ---")
	temperature := 18
	if temperature > 25 {
		fmt.Println("it's warm")
	} else if temperature > 15 {
		fmt.Println("it's mild")
	} else {
		fmt.Println("it's cold")
	}

	fmt.Println("--- if with an initialization statement ---")
	// The syntax "if init; condition { }" runs init first, then checks
	// condition. Any variable declared in init is scoped ONLY to the if,
	// the else-if branches, and the else branch: it does not leak into
	// the surrounding function. This keeps helper variables (like an
	// error) from cluttering the rest of the function.
	inputs := []string{"42", "not-a-number", "-7"}
	for _, raw := range inputs {
		if value, err := strconv.Atoi(raw); err != nil {
			fmt.Printf("%q is not a valid integer: %v\n", raw, err)
		} else if value < 0 {
			fmt.Printf("%q parsed as %d, but negative numbers are rejected\n", raw, value)
		} else {
			fmt.Printf("%q parsed as %d\n", raw, value)
		}
		// Note: `value` and `err` from the if statement above are not
		// visible here; they only existed inside the if/else chain.
	}

	fmt.Println("--- common mistake: declaring instead of reusing ---")
	// Using := inside an if's init clause always creates NEW variables
	// scoped to that if. A frequent bug is expecting it to update an
	// outer variable of the same name. Compare the intentional shadow
	// below with reusing an existing variable via plain assignment.
	total := 0
	if total := total + 100; total > 50 { // this "total" shadows the outer one
		fmt.Printf("shadowed total inside if: %d\n", total)
	}
	fmt.Printf("outer total is unchanged: %d\n", total)
}

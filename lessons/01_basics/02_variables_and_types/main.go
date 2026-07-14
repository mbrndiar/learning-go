// This lesson shows how Go declares variables and constants, what "zero
// values" are, how Go's static type system works, and how to convert
// between numeric types explicitly.
package main

import "fmt"

// Package-level constants are computed at compile time. Grouping related
// constants in a block avoids repeating the keyword.
const (
	maxRetries = 3
	appName    = "lesson-runner"
)

func main() {
	// var declares a variable with an explicit type. Because it has no
	// initial value, Go gives it the "zero value" for that type: 0 for
	// numbers, "" for strings, false for booleans, and nil for pointers,
	// slices, maps, channels, functions, and interfaces. Zero values mean
	// a variable is always usable immediately, never "uninitialized
	// garbage" like in some other languages.
	var count int
	var label string
	var ready bool
	fmt.Printf("zero values -> count=%d label=%q ready=%t\n", count, label, ready)

	// The := short variable declaration infers the type from the value on
	// the right. It can only be used inside functions, not at package
	// level, and only when at least one variable on the left is new.
	age := 30
	price := 19.99
	initial := 'G' // a rune literal; its type is rune (an alias for int32)
	fmt.Printf("inferred types -> age=%T price=%T initial=%T\n", age, price, initial)

	// Go is statically typed: once a variable has a type, it keeps it.
	// Mixing types in an expression without converting is a compile
	// error, which catches many bugs before the program ever runs.
	var whole int = 10
	var fraction float64 = 3.0
	// The next line would not compile because int and float64 are
	// different types: result := whole / fraction
	result := float64(whole) / fraction // explicit conversion required
	fmt.Printf("10 / 3.0 as float64 = %.4f\n", result)

	// Converting float to int truncates toward zero; it does not round.
	// Note: this only applies to converting a non-constant value. Go
	// rejects int(7.9) at compile time because a constant conversion must
	// be exactly representable; storing 7.9 in a float64 variable first
	// makes the conversion a runtime operation instead.
	positive := 7.9
	negative := -7.9
	truncated := int(positive)
	negativeTruncated := int(negative)
	fmt.Printf("int(7.9)=%d int(-7.9)=%d\n", truncated, negativeTruncated)

	// Numeric conversions can also lose information silently if the
	// target type is too small. This is a common source of bugs when
	// working with byte-sized data.
	var big int = 300
	var small byte = byte(big) // byte is uint8: 0-255, so this wraps around
	fmt.Printf("byte(300) = %d (wrapped, because byte only holds 0-255)\n", small)

	// Constants can be untyped until used, which lets the same constant
	// work as an int, float64, or other numeric type depending on
	// context. Here maxRetries (an untyped constant) is used as an int.
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("%s: attempt %d of %d\n", appName, attempt, maxRetries)
	}
}

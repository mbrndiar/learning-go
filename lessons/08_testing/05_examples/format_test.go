package examples

import "fmt"

// ExampleTitleCase demonstrates TitleCase. go test runs this function,
// captures everything it prints to stdout, and compares it against the
// text in the "Output:" comment. A mismatch fails the test. go doc also
// shows this example next to TitleCase's documentation.
func ExampleTitleCase() {
	fmt.Println(TitleCase("hello GO world"))
	// Output: Hello Go World
}

// ExampleFormatCents_negative shows the "_name" suffix convention: it
// documents FormatCents for a specific scenario (a negative amount) while
// still being associated with the FormatCents symbol in go doc output.
func ExampleFormatCents_negative() {
	fmt.Println(FormatCents(-105))
	// Output: -$1.05
}

func ExampleFormatCents() {
	fmt.Println(FormatCents(1099))
	// Output: $10.99
}

// Example (with no suffix and no associated symbol) documents the package
// as a whole and can combine multiple calls in one narrative.
func Example() {
	fmt.Println(TitleCase("go rocks"))
	fmt.Println(FormatCents(50))
	// Unordered output:
	// Go Rocks
	// $0.50
}

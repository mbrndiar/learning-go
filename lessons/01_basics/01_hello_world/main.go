// Package main is the entry point Go looks for when you run a program.
// Every standalone Go program needs exactly one package main and exactly
// one func main in that package; that function is where execution begins.
package main

import "fmt"

func main() {
	// fmt.Println writes its arguments to standard output, separated by
	// spaces, followed by a newline. It is the simplest way to see what
	// your program is doing while you learn.
	fmt.Println("Hello, Go!")

	// fmt.Printf lets you format values inside a string using verbs like
	// %s (string), %d (integer), %v (default format for any value), and
	// %T (the type of a value). The \n at the end is not automatic here;
	// you must add it yourself.
	name := "Gopher"
	fmt.Printf("Hello, %s! You are running Go.\n", name)

	// fmt.Sprintf builds a formatted string instead of printing it, which
	// is useful when you need the text for something other than the
	// screen, such as an error message or a log line.
	greeting := fmt.Sprintf("Welcome, %s.", name)
	fmt.Println(greeting)

	// %v prints the "default" representation of a value, and %T prints
	// its type. These two verbs are invaluable when you are not sure what
	// a value looks like or what type Go inferred for it.
	fmt.Printf("value=%v type=%T\n", 42, 42)
}

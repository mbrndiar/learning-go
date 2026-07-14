// Package main introduces structs: custom types that group related fields.
package main

import "fmt"

// Point groups two related fields. A struct type is a blueprint; it does not
// exist in memory until you create a value of it.
type Point struct {
	X int
	Y int
}

// Rectangle demonstrates a nested (composed) struct. TopLeft and BottomRight
// are themselves Point values, not pointers, so a Rectangle owns its corners.
type Rectangle struct {
	TopLeft     Point
	BottomRight Point
	Label       string
}

func main() {
	// Zero value: a struct's fields start at their own zero values. There is
	// no "null struct" in Go; a Point always has an X and a Y.
	var origin Point
	fmt.Println("zero value:", origin)

	// Keyed struct literal: the recommended form. Field order does not matter
	// and adding a field later will not silently break this code.
	a := Point{X: 1, Y: 2}

	// Positional struct literal: every field must be given, in declaration
	// order. Fragile if the struct's fields are reordered later, so prefer
	// keyed literals outside of very small, stable types.
	b := Point{3, 4}

	fmt.Println("a:", a, "b:", b)

	// Struct values are copied on assignment, on function return, and when
	// passed as arguments. Mutating a copy never affects the original.
	c := a
	c.X = 100
	fmt.Println("a unchanged:", a, "c changed:", c)

	// Structs are comparable with == when every field is comparable. Slices,
	// maps, and functions are not comparable, so a struct containing one of
	// those cannot be compared with ==.
	fmt.Println("a == b:", a == b)
	fmt.Println("a == Point{X: 1, Y: 2}:", a == Point{X: 1, Y: 2})

	// Nested struct literals: keyed literals nest naturally.
	box := Rectangle{
		TopLeft:     Point{X: 0, Y: 10},
		BottomRight: Point{X: 10, Y: 0},
		Label:       "box",
	}
	fmt.Printf("box: %+v\n", box)

	// Anonymous structs are useful for a one-off value that does not deserve
	// a named type, such as a small table of test-like data or a scratch
	// value inside a function.
	measurement := struct {
		Width, Height int
	}{Width: 3, Height: 4}
	fmt.Println("anonymous struct:", measurement)

	// The address-of operator works on struct literals too, giving a pointer
	// without an intermediate variable. This is common when a constructor
	// function returns *Point.
	origin2 := &Point{}
	fmt.Println("pointer to zero-value Point:", *origin2)
}

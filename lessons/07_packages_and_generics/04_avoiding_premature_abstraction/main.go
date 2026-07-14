// Package main is not about a new language feature. It is about judgment:
// when generics (or any abstraction) earn their complexity, and when a
// second concrete copy of a function is the more readable choice.
package main

import "fmt"

// --- Stage 1: two concrete functions, duplicated on purpose ---
//
// sumInts and sumFloats look almost identical. Duplicating a five-line
// function twice is completely fine; it costs far less than an abstraction
// that turns out to be wrong. This is "the rule of three" in practice: wait
// for a third real need before generalizing, because two examples rarely
// reveal the actual shape a good abstraction should take.

func sumInts(values []int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}

func sumFloats(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

// --- Stage 2: a third need arrives, and the shared shape is now obvious ---
//
// A Number constraint plus one generic Sum removes the duplication cleanly
// because all three cases (int, float64, and now int64) truly only differ
// by element type; nothing else about the loop changes.

type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](values []T) T {
	var total T
	for _, v := range values {
		total += v
	}
	return total
}

// --- Stage 3: a cautionary, over-generalized version ---
//
// aggregate below is what premature abstraction tends to look like: a single
// function that tries to cover sum, count, and "join as string" through a
// combinator argument. It is not wrong, but every caller now pays the price
// of reading a generic signature and a combine function just to do what
// Sum(values) already said directly. Prefer the narrowest tool (Sum) unless
// a real, current caller needs the extra flexibility that aggregate offers.
func aggregate[T any, Acc any](values []T, initial Acc, combine func(Acc, T) Acc) Acc {
	acc := initial
	for _, v := range values {
		acc = combine(acc, v)
	}
	return acc
}

func main() {
	ints := []int{1, 2, 3, 4}
	floats := []float64{1.5, 2.5, 3.0}

	fmt.Println("sumInts:  ", sumInts(ints))
	fmt.Println("sumFloats:", sumFloats(floats))

	fmt.Println("Sum(ints):    ", Sum(ints))
	fmt.Println("Sum(floats):  ", Sum(floats))
	fmt.Println("Sum(int64s):  ", Sum([]int64{10, 20, 30}))

	viaAggregate := aggregate(ints, 0, func(acc, v int) int { return acc + v })
	fmt.Println("aggregate as sum (works, but reads worse than Sum): ", viaAggregate)

	// A guideline for this course:
	//   - Two call sites doing the same thing: duplication is fine.
	//   - Three or more call sites sharing one true shape: extract Sum-like
	//     generic helper, constrained as narrowly as the real callers need.
	//   - A single hypothetical future caller: do not generalize yet. Add
	//     the parameter or the type parameter when that caller actually
	//     exists, not before.
	fmt.Println("guideline: prefer the narrowest abstraction that today's callers need")
}

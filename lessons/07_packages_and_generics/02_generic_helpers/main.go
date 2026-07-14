// Package main introduces type parameters, constraints, and small generic
// helper functions.
package main

import (
	"cmp"
	"fmt"
)

// Map applies fn to every element of in and returns a new slice of the
// results. The two type parameters, In and Out, may differ: Map can turn a
// []int into a []string, for example. "any" is an alias for "interface{}"
// and is used here because Map places no requirements on the element types
// beyond existing.
func Map[In, Out any](in []In, fn func(In) Out) []Out {
	out := make([]Out, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

// Filter returns the elements of in for which keep reports true. T is
// constrained only by "any" because Filter never compares or orders
// elements - it only calls keep and copies matching values.
func Filter[T any](in []T, keep func(T) bool) []T {
	var out []T
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reduce folds in into a single accumulated value, starting from initial and
// combining one element at a time with combine. This is a generalization of
// patterns like "sum all elements" or "build the longest string".
func Reduce[In, Acc any](in []In, initial Acc, combine func(Acc, In) Acc) Acc {
	acc := initial
	for _, v := range in {
		acc = combine(acc, v)
	}
	return acc
}

// Number is a custom constraint: an interface whose type set is every
// listed type instead of a set of methods. "~int" means "int, or any type
// whose underlying type is int," which lets Number also accept named types
// like `type Celsius int`. Constraints like this are how generic code
// states "any of these concrete kinds of number" without duplicating a
// function per type.
type Number interface {
	~int | ~int64 | ~float64
}

// Sum adds every element of in using the Number constraint above.
func Sum[T Number](in []T) T {
	var total T
	for _, v := range in {
		total += v
	}
	return total
}

// Max returns the largest element of in. cmp.Ordered is a standard-library
// constraint (package "cmp") satisfied by every built-in type that supports
// <, <=, >, >=. Max panics on an empty slice, matching the standard
// library's own slices.Max, so callers must check length first when it
// might be empty.
func Max[T cmp.Ordered](in []T) T {
	max := in[0]
	for _, v := range in[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func main() {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}

	doubled := Map(numbers, func(n int) int { return n * 2 })
	fmt.Println("doubled:", doubled)

	labels := Map(numbers, func(n int) string { return fmt.Sprintf("n%d", n) })
	fmt.Println("labels: ", labels)

	even := Filter(numbers, func(n int) bool { return n%2 == 0 })
	fmt.Println("even:   ", even)

	total := Reduce(numbers, 0, func(acc, n int) int { return acc + n })
	fmt.Println("total:  ", total)

	fmt.Println("Sum(numbers):        ", Sum(numbers))
	fmt.Println("Sum(float64 slice):  ", Sum([]float64{1.5, 2.5, 3.0}))
	fmt.Println("Max(numbers):        ", Max(numbers))
	fmt.Println("Max(string slice):   ", Max([]string{"pear", "apple", "plum"}))

	// Type inference: Go infers In/Out/T from the arguments in almost every
	// call above, so the type parameters rarely need to be written out. They
	// can be given explicitly when inference is not possible or for clarity:
	explicitTotal := Reduce[int, int](numbers, 0, func(acc, n int) int { return acc + n })
	fmt.Println("explicit type args:  ", explicitTotal)
}

// Package main introduces iterator functions and range-over-func: ranging
// directly over a function value using the standard "iter" package's
// function-shaped sequences.
package main

import (
	"fmt"
	"iter"
	"slices"
)

// Range builds an iterator in three layers:
//  1. Range itself returns a function (an iter.Seq[int]).
//  2. The range loop calls that function with a compiler-provided yield
//     callback.
//  3. The iterator calls yield once per value; false means the loop stopped
//     early, for example with break.
//
// This is the same closure idea as module 3, with the callback protocol
// standardized so "for v := range Range(...)" works without an intermediate
// slice.
func Range(start, end int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for v := start; v < end; v++ {
			if !yield(v) {
				return
			}
		}
	}
}

// Evens wraps another iter.Seq[int] and yields only its even values. This
// shows an iterator built by composing another iterator, the same way
// Filter in the previous lesson composed a slice.
func Evens(seq iter.Seq[int]) iter.Seq[int] {
	return func(yield func(int) bool) {
		for v := range seq {
			if v%2 != 0 {
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

func main() {
	fmt.Println("--- range-over-func with a custom iterator ---")
	for v := range Range(0, 5) {
		fmt.Println("value:", v)
	}

	fmt.Println("--- stopping early with break ---")
	for v := range Range(0, 100) {
		if v == 3 {
			break // this causes Range's yield call to return false and stop
		}
		fmt.Println("value before break:", v)
	}

	fmt.Println("--- composed iterator ---")
	for v := range Evens(Range(0, 10)) {
		fmt.Println("even:", v)
	}

	fmt.Println("--- standard library iterators ---")
	// slices.Values turns an existing slice into an iter.Seq without
	// copying it, so existing data can be ranged over uniformly alongside
	// custom sequences like Range and Evens above.
	names := []string{"ada", "grace", "linus"}
	for name := range slices.Values(names) {
		fmt.Println("name:", name)
	}

	// slices.Collect (the reverse of slices.Values) gathers any iter.Seq
	// back into a slice, which is useful once an iterator's values need to
	// be stored, sorted, or measured with len().
	collected := slices.Collect(Evens(Range(0, 12)))
	fmt.Println("collected evens:", collected)
}

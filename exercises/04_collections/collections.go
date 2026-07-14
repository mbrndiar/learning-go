// Package collections contains starter exercises covering slices, maps,
// copy, sorting, and shared-backing-array awareness. Replace every
// panic("TODO: ...") with a working implementation.
package collections

import "errors"

// ErrIndexOutOfRange is returned by RemoveAt when idx is not a valid index
// into the input slice.
var ErrIndexOutOfRange = errors.New("index out of range")

// Sum returns the sum of the elements of nums.
func Sum(nums []int) int {
	panic("TODO: implement Sum")
}

// Unique returns the distinct elements of nums, preserving the order of
// first appearance.
func Unique(nums []int) []int {
	panic("TODO: implement Unique")
}

// WordFrequency splits text on whitespace and returns a map counting how
// many times each word occurs.
func WordFrequency(text string) map[string]int {
	panic("TODO: implement WordFrequency")
}

// MergeCounts returns a new map containing the counts from a and b, summing
// values for keys present in both. Neither a nor b is mutated.
func MergeCounts(a, b map[string]int) map[string]int {
	panic("TODO: implement MergeCounts")
}

// SortDescending returns a new slice containing the elements of nums
// sorted from largest to smallest. nums itself is not mutated.
func SortDescending(nums []int) []int {
	panic("TODO: implement SortDescending")
}

// RemoveAt returns a new slice with the element at idx removed. nums itself
// is not mutated. It returns ErrIndexOutOfRange if idx is not a valid index
// into nums.
func RemoveAt(nums []int, idx int) ([]int, error) {
	panic("TODO: implement RemoveAt")
}

// CloneInts returns an independent copy of nums: mutating the returned
// slice must never affect nums, and vice versa.
func CloneInts(nums []int) []int {
	panic("TODO: implement CloneInts")
}

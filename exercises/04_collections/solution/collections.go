// Package solution is the reference implementation for
// exercises/04_collections.
package solution

import (
	"cmp"
	"errors"
	"slices"
	"strings"
)

// ErrIndexOutOfRange is returned by RemoveAt when idx is not a valid index
// into the input slice.
var ErrIndexOutOfRange = errors.New("index out of range")

// Sum returns the sum of the elements of nums.
func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Unique returns the distinct elements of nums, preserving the order of
// first appearance.
func Unique(nums []int) []int {
	seen := make(map[int]bool, len(nums))
	result := make([]int, 0, len(nums))
	for _, n := range nums {
		if seen[n] {
			continue
		}
		seen[n] = true
		result = append(result, n)
	}
	return result
}

// WordFrequency splits text on whitespace and returns a map counting how
// many times each word occurs.
func WordFrequency(text string) map[string]int {
	counts := make(map[string]int)
	for _, word := range strings.Fields(text) {
		counts[word]++
	}
	return counts
}

// MergeCounts returns a new map containing the counts from a and b, summing
// values for keys present in both. Neither a nor b is mutated.
func MergeCounts(a, b map[string]int) map[string]int {
	merged := make(map[string]int, len(a)+len(b))
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		merged[k] += v
	}
	return merged
}

// SortDescending returns a new slice containing the elements of nums
// sorted from largest to smallest. nums itself is not mutated because
// sorting happens on a freshly copied slice.
func SortDescending(nums []int) []int {
	result := make([]int, len(nums))
	copy(result, nums)
	slices.SortFunc(result, func(a, b int) int {
		return cmp.Compare(b, a)
	})
	return result
}

// RemoveAt returns a new slice with the element at idx removed. nums itself
// is not mutated. It returns ErrIndexOutOfRange if idx is not a valid index
// into nums.
func RemoveAt(nums []int, idx int) ([]int, error) {
	if idx < 0 || idx >= len(nums) {
		return nil, ErrIndexOutOfRange
	}
	result := make([]int, 0, len(nums)-1)
	result = append(result, nums[:idx]...)
	result = append(result, nums[idx+1:]...)
	return result, nil
}

// CloneInts returns an independent copy of nums: mutating the returned
// slice never affects nums, and vice versa, because copy populates a
// freshly allocated backing array.
func CloneInts(nums []int) []int {
	clone := make([]int, len(nums))
	copy(clone, nums)
	return clone
}

// Package controlflow contains starter exercises covering classification,
// loops, and search. Replace every panic("TODO: ...") with a working
// implementation.
package controlflow

import "errors"

// ErrInvalidScore is returned by Grade when score is outside the valid
// 0-100 range.
var ErrInvalidScore = errors.New("score must be between 0 and 100")

// ClassifyNumber returns "negative", "zero", or "positive" depending on the
// sign of n.
func ClassifyNumber(n int) string {
	panic("TODO: implement ClassifyNumber")
}

// Grade converts a 0-100 score into a letter grade:
//
//	90-100 -> "A"
//	80-89  -> "B"
//	70-79  -> "C"
//	60-69  -> "D"
//	0-59   -> "F"
//
// Scores outside 0-100 return ErrInvalidScore.
func Grade(score int) (string, error) {
	panic("TODO: implement Grade")
}

// SumRange returns the sum of all integers from start to end, inclusive.
// If start > end, it returns 0.
func SumRange(start, end int) int {
	panic("TODO: implement SumRange")
}

// FizzBuzz returns a slice of n strings for the numbers 1..n. Multiples of 3
// become "Fizz", multiples of 5 become "Buzz", multiples of both become
// "FizzBuzz", and all other numbers become their decimal string form. For
// n <= 0, it returns an empty (non-nil) slice.
func FizzBuzz(n int) []string {
	panic("TODO: implement FizzBuzz")
}

// CountDigits returns the number of decimal digits in n, ignoring sign.
// CountDigits(0) is 1.
func CountDigits(n int) int {
	panic("TODO: implement CountDigits")
}

// LinearSearch returns the index of the first occurrence of target in nums,
// or -1 if it is not present.
func LinearSearch(nums []int, target int) int {
	panic("TODO: implement LinearSearch")
}

// BinarySearch returns the index of target in nums, or -1 if it is not
// present. nums must already be sorted in ascending order.
func BinarySearch(nums []int, target int) int {
	panic("TODO: implement BinarySearch")
}

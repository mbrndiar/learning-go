// Package solution is the reference implementation for
// exercises/02_control_flow.
package solution

import (
	"errors"
	"strconv"
)

// ErrInvalidScore is returned by Grade when score is outside the valid
// 0-100 range.
var ErrInvalidScore = errors.New("score must be between 0 and 100")

// ClassifyNumber returns "negative", "zero", or "positive" depending on the
// sign of n.
func ClassifyNumber(n int) string {
	switch {
	case n < 0:
		return "negative"
	case n > 0:
		return "positive"
	default:
		return "zero"
	}
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
	if score < 0 || score > 100 {
		return "", ErrInvalidScore
	}
	switch {
	case score >= 90:
		return "A", nil
	case score >= 80:
		return "B", nil
	case score >= 70:
		return "C", nil
	case score >= 60:
		return "D", nil
	default:
		return "F", nil
	}
}

// SumRange returns the sum of all integers from start to end, inclusive.
// If start > end, it returns 0.
func SumRange(start, end int) int {
	sum := 0
	for i := start; i <= end; i++ {
		sum += i
	}
	return sum
}

// FizzBuzz returns a slice of n strings for the numbers 1..n. Multiples of 3
// become "Fizz", multiples of 5 become "Buzz", multiples of both become
// "FizzBuzz", and all other numbers become their decimal string form. For
// n <= 0, it returns an empty (non-nil) slice.
func FizzBuzz(n int) []string {
	result := make([]string, 0, max(n, 0))
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0:
			result = append(result, "FizzBuzz")
		case i%3 == 0:
			result = append(result, "Fizz")
		case i%5 == 0:
			result = append(result, "Buzz")
		default:
			result = append(result, strconv.Itoa(i))
		}
	}
	return result
}

// CountDigits returns the number of decimal digits in n, ignoring sign.
// CountDigits(0) is 1.
func CountDigits(n int) int {
	if n < 0 {
		n = -n
	}
	if n == 0 {
		return 1
	}
	count := 0
	for n > 0 {
		count++
		n /= 10
	}
	return count
}

// LinearSearch returns the index of the first occurrence of target in nums,
// or -1 if it is not present.
func LinearSearch(nums []int, target int) int {
	for i, n := range nums {
		if n == target {
			return i
		}
	}
	return -1
}

// BinarySearch returns the index of target in nums, or -1 if it is not
// present. nums must already be sorted in ascending order.
func BinarySearch(nums []int, target int) int {
	lo, hi := 0, len(nums)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		switch {
		case nums[mid] == target:
			return mid
		case nums[mid] < target:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return -1
}

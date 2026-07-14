// Package basictests introduces the fundamentals of the testing package:
// test function naming, *testing.T, and the difference between reporting a
// failure and stopping a test immediately.
package basictests

import "strings"

// Reverse returns s with its runes in reverse order. It operates on runes,
// not bytes, so multi-byte UTF-8 characters are preserved correctly.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome reports whether s reads the same forwards and backwards,
// ignoring case.
func IsPalindrome(s string) bool {
	lower := strings.ToLower(s)
	return lower == Reverse(lower)
}

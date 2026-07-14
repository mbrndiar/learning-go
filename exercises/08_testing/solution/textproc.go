// Package solution is the reference implementation for exercises/08_testing.
// It mirrors the starter's textproc functions and adds the table-driven
// tests, benchmark, and fuzz target that the starter package leaves as TODOs.
package solution

import (
	"fmt"
	"strings"
)

// Normalize trims leading and trailing whitespace, collapses any interior
// run of whitespace to a single space, and lowercases the result.
func Normalize(s string) string {
	fields := strings.Fields(s)
	return strings.ToLower(strings.Join(fields, " "))
}

// WordFrequency splits the normalized form of s on whitespace and returns how
// many times each word occurs. An empty or whitespace-only s yields an empty,
// non-nil map.
func WordFrequency(s string) map[string]int {
	counts := make(map[string]int)
	normalized := Normalize(s)
	if normalized == "" {
		return counts
	}
	for _, word := range strings.Split(normalized, " ") {
		counts[word]++
	}
	return counts
}

// Reverse returns s with its runes in reverse order. Multi-byte runes stay
// intact; only rune order is reversed, never their internal bytes.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// SafeSlice returns the substring of s starting at rune index start with the
// given rune length. Unlike raw slice expressions, it never panics: an
// out-of-range or negative start/length returns a descriptive error instead.
func SafeSlice(s string, start, length int) (string, error) {
	if start < 0 {
		return "", fmt.Errorf("textproc: start %d is negative", start)
	}
	if length < 0 {
		return "", fmt.Errorf("textproc: length %d is negative", length)
	}
	runes := []rune(s)
	if start > len(runes) {
		return "", fmt.Errorf("textproc: start %d exceeds rune length %d", start, len(runes))
	}
	end := start + length
	if end > len(runes) {
		return "", fmt.Errorf("textproc: end %d exceeds rune length %d", end, len(runes))
	}
	return string(runes[start:end]), nil
}

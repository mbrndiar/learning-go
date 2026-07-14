// Package solution is the reference implementation for exercises/01_basics.
package solution

import (
	"strconv"
	"strings"
)

// CelsiusToFahrenheit converts a Celsius temperature to Fahrenheit using
// F = C*9/5 + 32.
func CelsiusToFahrenheit(c float64) float64 {
	return c*9/5 + 32
}

// FahrenheitToCelsius converts a Fahrenheit temperature to Celsius using
// C = (F-32)*5/9.
func FahrenheitToCelsius(f float64) float64 {
	return (f - 32) * 5 / 9
}

// ParseIntOrDefault converts s to an int. If s cannot be parsed as a base-10
// integer, def is returned instead.
func ParseIntOrDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// ReverseString returns s with its runes in reverse order. Multi-byte runes
// remain intact because reversal happens on the []rune view of s, not on
// its raw bytes.
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CountVowels returns the number of vowels (a, e, i, o, u, either case) in s,
// counted by rune.
func CountVowels(s string) int {
	count := 0
	for _, r := range s {
		switch r {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			count++
		}
	}
	return count
}

// IsPalindrome reports whether s reads the same forwards and backwards,
// ignoring ASCII case, comparing by rune.
func IsPalindrome(s string) bool {
	runes := []rune(strings.ToLower(s))
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if runes[i] != runes[j] {
			return false
		}
	}
	return true
}

// ByteAndRuneLen returns the byte length and the rune length of s, in that
// order.
func ByteAndRuneLen(s string) (byteLen int, runeLen int) {
	return len(s), len([]rune(s))
}

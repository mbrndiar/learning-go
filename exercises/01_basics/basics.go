// Package basics contains starter exercises covering type conversions and
// string/rune helpers. Replace every panic("TODO: ...") with a working
// implementation.
package basics

// CelsiusToFahrenheit converts a Celsius temperature to Fahrenheit using
// F = C*9/5 + 32.
func CelsiusToFahrenheit(c float64) float64 {
	panic("TODO: implement CelsiusToFahrenheit")
}

// FahrenheitToCelsius converts a Fahrenheit temperature to Celsius using
// C = (F-32)*5/9.
func FahrenheitToCelsius(f float64) float64 {
	panic("TODO: implement FahrenheitToCelsius")
}

// ParseIntOrDefault converts s to an int. If s cannot be parsed as a base-10
// integer, def is returned instead.
func ParseIntOrDefault(s string, def int) int {
	panic("TODO: implement ParseIntOrDefault")
}

// ReverseString returns s with its runes in reverse order. Multi-byte runes
// must remain intact (no broken UTF-8 sequences).
func ReverseString(s string) string {
	panic("TODO: implement ReverseString")
}

// CountVowels returns the number of vowels (a, e, i, o, u, either case) in s,
// counted by rune.
func CountVowels(s string) int {
	panic("TODO: implement CountVowels")
}

// IsPalindrome reports whether s reads the same forwards and backwards,
// ignoring ASCII case, comparing by rune.
func IsPalindrome(s string) bool {
	panic("TODO: implement IsPalindrome")
}

// ByteAndRuneLen returns the byte length and the rune length of s, in that
// order.
func ByteAndRuneLen(s string) (byteLen int, runeLen int) {
	panic("TODO: implement ByteAndRuneLen")
}

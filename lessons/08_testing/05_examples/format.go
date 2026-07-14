// Package examples implements small text-formatting helpers used to
// demonstrate Go's Example functions: tests that double as documentation
// and are verified by comparing their printed output to a "// Output:"
// comment.
package examples

import (
	"fmt"
	"strings"
)

// TitleCase upper-cases the first letter of every space-separated word and
// lower-cases the rest.
func TitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		lower := strings.ToLower(word)
		words[i] = strings.ToUpper(lower[:1]) + lower[1:]
	}
	return strings.Join(words, " ")
}

// FormatCents renders an integer number of cents as a "$X.YY" string.
func FormatCents(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$%d.%02d", sign, cents/100, cents%100)
}

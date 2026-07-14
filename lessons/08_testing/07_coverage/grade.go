// Package coverage implements a small grade classifier used to practice
// reading `go test -cover` output: this package's tests intentionally leave
// one branch unexercised so you can find the gap yourself with the
// coverage tool instead of only reading the source.
package coverage

import "fmt"

// Classify maps a numeric score in [0, 100] to a letter grade.
func Classify(score int) (string, error) {
	switch {
	case score < 0 || score > 100:
		return "", fmt.Errorf("classify: score %d out of range [0, 100]", score)
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

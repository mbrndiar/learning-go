// Package pricing lives under an "internal" directory. The Go toolchain
// enforces that any package rooted at a path containing "internal" can only
// be imported by packages whose own import path shares the directory that
// is the parent of "internal" - here, everything under this lesson's
// "01_package_organization" tree. Code in a sibling lesson, a different
// module, or an external repository cannot import this package at all; the
// compiler rejects the import rather than merely discouraging it by
// convention. This is Go's built-in way to draw a hard API boundary inside
// a single module.
package pricing

// ApplyDiscount reduces price by percent (0-100) and reports the result.
// This is intentionally simple: it exists to be imported, not to teach
// pricing math.
func ApplyDiscount(price float64, percent float64) float64 {
	if percent <= 0 {
		return price
	}
	if percent >= 100 {
		return 0
	}
	return price * (1 - percent/100)
}

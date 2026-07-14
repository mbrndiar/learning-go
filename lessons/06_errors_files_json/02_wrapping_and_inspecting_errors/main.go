// Package main covers wrapping errors with fmt.Errorf's %w verb and
// inspecting wrapped error chains with errors.Is and errors.As.
package main

import (
	"errors"
	"fmt"
)

// ErrPermission is a sentinel error representing an access failure deep in a
// call chain.
var ErrPermission = errors.New("permission denied")

// QuotaError is a typed error carrying structured detail.
type QuotaError struct {
	Limit int
	Used  int
}

func (e *QuotaError) Error() string {
	return fmt.Sprintf("quota exceeded: used %d of %d", e.Used, e.Limit)
}

// readSecret simulates a low-level operation that fails with a sentinel
// error.
func readSecret(allowed bool) error {
	if !allowed {
		return ErrPermission
	}
	return nil
}

// checkQuota simulates a mid-level operation that fails with a typed error.
func checkQuota(used, limit int) error {
	if used > limit {
		return &QuotaError{Limit: limit, Used: used}
	}
	return nil
}

// loadResource is the high-level operation callers actually invoke. It wraps
// each low-level error with %w instead of %v. %w preserves the original
// error inside the returned one, forming a chain that errors.Is and
// errors.As can walk through, while %v would flatten it into a plain string
// and destroy that chain.
func loadResource(allowed bool, used, limit int) error {
	if err := readSecret(allowed); err != nil {
		return fmt.Errorf("loadResource: read secret: %w", err)
	}
	if err := checkQuota(used, limit); err != nil {
		return fmt.Errorf("loadResource: check quota: %w", err)
	}
	return nil
}

func main() {
	cases := []struct {
		name    string
		allowed bool
		used    int
		limit   int
	}{
		{name: "denied", allowed: false, used: 0, limit: 10},
		{name: "over quota", allowed: true, used: 12, limit: 10},
		{name: "success", allowed: true, used: 3, limit: 10},
	}

	for _, c := range cases {
		err := loadResource(c.allowed, c.used, c.limit)
		fmt.Println("case:", c.name)

		if err == nil {
			fmt.Println("  no error")
			continue
		}

		// The wrapped message still includes context from every layer.
		fmt.Println("  message:", err)

		// errors.Is walks the chain produced by %w and reports whether any
		// error in it matches the target, regardless of how deeply it is
		// wrapped. Comparing with == would only work on the outermost
		// fmt.Errorf value, which never equals ErrPermission directly.
		if errors.Is(err, ErrPermission) {
			fmt.Println("  matched: permission denied (via errors.Is)")
		}

		// errors.As walks the same chain looking for an error whose
		// *concrete type* matches the target pointer, and - if found -
		// assigns it so the structured fields become reachable.
		var quotaErr *QuotaError
		if errors.As(err, &quotaErr) {
			fmt.Printf("  matched: quota error (via errors.As): used=%d limit=%d\n", quotaErr.Used, quotaErr.Limit)
		}

		// Unwrap manually to show one layer of the chain at a time.
		fmt.Println("  errors.Unwrap once:", errors.Unwrap(err))
	}

	// errors.Join (standard library since Go 1.20) combines independent
	// errors into one value that errors.Is/As can still inspect.
	joined := errors.Join(ErrPermission, &QuotaError{Limit: 5, Used: 9})
	fmt.Println("joined error:", joined)
	fmt.Println("joined matches ErrPermission:", errors.Is(joined, ErrPermission))
}

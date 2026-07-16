// Package cli owns the testable monitor command boundary.
package cli

import (
	"context"
	"fmt"
	"io"
)

// Run is the stable check/serve process boundary.
func Run(_ context.Context, _ []string, _ io.Writer, stderr io.Writer) int {
	writePlaceholder(stderr)
	return 1
}

func writePlaceholder(stderr io.Writer) {
	fmt.Fprintln(stderr, "monitor: not implemented")
}

// Package app owns the testable comparative command boundary.
package app

import (
	"context"
	"fmt"
	"io"
)

// Run is the stable process boundary used by the thin command.
func Run(ctx context.Context, args []string, stdout io.Writer, _ io.Writer) int {
	writePlaceholder(stdout)
	return 1
}

func writePlaceholder(stdout io.Writer) {
	fmt.Fprintln(stdout, `{"ok":false,"error":{"category":"not_implemented","details":{}}}`)
}

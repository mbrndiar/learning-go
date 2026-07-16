// Package contract contains reusable comparative harness assertions.
package contract

import (
	"bytes"
	"context"
	"testing"
)

// Harness adapts either implementation tree to the shared smoke contract.
type Harness struct {
	Name          string
	Implemented   bool
	ParseValue    func() error
	OpenStore     func(context.Context) error
	IsIncomplete  func(error) bool
	RunCLI        func(context.Context, *bytes.Buffer, *bytes.Buffer) int
	Placeholder   string
	PlaceholderRC int
}

// RunHarness proves that packages load and unfinished behavior is explicit.
func RunHarness(t *testing.T, harness Harness) {
	t.Helper()

	t.Run(harness.Name, func(t *testing.T) {
		if harness.Implemented {
			t.Skip("implementation placeholders have been replaced")
		}

		t.Run("domain placeholder", func(t *testing.T) {
			if err := harness.ParseValue(); !harness.IsIncomplete(err) {
				t.Fatalf("ParseValue() error = %v, want ErrNotImplemented", err)
			}
		})

		t.Run("storage placeholder", func(t *testing.T) {
			if err := harness.OpenStore(context.Background()); !harness.IsIncomplete(err) {
				t.Fatalf("Open() error = %v, want ErrNotImplemented", err)
			}
		})

		t.Run("command placeholder", func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := harness.RunCLI(context.Background(), &stdout, &stderr)
			if exitCode != harness.PlaceholderRC {
				t.Fatalf("Run() exit = %d, want %d", exitCode, harness.PlaceholderRC)
			}
			if stdout.String() != harness.Placeholder {
				t.Fatalf("stdout = %q, want %q", stdout.String(), harness.Placeholder)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	})
}

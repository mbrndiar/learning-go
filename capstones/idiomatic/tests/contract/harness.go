// Package contract contains reusable idiomatic harness assertions.
package contract

import (
	"bytes"
	"context"
	"net/http"
	"testing"
)

// ProbeResult is the implementation-neutral placeholder observation.
type ProbeResult struct {
	Status  string
	Message string
}

// HTTPResult is the implementation-neutral placeholder handler response.
type HTTPResult struct {
	StatusCode  int
	ContentType string
	Body        string
}

// Harness adapts either implementation tree to the shared smoke contract.
type Harness struct {
	Name              string
	Implemented       bool
	LoadConfig        func() error
	Probe             func(context.Context) ProbeResult
	Record            func() error
	Start             func(context.Context) error
	Wait              func() error
	Serve             func() HTTPResult
	RunCLI            func(context.Context, *bytes.Buffer, *bytes.Buffer) int
	IsIncomplete      func(error) bool
	ProbeMessage      string
	APIResponse       string
	CommandDiagnostic string
	PlaceholderRC     int
}

// RunHarness proves that packages load and unfinished behavior is explicit.
func RunHarness(t *testing.T, harness Harness) {
	t.Helper()

	t.Run(harness.Name, func(t *testing.T) {
		if harness.Implemented {
			t.Skip("implementation placeholders have been replaced")
		}

		t.Run("error placeholders", func(t *testing.T) {
			checks := []struct {
				name string
				call func() error
			}{
				{name: "LoadConfig", call: harness.LoadConfig},
				{name: "Record", call: harness.Record},
				{name: "Start", call: func() error {
					return harness.Start(context.Background())
				}},
				{name: "Wait", call: harness.Wait},
			}
			for _, check := range checks {
				t.Run(check.name, func(t *testing.T) {
					if err := check.call(); !harness.IsIncomplete(err) {
						t.Fatalf("error = %v, want ErrNotImplemented", err)
					}
				})
			}
		})

		t.Run("probe placeholder", func(t *testing.T) {
			result := harness.Probe(context.Background())
			if result.Status != "unknown" || result.Message != harness.ProbeMessage {
				t.Fatalf("Probe() = %+v, want unknown explicit placeholder", result)
			}
		})

		t.Run("api placeholder", func(t *testing.T) {
			result := harness.Serve()
			if result.StatusCode != http.StatusNotImplemented {
				t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusNotImplemented)
			}
			if result.ContentType != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q", result.ContentType)
			}
			if result.Body != harness.APIResponse {
				t.Fatalf("body = %q, want %q", result.Body, harness.APIResponse)
			}
		})

		t.Run("command placeholder", func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := harness.RunCLI(context.Background(), &stdout, &stderr)
			if exitCode != harness.PlaceholderRC {
				t.Fatalf("Run() exit = %d, want %d", exitCode, harness.PlaceholderRC)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.String() != harness.CommandDiagnostic {
				t.Fatalf("stderr = %q, want %q", stderr.String(), harness.CommandDiagnostic)
			}
		})
	})
}

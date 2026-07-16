// Package m5 contains shared Milestone 5 process-boundary assertions.
package m5

import (
	"encoding/json"
	"testing"
)

// RequireJSONError checks the stable command error envelope.
func RequireJSONError(t *testing.T, data []byte, wantCode string) {
	t.Helper()
	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("decode JSON error: %v; body=%q", err, data)
	}
	if response.Error.Code != wantCode {
		t.Fatalf("error code = %q, want %q", response.Error.Code, wantCode)
	}
}

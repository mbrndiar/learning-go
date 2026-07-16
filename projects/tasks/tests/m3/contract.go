package m3

import (
	"encoding/json"
	"net/http"
	"testing"
)

// AssertErrorResponse applies the shared Milestone 3 JSON error-envelope contract.
func AssertErrorResponse(
	t testing.TB,
	status int,
	headers http.Header,
	body []byte,
	wantStatus int,
	code string,
	message string,
	field string,
) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if status != wantStatus || json.Unmarshal(body, &envelope) != nil ||
		envelope.Error.Code != code || envelope.Error.Message != message {
		t.Fatalf("error response = %d %s", status, body)
	}
	if headers.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", headers.Get("Content-Type"))
	}
	if field == "" && envelope.Error.Details != nil {
		t.Fatalf("unexpected details: %v", envelope.Error.Details)
	}
	if field != "" && envelope.Error.Details["field"] != field {
		t.Fatalf("details = %v, want field %q", envelope.Error.Details, field)
	}
}

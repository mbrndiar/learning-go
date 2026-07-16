// Package m4 contains shared Milestone 4 HTTP assertions.
package m4

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Request invokes an ordinary handler and checks its status and JSON content type.
func Request(t *testing.T, handler http.Handler, method, target string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))
	if recorder.Code != wantStatus {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, target, recorder.Code, wantStatus, recorder.Body)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	return recorder
}

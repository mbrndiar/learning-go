package taskapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// discardLogger returns a logger that drops output so error-path tests stay
// quiet while still exercising the logging code.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// errStore is a Store whose methods fail, letting tests drive the 500 path.
type errStore struct{ err error }

func (s errStore) List(context.Context) ([]Task, error) { return nil, s.err }
func (s errStore) Get(context.Context, int64) (Task, error) {
	return Task{}, s.err
}
func (s errStore) Add(context.Context, string) (Task, error) {
	return Task{}, s.err
}
func (s errStore) Complete(context.Context, int64) (Task, error) {
	return Task{}, s.err
}
func (s errStore) Remove(context.Context, int64) error { return s.err }

func newTestServer(t *testing.T, opts ...APIOption) *httptest.Server {
	t.Helper()
	store := newMemoryStore(t)
	api, err := NewAPI(store, opts...)
	if err != nil {
		t.Fatalf("NewAPI() error = %v", err)
	}
	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)
	return server
}

func doRequest(t *testing.T, server *httptest.Server, method, path, body string) (*http.Response, string) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, server.URL+path, reader)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return resp, string(data)
}

func TestNewAPIRejectsNilStore(t *testing.T) {
	if _, err := NewAPI(nil); err == nil {
		t.Fatal("NewAPI(nil) error = nil, want error")
	}
}

func TestAPIRoundTrip(t *testing.T) {
	server := newTestServer(t)

	resp, body := doRequest(t, server, http.MethodPost, "/tasks", `{"title":"ship it"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /tasks status = %d, want 201; body=%s", resp.StatusCode, body)
	}
	var created Task
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatalf("decode created task: %v", err)
	}
	if created.ID <= 0 || created.Title != "ship it" {
		t.Fatalf("created = %+v, want positive id and title", created)
	}

	resp, body = doRequest(t, server, http.MethodGet, "/tasks", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /tasks status = %d, want 200", resp.StatusCode)
	}
	var list []Task
	if err := json.Unmarshal([]byte(body), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}

	resp, _ = doRequest(t, server, http.MethodPost, "/tasks/1/complete", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("complete status = %d, want 200", resp.StatusCode)
	}

	resp, _ = doRequest(t, server, http.MethodDelete, "/tasks/1", "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", resp.StatusCode)
	}

	resp, _ = doRequest(t, server, http.MethodGet, "/tasks/1", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want 404", resp.StatusCode)
	}
}

func TestAPIStructuredNotFound(t *testing.T) {
	server := newTestServer(t)
	resp, body := doRequest(t, server, http.MethodGet, "/tasks/999", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content-type = %q, want json", ct)
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if payload.Error == "" {
		t.Fatal("error body missing error field")
	}
}

func TestAPIValidationErrors(t *testing.T) {
	server := newTestServer(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"invalid json", http.MethodPost, "/tasks", `{"title":`, http.StatusBadRequest},
		{"unknown field", http.MethodPost, "/tasks", `{"title":"x","bogus":true}`, http.StatusBadRequest},
		{"empty title", http.MethodPost, "/tasks", `{"title":"   "}`, http.StatusBadRequest},
		{"trailing data", http.MethodPost, "/tasks", `{"title":"a"}{}`, http.StatusBadRequest},
		{"non-numeric id", http.MethodGet, "/tasks/abc", "", http.StatusBadRequest},
		{"zero id", http.MethodGet, "/tasks/0", "", http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, body := doRequest(t, server, test.method, test.path, test.body)
			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, test.wantStatus, body)
			}
		})
	}
}

func TestAPIRequestSizeLimit(t *testing.T) {
	server := newTestServer(t, WithMaxBodyBytes(32))
	oversized := `{"title":"` + strings.Repeat("a", 200) + `"}`
	resp, body := doRequest(t, server, http.MethodPost, "/tasks", oversized)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for oversized body; body=%s", resp.StatusCode, body)
	}
}

func TestAPIMethodNotAllowed(t *testing.T) {
	server := newTestServer(t)
	resp, _ := doRequest(t, server, http.MethodPut, "/tasks", "")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("PUT /tasks status = %d, want 405", resp.StatusCode)
	}
}

func TestAPICompleteAndRemoveErrors(t *testing.T) {
	server := newTestServer(t)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"complete missing", http.MethodPost, "/tasks/999/complete", http.StatusNotFound},
		{"remove missing", http.MethodDelete, "/tasks/999", http.StatusNotFound},
		{"complete bad id", http.MethodPost, "/tasks/abc/complete", http.StatusBadRequest},
		{"remove bad id", http.MethodDelete, "/tasks/abc", http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, body := doRequest(t, server, test.method, test.path, "")
			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", resp.StatusCode, test.wantStatus, body)
			}
		})
	}
}

func TestNewServerHasFiniteTimeouts(t *testing.T) {
	server := NewServer(":9999", http.NewServeMux())
	if server.Addr != ":9999" {
		t.Fatalf("Addr = %q, want :9999", server.Addr)
	}
	if server.ReadHeaderTimeout <= 0 || server.ReadTimeout <= 0 || server.WriteTimeout <= 0 || server.IdleTimeout <= 0 {
		t.Fatalf("server timeouts must all be finite and positive: %+v", server)
	}
}

func TestAPIInternalErrorIsStructured(t *testing.T) {
	api, err := NewAPI(errStore{err: errors.New("db exploded")}, WithLogger(discardLogger()))
	if err != nil {
		t.Fatalf("NewAPI() error = %v", err)
	}
	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)

	resp, body := doRequest(t, server, http.MethodGet, "/tasks", "")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	if !strings.Contains(body, "internal server error") {
		t.Fatalf("body = %q, want structured internal error", body)
	}
	if strings.Contains(body, "db exploded") {
		t.Fatalf("body leaked internal error detail: %q", body)
	}
}

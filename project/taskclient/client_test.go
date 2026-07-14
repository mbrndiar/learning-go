package taskclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeAPI is a minimal, configurable stand-in for the real task API so the
// client can be tested in isolation, including malformed responses.
type fakeAPI struct {
	status int
	body   string
	delay  time.Duration
	gotReq *http.Request
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.gotReq = r
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-r.Context().Done():
			return
		}
	}
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprint(w, f.body)
}

func newClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := New(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return client
}

func TestNewValidatesBaseURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"relative", "/tasks"},
		{"scheme only", "http://"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(test.url); err == nil {
				t.Fatalf("New(%q) error = nil, want error", test.url)
			}
		})
	}
}

func TestClientListDecodesTasks(t *testing.T) {
	fake := &fakeAPI{body: `[{"id":1,"title":"a","done":false},{"id":2,"title":"b","done":true}]`}
	client := newClient(t, fake)

	tasks, err := client.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 2 || tasks[1].ID != 2 || !tasks[1].Done {
		t.Fatalf("List() = %+v, want two tasks", tasks)
	}
}

func TestClientAddSendsTitle(t *testing.T) {
	fake := &fakeAPI{status: http.StatusCreated, body: `{"id":7,"title":"remote","done":false}`}
	client := newClient(t, fake)

	task, err := client.Add(context.Background(), "remote")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if task.ID != 7 {
		t.Fatalf("Add() id = %d, want 7", task.ID)
	}
	if fake.gotReq.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", fake.gotReq.Method)
	}
}

func TestClientGetTranslatesNotFound(t *testing.T) {
	fake := &fakeAPI{status: http.StatusNotFound, body: `{"error":"task not found"}`}
	client := newClient(t, fake)

	_, err := client.Get(context.Background(), 5)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestClientRemoveNotFound(t *testing.T) {
	fake := &fakeAPI{status: http.StatusNotFound, body: `{"error":"missing"}`}
	client := newClient(t, fake)
	if err := client.Remove(context.Background(), 5); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Remove() error = %v, want ErrNotFound", err)
	}
}

func TestClientCompleteSucceeds(t *testing.T) {
	fake := &fakeAPI{body: `{"id":3,"title":"done","done":true}`}
	client := newClient(t, fake)

	task, err := client.Complete(context.Background(), 3)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if !task.Done || task.ID != 3 {
		t.Fatalf("Complete() = %+v, want id=3 done=true", task)
	}
	if fake.gotReq.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", fake.gotReq.Method)
	}
}

func TestClientRemoveSucceeds(t *testing.T) {
	fake := &fakeAPI{status: http.StatusNoContent}
	client := newClient(t, fake)
	if err := client.Remove(context.Background(), 3); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if fake.gotReq.Method != http.MethodDelete {
		t.Fatalf("method = %s, want DELETE", fake.gotReq.Method)
	}
}

func TestClientAddRejectsValidationError(t *testing.T) {
	fake := &fakeAPI{status: http.StatusBadRequest, body: `{"error":"title required"}`}
	client := newClient(t, fake)

	_, err := client.Add(context.Background(), "x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("Add() error = %v, want APIError with status 400", err)
	}
}

func TestClientPlainTextErrorMessage(t *testing.T) {
	fake := &fakeAPI{status: http.StatusBadGateway, body: "upstream exploded"}
	client := newClient(t, fake)

	_, err := client.List(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("List() error = %v, want *APIError", err)
	}
	if !strings.Contains(apiErr.Message, "upstream exploded") {
		t.Fatalf("Message = %q, want plain-text body", apiErr.Message)
	}
}

func TestClientPreservesStatusInAPIError(t *testing.T) {
	fake := &fakeAPI{status: http.StatusInternalServerError, body: `{"error":"boom"}`}
	client := newClient(t, fake)

	_, err := client.List(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("List() error = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Message, "boom") {
		t.Fatalf("Message = %q, want to contain server message", apiErr.Message)
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatal("500 error should not satisfy ErrNotFound")
	}
}

func TestClientRejectsInvalidResponse(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"malformed json", `{"id":1,`},
		{"trailing json", `{"id":1,"title":"x","done":false}{}`},
		{"non-positive id", `{"id":0,"title":"x","done":false}`},
		{"empty title", `{"id":1,"title":"  ","done":false}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newClient(t, &fakeAPI{body: test.body})
			if _, err := client.Get(context.Background(), 1); !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("Get() error = %v, want ErrInvalidResponse", err)
			}
		})
	}
}

func TestClientRejectsNonPositiveID(t *testing.T) {
	client := newClient(t, &fakeAPI{})
	if _, err := client.Get(context.Background(), 0); err == nil {
		t.Fatal("Get(0) error = nil, want error")
	}
	if _, err := client.Complete(context.Background(), -1); err == nil {
		t.Fatal("Complete(-1) error = nil, want error")
	}
	if err := client.Remove(context.Background(), 0); err == nil {
		t.Fatal("Remove(0) error = nil, want error")
	}
}

func TestClientTranslatesTimeout(t *testing.T) {
	fake := &fakeAPI{body: `[]`, delay: 200 * time.Millisecond}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	client, err := New(server.URL, WithHTTPClient(server.Client()), WithTimeout(20*time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.List(context.Background())
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("List() error = %v, want ErrTimeout", err)
	}
}

func TestClientTranslatesUnavailable(t *testing.T) {
	server := httptest.NewServer(&fakeAPI{})
	url := server.URL
	server.Close() // the address is now unreachable

	client, err := New(url, WithTimeout(500*time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.List(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("List() error = %v, want ErrUnavailable", err)
	}
}

func TestClientHonorsContextCancellation(t *testing.T) {
	client := newClient(t, &fakeAPI{body: `[]`, delay: time.Second})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("List(cancelled) error = %v, want context.Canceled", err)
	}
}

func TestClientResolvesBasePathPrefix(t *testing.T) {
	mux := http.NewServeMux()
	var gotPath string
	mux.HandleFunc("/api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[]`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := New(server.URL+"/api/v1/", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.List(context.Background()); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if gotPath != "/api/v1/tasks" {
		t.Fatalf("server saw path %q, want /api/v1/tasks", gotPath)
	}
}

func TestAPIErrorMessage(t *testing.T) {
	withMsg := &APIError{StatusCode: 404, Message: "task not found"}
	if !strings.Contains(withMsg.Error(), "404") || !strings.Contains(withMsg.Error(), "task not found") {
		t.Fatalf("Error() = %q, want status and message", withMsg.Error())
	}
	noMsg := &APIError{StatusCode: 500}
	if !strings.Contains(noMsg.Error(), "500") {
		t.Fatalf("Error() = %q, want status", noMsg.Error())
	}
}

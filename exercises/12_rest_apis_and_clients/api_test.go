package restapis

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	createdTitle string
	listFilter   *bool
	getID        int64
	createErr    error
	listErr      error
	getErr       error
}

func (s *fakeStore) Create(_ context.Context, title string) (Task, error) {
	s.createdTitle = title
	return Task{ID: 1, Title: title}, s.createErr
}
func (s *fakeStore) List(_ context.Context, done *bool) ([]Task, error) {
	if done != nil {
		value := *done
		s.listFilter = &value
	}
	return []Task{{ID: 1, Title: "example"}}, s.listErr
}
func (s *fakeStore) Get(_ context.Context, id int64) (Task, error) {
	s.getID = id
	return Task{ID: id, Title: "example"}, s.getErr
}

type fakeClient struct {
	url    string
	status Status
	err    error
}

func (c *fakeClient) Check(_ context.Context, url string) (Status, error) {
	c.url = url
	return c.status, c.err
}

func request(t *testing.T, handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(method, target, strings.NewReader(body)))
	return rec
}

func TestStrictJSONShapeAndValidation(t *testing.T) {
	store := &fakeStore{}
	handler := NewAPI(store, &fakeClient{}).Handler()
	for _, body := range []string{
		`{"title":""}`,
		`{"title":"ok","extra":true}`,
		`{"title":"ok"} {"title":"again"}`,
	} {
		rec := request(t, handler, http.MethodPost, "/tasks", body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("body %q status = %d, want 400", body, rec.Code)
		}
		var envelope map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil || envelope["error"] == "" {
			t.Fatalf("error envelope = %v, %v", envelope, err)
		}
	}
	rec := request(t, handler, http.MethodPost, "/tasks", `{"title":"  learn HTTP  "}`)
	if rec.Code != http.StatusCreated || store.createdTitle != "learn HTTP" {
		t.Fatalf("create status/title = %d, %q", rec.Code, store.createdTitle)
	}
}

func TestFilterAndIDParsing(t *testing.T) {
	store := &fakeStore{}
	handler := NewAPI(store, &fakeClient{}).Handler()
	if rec := request(t, handler, http.MethodGet, "/tasks?done=true", ""); rec.Code != http.StatusOK {
		t.Fatalf("filter status = %d", rec.Code)
	}
	if store.listFilter == nil || !*store.listFilter {
		t.Fatalf("filter = %v", store.listFilter)
	}
	if rec := request(t, handler, http.MethodGet, "/tasks/not-a-number", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d", rec.Code)
	}
	if rec := request(t, handler, http.MethodGet, "/tasks/23", ""); rec.Code != http.StatusOK || store.getID != 23 {
		t.Fatalf("get status/id = %d, %d", rec.Code, store.getID)
	}
	if rec := request(t, handler, http.MethodGet, "/tasks?done=perhaps", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad filter status = %d", rec.Code)
	}
}

func TestErrorMappingAndInjectedClient(t *testing.T) {
	store := &fakeStore{getErr: ErrNotFound}
	client := &fakeClient{status: Status{Status: "ready"}}
	handler := NewAPI(store, client).Handler()
	if rec := request(t, handler, http.MethodGet, "/tasks/99", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("not-found status = %d", rec.Code)
	}
	rec := request(t, handler, http.MethodPost, "/checks", `{"url":"https://service.example/health"}`)
	if rec.Code != http.StatusOK || client.url != "https://service.example/health" {
		t.Fatalf("check status/url = %d, %q", rec.Code, client.url)
	}
	client.err = ErrUpstream
	if rec := request(t, handler, http.MethodPost, "/checks", `{"url":"https://service.example/health"}`); rec.Code != http.StatusBadGateway {
		t.Fatalf("upstream status = %d", rec.Code)
	}
}

func TestHTTPStatusClientMalformedAndNonSuccess(t *testing.T) {
	malformed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":`))
	}))
	t.Cleanup(malformed.Close)
	client := &HTTPStatusClient{HTTP: malformed.Client()}
	if _, err := client.Check(context.Background(), malformed.URL); err == nil {
		t.Fatal("malformed JSON returned nil error")
	}

	failed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(failed.Close)
	client = &HTTPStatusClient{HTTP: failed.Client()}
	if _, err := client.Check(context.Background(), failed.URL); !errors.Is(err, ErrUpstream) {
		t.Fatalf("non-success error = %v, want ErrUpstream", err)
	}

	incomplete := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":""}`))
	}))
	t.Cleanup(incomplete.Close)
	client = &HTTPStatusClient{HTTP: incomplete.Client()}
	if _, err := client.Check(context.Background(), incomplete.URL); !errors.Is(err, ErrUpstream) {
		t.Fatalf("incomplete error = %v, want ErrUpstream", err)
	}
}

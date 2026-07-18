package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/api"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/m4"
)

func TestHealthAndMethods(t *testing.T) {
	handler := api.NewHandler(history.NewMemoryStore(5), []domain.Target{{Name: "catalog"}})
	recorder := m4.Request(t, handler, http.MethodGet, "/healthz", http.StatusOK)
	if recorder.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
	recorder = m4.Request(t, handler, http.MethodPost, "/healthz", http.StatusMethodNotAllowed)
	if recorder.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("Allow = %q", recorder.Header().Get("Allow"))
	}
	m4.Request(t, handler, http.MethodGet, "/missing", http.StatusNotFound)
}

func TestTargetsUseConfigurationOrder(t *testing.T) {
	store := history.NewMemoryStore(5)
	record(t, store, domain.Observation{
		Target: "a", CheckedAt: time.Unix(0, 0), Status: domain.StatusHealthy, Message: "ok",
	})
	record(t, store, domain.Observation{
		Target: "b", CheckedAt: time.Unix(1, 0), Status: domain.StatusDegraded, Message: "bad status",
	})
	handler := api.NewHandler(store, []domain.Target{{Name: "b"}, {Name: "new"}, {Name: "a"}})
	recorder := m4.Request(t, handler, http.MethodGet, "/v1/targets", http.StatusOK)
	var response domain.TargetsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Targets) != 3 ||
		response.Targets[0].Target != "b" ||
		response.Targets[1].Target != "new" ||
		response.Targets[1].Status != domain.StatusUnknown ||
		response.Targets[1].Observation != nil ||
		response.Targets[2].Target != "a" {
		t.Fatalf("response = %+v", response)
	}
}

func TestHistoryRoutesAndLimits(t *testing.T) {
	store := history.NewMemoryStore(3)
	for index, status := range []domain.Status{
		domain.StatusHealthy,
		domain.StatusDegraded,
		domain.StatusUnhealthy,
	} {
		record(t, store, domain.Observation{
			Target: "catalog", CheckedAt: time.Unix(int64(index), 0), Status: status,
		})
	}
	handler := api.NewHandler(store, []domain.Target{{Name: "catalog"}})
	recorder := m4.Request(t, handler, http.MethodGet, "/v1/history/catalog?limit=2", http.StatusOK)
	var response domain.HistoryResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Observations) != 2 ||
		response.Observations[0].Sequence != 2 ||
		response.Observations[1].Sequence != 3 {
		t.Fatalf("history = %+v", response.Observations)
	}
	recorder = m4.Request(t, handler, http.MethodGet, "/v1/history/catalog", http.StatusOK)
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Observations) != 3 {
		t.Fatalf("default history = %+v", response.Observations)
	}

	for _, target := range []string{
		"/v1/history/catalog?limit=0",
		"/v1/history/catalog?limit=4",
		"/v1/history/catalog?limit=x",
		"/v1/history/catalog?limit=1&limit=2",
		"/v1/history/catalog?limit=",
	} {
		m4.Request(t, handler, http.MethodGet, target, http.StatusBadRequest)
	}
	m4.Request(t, handler, http.MethodGet, "/v1/history/missing", http.StatusNotFound)
	m4.Request(t, handler, http.MethodGet, "/v1/history/", http.StatusNotFound)
	recorder = m4.Request(t, handler, http.MethodDelete, "/v1/history/catalog", http.StatusMethodNotAllowed)
	if recorder.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("Allow = %q", recorder.Header().Get("Allow"))
	}
}

func TestStoppingState(t *testing.T) {
	state := api.NewState()
	handler := api.NewHandlerWithOptions(history.NewMemoryStore(2), []domain.Target{{Name: "catalog"}}, api.Options{
		HistoryLimit: 2,
		State:        state,
	})
	if state.Stopping() {
		t.Fatal("new state is stopping")
	}
	state.Stop()
	if !state.Stopping() {
		t.Fatal("state did not stop")
	}
	recorder := m4.Request(t, handler, http.MethodGet, "/healthz", http.StatusServiceUnavailable)
	if recorder.Body.String() != "{\"status\":\"stopping\"}\n" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
	m4.Request(t, handler, http.MethodGet, "/v1/targets", http.StatusServiceUnavailable)
	m4.Request(t, handler, http.MethodGet, "/v1/history/catalog", http.StatusServiceUnavailable)
}

func TestInternalHistoryErrorsAndLogging(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := api.NewHandlerWithOptions(errorStore{}, []domain.Target{{Name: "catalog"}}, api.Options{
		HistoryLimit: 2,
		Logger:       logger,
	})
	m4.Request(t, handler, http.MethodGet, "/v1/history/catalog", http.StatusInternalServerError)
	if !strings.Contains(logs.String(), `"status":500`) || !strings.Contains(logs.String(), `"path":"/v1/history/catalog"`) {
		t.Fatalf("logs = %q", logs.String())
	}
	m4.Request(t, api.NewHandler(nil, nil), http.MethodGet, "/v1/targets", http.StatusInternalServerError)
}

func TestResponseEncodingFailureIsLogged(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := api.NewHandlerWithOptions(history.NewMemoryStore(2), nil, api.Options{
		HistoryLimit: 2,
		Logger:       logger,
	})
	writer := &failingResponseWriter{header: make(http.Header)}
	handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if writer.status != http.StatusOK {
		t.Fatalf("status = %d, want %d", writer.status, http.StatusOK)
	}
	if !strings.Contains(logs.String(), `"msg":"encode HTTP response"`) ||
		!strings.Contains(logs.String(), `"error":"fixture response write failure"`) {
		t.Fatalf("logs = %q", logs.String())
	}
}

func record(t *testing.T, store *history.MemoryStore, observation domain.Observation) {
	t.Helper()
	if err := store.Record(observation); err != nil {
		t.Fatal(err)
	}
}

type errorStore struct{}

func (errorStore) Record(domain.Observation) error {
	return nil
}

func (errorStore) Current() []domain.Observation {
	return nil
}

func (errorStore) History(string, int) ([]domain.Observation, error) {
	return nil, errors.New("fixture history failure")
}

type failingResponseWriter struct {
	header http.Header
	status int
}

func (writer *failingResponseWriter) Header() http.Header {
	return writer.header
}

func (writer *failingResponseWriter) WriteHeader(status int) {
	writer.status = status
}

func (*failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("fixture response write failure")
}

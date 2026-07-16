package m5

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/api"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m3"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m4"
)

type ServerFactory = m4.ServerFactory

func AssertServerContract(t *testing.T, factory ServerFactory) {
	t.Helper()
	m4.AssertServerContract(t, factory)

	t.Run("framework defaults are normalized", func(t *testing.T) {
		handler := factory(noopService{}, discardLogger())
		methodCases := []struct {
			method string
			path   string
			allow  string
		}{
			{http.MethodHead, "/health", "GET"},
			{http.MethodOptions, "/health", "GET"},
			{http.MethodHead, "/tasks", "GET, POST"},
			{http.MethodOptions, "/tasks", "GET, POST"},
			{http.MethodHead, "/tasks/1", "GET, PATCH, DELETE"},
			{http.MethodOptions, "/tasks/1", "GET, PATCH, DELETE"},
		}
		for _, test := range methodCases {
			response := serve(handler, test.method, test.path)
			if response.Header().Get("Allow") != test.allow {
				t.Fatalf("%s %s Allow = %q, want %q",
					test.method, test.path, response.Header().Get("Allow"), test.allow)
			}
			m3.AssertErrorResponse(t, response.Code, response.Header(), response.Body.Bytes(),
				http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed for this path", "")
		}

		for _, requestPath := range []string{
			"/Health",
			"/tasks/",
			"/tasks//",
			"/tasks/1/",
			"/tasks//1",
			"/TASKS",
			"/missing/",
		} {
			response := serve(handler, http.MethodGet, requestPath)
			if location := response.Header().Get("Location"); location != "" {
				t.Fatalf("GET %s redirected to %q", requestPath, location)
			}
			m3.AssertErrorResponse(t, response.Code, response.Header(), response.Body.Bytes(),
				http.StatusNotFound, "not_found", "route was not found", "")
		}

		response := serve(handler, http.MethodOptions, "/missing")
		m3.AssertErrorResponse(t, response.Code, response.Header(), response.Body.Bytes(),
			http.StatusNotFound, "not_found", "route was not found", "")
	})

	t.Run("panics are logged and sanitized", func(t *testing.T) {
		var logs bytes.Buffer
		handler := factory(panicService{}, slog.New(slog.NewTextHandler(&logs, nil)))
		response := serve(handler, http.MethodGet, "/tasks")
		m3.AssertErrorResponse(t, response.Code, response.Header(), response.Body.Bytes(),
			http.StatusInternalServerError, "internal_error", "the server could not complete the request", "")
		if strings.Contains(response.Body.String(), "private panic detail") {
			t.Fatalf("panic leaked in response: %q", response.Body.String())
		}
		if !strings.Contains(logs.String(), "private panic detail") ||
			!strings.Contains(logs.String(), "task HTTP handler panicked") {
			t.Fatalf("panic was not logged with internal detail: %q", logs.String())
		}
	})
}

type noopService struct{}

func (noopService) Create(context.Context, task.CreateInput) (task.Task, error) {
	return task.Task{ID: 1, Title: "created"}, nil
}

func (noopService) List(context.Context, task.ListFilter) ([]task.Task, error) {
	return []task.Task{}, nil
}

func (noopService) Get(context.Context, int64) (task.Task, error) {
	return task.Task{ID: 1, Title: "found"}, nil
}

func (noopService) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	return task.Task{ID: 1, Title: "updated"}, nil
}

func (noopService) Delete(context.Context, int64) error {
	return nil
}

type panicService struct{}

func (panicService) Create(context.Context, task.CreateInput) (task.Task, error) {
	panic("private panic detail")
}

func (panicService) List(context.Context, task.ListFilter) ([]task.Task, error) {
	panic("private panic detail")
}

func (panicService) Get(context.Context, int64) (task.Task, error) {
	panic("private panic detail")
}

func (panicService) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	panic("private panic detail")
}

func (panicService) Delete(context.Context, int64) error {
	panic("private panic detail")
}

func serve(handler http.Handler, method, requestPath string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, requestPath, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var _ api.Service = noopService{}
var _ api.Service = panicService{}

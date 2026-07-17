package nethttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m3"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m5"
)

func TestMilestone5HTTPContract(t *testing.T) {
	m5.AssertServerContract(t, func(service api.Service, logger *slog.Logger) http.Handler {
		return nethttp.New(service, logger)
	})
}

type memoryService struct {
	nextID int64
	tasks  []task.Task
	err    error
}

func (service *memoryService) Create(_ context.Context, input task.CreateInput) (task.Task, error) {
	if service.err != nil {
		return task.Task{}, service.err
	}
	service.nextID++
	value := task.Task{ID: service.nextID, Title: input.Title}
	service.tasks = append(service.tasks, value)
	return value, nil
}

func (service *memoryService) List(_ context.Context, filter task.ListFilter) ([]task.Task, error) {
	if service.err != nil {
		return nil, service.err
	}
	result := make([]task.Task, 0)
	for _, value := range service.tasks {
		if filter.Completed == nil || value.Completed == *filter.Completed {
			result = append(result, value)
		}
	}
	return result, nil
}

func (service *memoryService) Get(_ context.Context, id int64) (task.Task, error) {
	if service.err != nil {
		return task.Task{}, service.err
	}
	for _, value := range service.tasks {
		if value.ID == id {
			return value, nil
		}
	}
	return task.Task{}, task.NewNotFoundError(id)
}

func (service *memoryService) Update(_ context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	for index := range service.tasks {
		if service.tasks[index].ID == id {
			if input.Title != nil {
				service.tasks[index].Title = *input.Title
			}
			if input.Completed != nil {
				service.tasks[index].Completed = *input.Completed
			}
			return service.tasks[index], nil
		}
	}
	return task.Task{}, task.NewNotFoundError(id)
}

func (service *memoryService) Delete(_ context.Context, id int64) error {
	for index := range service.tasks {
		if service.tasks[index].ID == id {
			service.tasks = append(service.tasks[:index], service.tasks[index+1:]...)
			return nil
		}
	}
	return task.NewNotFoundError(id)
}

func TestHTTPCRUDFilterAndNoContent(t *testing.T) {
	handler := nethttp.New(&memoryService{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	assertResponse(t, handler, http.MethodGet, "/health", nil, "", 200, `{"status":"ok"}`)
	assertResponse(t, handler, http.MethodGet, "/tasks", nil, "", 200, `[]`)
	assertResponse(t, handler, http.MethodPost, "/tasks", []byte(`{"title":"  Learn REST 🐍  "}`),
		"application/json", 201, `{"id":1,"title":"Learn REST 🐍","completed":false}`)
	assertResponse(t, handler, http.MethodGet, "/tasks?completed=false", nil, "", 200,
		`[{"id":1,"title":"Learn REST 🐍","completed":false}]`)
	assertResponse(t, handler, http.MethodPatch, "/tasks/1",
		[]byte(`{"title":"Read OpenAPI","completed":true}`), "application/json", 200,
		`{"id":1,"title":"Read OpenAPI","completed":true}`)
	response := serve(handler, http.MethodDelete, "/tasks/1", nil, "")
	if response.Code != http.StatusNoContent || response.Body.Len() != 0 ||
		response.Header().Get("Content-Type") != "" {
		t.Fatalf("DELETE response = %d %q %q", response.Code, response.Body.String(), response.Header().Get("Content-Type"))
	}
	assertError(t, handler, http.MethodGet, "/tasks/1", nil, "", 404, "not_found", "task 1 was not found", "")
}

func TestHTTPBodyValidationContract(t *testing.T) {
	handler := nethttp.New(&memoryService{}, slog.Default())
	cases := []struct {
		name, method, path, body, contentType string
		status                                int
		code, message, field                  string
	}{
		{"missing content type", "POST", "/tasks", `{}`, "", 400, "invalid_json", "request Content-Type must be application/json", ""},
		{"wrong content type", "POST", "/tasks", `{}`, "text/plain", 400, "invalid_json", "request Content-Type must be application/json", ""},
		{"wrong charset", "POST", "/tasks", `{}`, "application/json; charset=iso-8859-1", 400, "invalid_json", "request JSON charset must be UTF-8", ""},
		{"invalid UTF-8", "POST", "/tasks", string([]byte{0xff}), "application/json", 400, "invalid_json", "request body must be valid JSON", ""},
		{"malformed", "POST", "/tasks", `{`, "application/json", 400, "invalid_json", "request body must be valid JSON", ""},
		{"duplicate", "POST", "/tasks", `{"title":"a","title":"b"}`, "application/json", 400, "invalid_json", "request body must be valid JSON", ""},
		{"trailing", "POST", "/tasks", `{"title":"a"} {}`, "application/json", 400, "invalid_json", "request body must be valid JSON", ""},
		{"constant", "POST", "/tasks", `{"title":NaN}`, "application/json", 400, "invalid_json", "request body must be valid JSON", ""},
		{"shape", "POST", "/tasks", `[]`, "application/json", 422, "validation_error", "request body must be a JSON object", "body"},
		{"missing", "POST", "/tasks", `{}`, "application/json", 422, "validation_error", "missing property: title", "title"},
		{"unknown", "POST", "/tasks", `{"title":"x","done":false}`, "application/json", 422, "validation_error", "unknown property: done", "done"},
		{"null", "POST", "/tasks", `{"title":null}`, "application/json", 422, "validation_error", "title must be a string", "title"},
		{"wrong title type", "POST", "/tasks", `{"title":7}`, "application/json", 422, "validation_error", "title must be a string", "title"},
		{"empty title", "POST", "/tasks", `{"title":" "}`, "application/json", 422, "validation_error", "title must contain between 1 and 120 characters", "title"},
		{"multiline", "POST", "/tasks", `{"title":"first\nsecond"}`, "application/json", 422, "validation_error", "title must occupy one physical line", "title"},
		{"control", "POST", "/tasks", `{"title":"control\u0000"}`, "application/json", 422, "validation_error", "title must not contain control characters", "title"},
		{"empty update", "PATCH", "/tasks/1", `{}`, "application/json", 422, "validation_error", "update must include title or completed", "update"},
		{"null completed", "PATCH", "/tasks/1", `{"completed":null}`, "application/json", 422, "validation_error", "completed must be a Boolean", "completed"},
		{"wrong completed", "PATCH", "/tasks/1", `{"completed":0}`, "application/json", 422, "validation_error", "completed must be a Boolean", "completed"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			assertError(t, handler, test.method, test.path, []byte(test.body), test.contentType,
				test.status, test.code, test.message, test.field)
		})
	}
}

func TestHTTPQueryPathMethodAndRouteContract(t *testing.T) {
	handler := nethttp.New(&memoryService{}, slog.Default())
	requestCases := []struct{ path, message, field string }{
		{"/tasks?completed=True", "completed filter must be true or false", "completed"},
		{"/tasks?completed=true&completed=false", "completed filter must be true or false", "completed"},
		{"/tasks?other=true", "unknown query parameter: other", "other"},
		{"/health?verbose=true", "unknown query parameter: verbose", "verbose"},
		{"/tasks/0", "task ID must be a positive integer", "id"},
		{"/tasks/+1", "task ID must be a positive integer", "id"},
		{"/tasks/%D9%A1", "task ID must be a positive integer", "id"},
	}
	for _, test := range requestCases {
		assertError(t, handler, "GET", test.path, nil, "", 422, "validation_error", test.message, test.field)
	}
	methodCases := []struct{ method, path, allow string }{
		{"POST", "/health", "GET"}, {"PUT", "/tasks", "GET, POST"}, {"POST", "/tasks/1", "GET, PATCH, DELETE"},
		{"HEAD", "/tasks", "GET, POST"}, {"OPTIONS", "/tasks/1", "GET, PATCH, DELETE"},
	}
	for _, test := range methodCases {
		response := serve(handler, test.method, test.path, nil, "")
		if response.Header().Get("Allow") != test.allow {
			t.Fatalf("%s %s Allow = %q", test.method, test.path, response.Header().Get("Allow"))
		}
		assertErrorRecorder(t, response, 405, "method_not_allowed", "method is not allowed for this path", "")
	}
	for _, path := range []string{
		"/missing", "/tasks/", "/tasks//", "/tasks/../tasks", "/tasks/1/extra", "/docs", "/openapi.json",
	} {
		assertError(t, handler, "GET", path, nil, "", 404, "not_found", "route was not found", "")
	}
}

func TestHTTPInternalErrorsAreLoggedAndSanitized(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := nethttp.New(&memoryService{err: errors.New("private storage detail")}, logger)
	response := serve(handler, "GET", "/tasks", nil, "")
	assertErrorRecorder(t, response, 500, "internal_error", "the server could not complete the request", "")
	if strings.Contains(response.Body.String(), "private") || !strings.Contains(logs.String(), "private storage detail") {
		t.Fatalf("response/log sanitization failed: body=%q logs=%q", response.Body.String(), logs.String())
	}
}

func serve(handler http.Handler, method, path string, body []byte, contentType string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertResponse(t *testing.T, handler http.Handler, method, path string, body []byte, contentType string, status int, expected string) {
	t.Helper()
	response := serve(handler, method, path, body, contentType)
	if response.Code != status {
		t.Fatalf("%s %s status = %d, body=%s", method, path, response.Code, response.Body.String())
	}
	var actualValue, expectedValue any
	if json.Unmarshal(response.Body.Bytes(), &actualValue) != nil || json.Unmarshal([]byte(expected), &expectedValue) != nil {
		t.Fatal("invalid test JSON")
	}
	actual, _ := json.Marshal(actualValue)
	want, _ := json.Marshal(expectedValue)
	if !bytes.Equal(actual, want) {
		t.Fatalf("%s %s body = %s, want %s", method, path, actual, want)
	}
}

func assertError(t *testing.T, handler http.Handler, method, path string, body []byte, contentType string, status int, code, message, field string) {
	t.Helper()
	assertErrorRecorder(t, serve(handler, method, path, body, contentType), status, code, message, field)
}

func assertErrorRecorder(t *testing.T, response *httptest.ResponseRecorder, status int, code, message, field string) {
	t.Helper()
	m3.AssertErrorResponse(t, response.Code, response.Header(), response.Body.Bytes(), status, code, message, field)
}

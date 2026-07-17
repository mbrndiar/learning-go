package m5_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api"
	apichi "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/chi"
	apigin "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/gin"
	apinethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

func TestOpenAPIDocumentIsLocalValidAndComplete(t *testing.T) {
	document, content := loadOpenAPI(t)
	// The course must remain runnable offline, so the contract cannot depend on
	// schemas fetched from a network during validation.
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "$ref:") && !strings.Contains(line, `"#/`) {
			t.Fatalf("external OpenAPI reference is not allowed: %s", line)
		}
	}

	expected := []struct {
		path     string
		method   string
		statuses []int
		body     bool
	}{
		{"/health", http.MethodGet, []int{200, 405, 500}, false},
		{"/tasks", http.MethodPost, []int{201, 400, 405, 422, 500}, true},
		{"/tasks", http.MethodGet, []int{200, 405, 422, 500}, false},
		{"/tasks/{taskId}", http.MethodGet, []int{200, 404, 405, 422, 500}, false},
		{"/tasks/{taskId}", http.MethodPatch, []int{200, 400, 404, 405, 422, 500}, true},
		{"/tasks/{taskId}", http.MethodDelete, []int{204, 404, 405, 422, 500}, false},
	}
	if document.Paths.Len() != 3 {
		t.Fatalf("path count = %d, want 3", document.Paths.Len())
	}
	for path, count := range map[string]int{
		"/health": 1, "/tasks": 2, "/tasks/{taskId}": 3,
	} {
		pathItem := document.Paths.Value(path)
		if pathItem == nil {
			t.Fatalf("missing path %s", path)
		}
		if operations := pathItem.Operations(); len(operations) != count {
			t.Fatalf("%s operation count = %d, want %d", path, len(operations), count)
		}
	}
	for _, expectedOperation := range expected {
		pathItem := document.Paths.Value(expectedOperation.path)
		if pathItem == nil {
			t.Fatalf("missing path %s", expectedOperation.path)
		}
		operation := pathItem.GetOperation(expectedOperation.method)
		if operation == nil {
			t.Fatalf("missing operation %s %s", expectedOperation.method, expectedOperation.path)
		}
		if expectedOperation.body {
			if operation.RequestBody == nil || operation.RequestBody.Value == nil ||
				!operation.RequestBody.Value.Required ||
				operation.RequestBody.Value.Content.Get("application/json") == nil {
				t.Fatalf("%s %s does not require a JSON body",
					expectedOperation.method, expectedOperation.path)
			}
		}
		if operation.Responses.Len() != len(expectedOperation.statuses) {
			t.Fatalf("%s %s response count = %d, want %d",
				expectedOperation.method, expectedOperation.path,
				operation.Responses.Len(), len(expectedOperation.statuses))
		}
		for _, status := range expectedOperation.statuses {
			response := operation.Responses.Status(status)
			if response == nil || response.Value == nil {
				t.Fatalf("%s %s is missing status %d",
					expectedOperation.method, expectedOperation.path, status)
			}
			mediaType := response.Value.Content.Get("application/json")
			if status == http.StatusNoContent {
				if len(response.Value.Content) != 0 {
					t.Fatalf("%s %s status 204 declares content",
						expectedOperation.method, expectedOperation.path)
				}
			} else if mediaType == nil || mediaType.Schema == nil {
				t.Fatalf("%s %s status %d lacks a JSON schema",
					expectedOperation.method, expectedOperation.path, status)
			}
		}
	}

	for _, name := range []string{
		"InvalidJson", "NotFound", "MethodNotAllowed", "ValidationError", "InternalError",
	} {
		response := document.Components.Responses[name]
		if response == nil || response.Value == nil ||
			response.Value.Content.Get("application/json") == nil {
			t.Fatalf("component response %s is missing JSON content", name)
		}
	}
}

func TestRepresentativeServerTrafficValidatesAgainstOpenAPI(t *testing.T) {
	document, _ := loadOpenAPI(t)
	router, err := legacy.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	factories := []struct {
		name string
		new  func(api.Service, *slog.Logger) http.Handler
	}{
		{"nethttp", apinethttp.New},
		{"chi", apichi.New},
		{"gin", apigin.New},
	}
	// Validate the same exchanges through every adapter so a framework-specific
	// response cannot drift from the published contract unnoticed.
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			handler := factory.new(&memoryService{}, logger)
			validateExchange(t, router, handler, http.MethodGet, "/health", nil, "")
			validateExchange(t, router, handler, http.MethodPost, "/tasks",
				[]byte(`{"title":"OpenAPI traffic"}`), "application/json")
			validateExchange(t, router, handler, http.MethodGet, "/tasks/1", nil, "")
			validateExchange(t, router, handler, http.MethodDelete, "/tasks/1", nil, "")
		})
	}
}

func loadOpenAPI(t *testing.T) (*openapi3.T, []byte) {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate OpenAPI test source")
	}
	path := filepath.Join(filepath.Dir(source), "..", "..", "docs", "openapi.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	loader := openapi3.NewLoader()
	loader.Context = ctx
	loader.IsExternalRefsAllowed = false
	document, err := loader.LoadFromData(content)
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(ctx); err != nil {
		t.Fatal(err)
	}
	return document, content
}

func validateExchange(
	t *testing.T,
	router routers.Router,
	handler http.Handler,
	method string,
	requestPath string,
	body []byte,
	contentType string,
) {
	t.Helper()
	request := httptest.NewRequest(method, "http://127.0.0.1:8000"+requestPath, bytes.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	route, pathParams, err := router.FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	requestInput := &openapi3filter.RequestValidationInput{
		Request: request, PathParams: pathParams, Route: route,
	}
	if err := openapi3filter.ValidateRequest(request.Context(), requestInput); err != nil {
		t.Fatalf("request validation: %v", err)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	result := response.Result()
	defer result.Body.Close()
	responseInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestInput,
		Status:                 result.StatusCode,
		Header:                 result.Header,
		Body:                   result.Body,
	}
	if err := openapi3filter.ValidateResponse(request.Context(), responseInput); err != nil {
		t.Fatalf("response validation: %v", err)
	}
}

type memoryService struct {
	nextID int64
	tasks  []task.Task
}

func (service *memoryService) Create(_ context.Context, input task.CreateInput) (task.Task, error) {
	service.nextID++
	value := task.Task{ID: service.nextID, Title: input.Title}
	service.tasks = append(service.tasks, value)
	return value, nil
}

func (service *memoryService) List(context.Context, task.ListFilter) ([]task.Task, error) {
	return append([]task.Task(nil), service.tasks...), nil
}

func (service *memoryService) Get(_ context.Context, id int64) (task.Task, error) {
	for _, value := range service.tasks {
		if value.ID == id {
			return value, nil
		}
	}
	return task.Task{}, task.NewNotFoundError(id)
}

func (service *memoryService) Update(
	_ context.Context,
	id int64,
	input task.UpdateInput,
) (task.Task, error) {
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

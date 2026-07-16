package chi_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/api/chi"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

func TestPlaceholderIsStableAndHasNoServiceSideEffects(t *testing.T) {
	service := &countingService{}
	handler := chi.New(service, slog.Default())
	for _, request := range []*httptest.ResponseRecorder{
		serve(handler, "GET", "/health"),
		serve(handler, "POST", "/tasks"),
		serve(handler, "GET", "/missing"),
	} {
		if request.Code != 501 ||
			request.Body.String() != `{"error":{"code":"not_implemented","message":"this endpoint is not implemented"}}`+"\n" {
			t.Fatalf("response = %d %q", request.Code, request.Body.String())
		}
	}
	if service.calls != 0 {
		t.Fatalf("service calls = %d", service.calls)
	}
}

type countingService struct{ calls int }

func (service *countingService) Create(context.Context, task.CreateInput) (task.Task, error) {
	service.calls++
	return task.Task{}, errors.New("unexpected call")
}
func (service *countingService) List(context.Context, task.ListFilter) ([]task.Task, error) {
	service.calls++
	return nil, errors.New("unexpected call")
}
func (service *countingService) Get(context.Context, int64) (task.Task, error) {
	service.calls++
	return task.Task{}, errors.New("unexpected call")
}
func (service *countingService) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	service.calls++
	return task.Task{}, errors.New("unexpected call")
}
func (service *countingService) Delete(context.Context, int64) error {
	service.calls++
	return errors.New("unexpected call")
}

func serve(handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}, method, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(nil))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

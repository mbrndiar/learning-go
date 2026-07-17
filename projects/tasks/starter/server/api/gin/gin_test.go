package gin_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/server/api/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

func TestStarterGinIsInert(t *testing.T) {
	handler := gin.New(panicService{}, nil)
	for _, test := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
		{http.MethodGet, "/tasks"},
		{http.MethodPost, "/tasks"},
		{http.MethodPatch, "/tasks/1"},
		{http.MethodDelete, "/tasks/1"},
		{http.MethodGet, "/missing"},
		{http.MethodHead, "/tasks"},
	} {
		request := httptest.NewRequest(test.method, test.path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNotImplemented ||
			strings.TrimSpace(response.Body.String()) !=
				`{"error":{"code":"not_implemented","message":"this endpoint is not implemented"}}` {
			t.Fatalf("%s %s response = %d %q",
				test.method, test.path, response.Code, response.Body.String())
		}
	}
}

type panicService struct{}

func (panicService) Create(context.Context, task.CreateInput) (task.Task, error) {
	panic("starter called service")
}

func (panicService) List(context.Context, task.ListFilter) ([]task.Task, error) {
	panic("starter called service")
}

func (panicService) Get(context.Context, int64) (task.Task, error) {
	panic("starter called service")
}

func (panicService) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	panic("starter called service")
}

func (panicService) Delete(context.Context, int64) error {
	panic("starter called service")
}

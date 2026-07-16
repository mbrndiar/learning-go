package gin

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ginlib "github.com/gin-gonic/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

func TestRecoveryDoesNotDoubleWriteCommittedResponse(t *testing.T) {
	var logs bytes.Buffer
	handler := New(recoveryService{}, slog.New(slog.NewTextHandler(&logs, nil))).(*Handler)
	handler.engine.GET("/panic-after-write", func(context *ginlib.Context) {
		context.String(http.StatusAccepted, "partial response")
		panic("panic after committed response")
	})

	request := httptest.NewRequest(http.MethodGet, "/panic-after-write", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || response.Body.String() != "partial response" {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "internal_error") ||
		!strings.Contains(logs.String(), "panic after committed response") {
		t.Fatalf("response/log = %q / %q", response.Body.String(), logs.String())
	}
}

type recoveryService struct{}

func (recoveryService) Create(context.Context, task.CreateInput) (task.Task, error) {
	return task.Task{}, nil
}

func (recoveryService) List(context.Context, task.ListFilter) ([]task.Task, error) {
	return nil, nil
}

func (recoveryService) Get(context.Context, int64) (task.Task, error) {
	return task.Task{}, nil
}

func (recoveryService) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	return task.Task{}, nil
}

func (recoveryService) Delete(context.Context, int64) error {
	return nil
}

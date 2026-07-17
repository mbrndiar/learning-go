package gin_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	ginlib "github.com/gin-gonic/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m5"
)

func TestMilestone5HTTPContract(t *testing.T) {
	m5.AssertServerContract(t, func(service api.Service, logger *slog.Logger) http.Handler {
		return gin.New(service, logger)
	})
}

func TestNilLoggerUsesDefaultWithoutChangingGinGlobals(t *testing.T) {
	mode := ginlib.Mode()
	defaultWriter := ginlib.DefaultWriter
	defaultErrorWriter := ginlib.DefaultErrorWriter
	handler := gin.New(nilService{}, nil)
	if handler == nil {
		t.Fatal("New returned nil")
	}
	if ginlib.Mode() != mode || ginlib.DefaultWriter != defaultWriter ||
		ginlib.DefaultErrorWriter != defaultErrorWriter {
		t.Fatal("New changed Gin process globals")
	}
}

type nilService struct{}

func (nilService) Create(context.Context, task.CreateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

func (nilService) List(context.Context, task.ListFilter) ([]task.Task, error) {
	return nil, task.ErrNotImplemented
}

func (nilService) Get(context.Context, int64) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

func (nilService) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

func (nilService) Delete(context.Context, int64) error {
	return task.ErrNotImplemented
}

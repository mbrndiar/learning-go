// Package api defines transport-neutral HTTP boundary policy for Task routers.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

const MaxBodyBytes = 1 << 20

type Service interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context, task.ListFilter) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Delete(context.Context, int64) error
}

type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error Error `json:"error"`
}

type HTTPError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
}

func (e *HTTPError) Error() string {
	if e == nil {
		return task.ErrNotImplemented.Error()
	}
	return e.Message
}

func TaskDTO(value task.Task) Task {
	return Task{ID: value.ID, Title: value.Title, Completed: value.Completed}
}

func TaskDTOs(values []task.Task) []Task {
	return nil
}

func ValidateNoQuery(query url.Values) *HTTPError {
	return notImplemented()
}

func ParseListFilter(query url.Values) (task.ListFilter, *HTTPError) {
	return task.ListFilter{}, notImplemented()
}

func ParseID(raw string) (int64, *HTTPError) {
	return 0, notImplemented()
}

func DecodeCreate(request *http.Request) (task.CreateInput, *HTTPError) {
	return task.CreateInput{}, notImplemented()
}

func DecodeUpdate(request *http.Request) (task.UpdateInput, *HTTPError) {
	return task.UpdateInput{}, notImplemented()
}

func WriteJSON(writer http.ResponseWriter, status int, value any) {
	WriteError(writer, notImplemented())
}

func WriteError(writer http.ResponseWriter, boundaryError *HTTPError) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusNotImplemented)
	_, _ = writer.Write([]byte(`{"error":{"code":"not_implemented","message":"this endpoint is not implemented"}}` + "\n"))
}

func MapError(err error, logger *slog.Logger) *HTTPError {
	return notImplemented()
}

func MethodNotAllowed(allow string) *HTTPError {
	return notImplemented()
}

func RouteNotFound() *HTTPError {
	return notImplemented()
}

func notImplemented() *HTTPError {
	return &HTTPError{
		Status: http.StatusNotImplemented, Code: "not_implemented",
		Message: "this endpoint is not implemented",
	}
}

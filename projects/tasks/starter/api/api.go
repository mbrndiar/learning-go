// Package api defines transport-neutral HTTP boundary policy for Task routers.
// The wire-level source of truth remains projects/tasks/docs/SPEC.md and
// projects/tasks/docs/openapi.yaml.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// MaxBodyBytes bounds how much of a request body a boundary decoder will read.
const MaxBodyBytes = 1 << 20

// Service is the application capability every router adapter must call to
// fulfill a routed request; it is satisfied by task.Service.
type Service interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context, task.ListFilter) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Delete(context.Context, int64) error
}

// Task is the wire representation of a task.Task in success responses.
type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// Error is the machine-readable body carried inside ErrorEnvelope.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorEnvelope is the JSON shape every documented error response uses.
type ErrorEnvelope struct {
	Error Error `json:"error"`
}

// HTTPError carries the status and error body an adapter must write for a
// boundary failure. Adapters own writing it; they must not invent their own
// error shapes.
type HTTPError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
}

// Error implements error.
func (e *HTTPError) Error() string {
	if e == nil {
		return task.ErrNotImplemented.Error()
	}
	return e.Message
}

// TaskDTO converts one domain task.Task into its wire representation.
func TaskDTO(value task.Task) Task {
	return Task{ID: value.ID, Title: value.Title, Completed: value.Completed}
}

// TaskDTOs converts a slice of domain tasks into their wire representation,
// preserving order.
func TaskDTOs(values []task.Task) []Task {
	return nil
}

// ValidateNoQuery rejects any query parameter on endpoints that accept none.
func ValidateNoQuery(query url.Values) *HTTPError {
	return notImplemented()
}

// ParseListFilter validates and converts the List endpoint's query parameters
// into a domain ListFilter.
func ParseListFilter(query url.Values) (task.ListFilter, *HTTPError) {
	return task.ListFilter{}, notImplemented()
}

// ParseID validates a path segment as a positive task ID.
func ParseID(raw string) (int64, *HTTPError) {
	return 0, notImplemented()
}

// DecodeCreate validates content type and JSON shape, then produces a domain
// CreateInput. It rejects unknown properties and values of the wrong type.
func DecodeCreate(request *http.Request) (task.CreateInput, *HTTPError) {
	return task.CreateInput{}, notImplemented()
}

// DecodeUpdate validates content type and JSON shape, then produces a domain
// UpdateInput. It rejects unknown properties and values of the wrong type.
func DecodeUpdate(request *http.Request) (task.UpdateInput, *HTTPError) {
	return task.UpdateInput{}, notImplemented()
}

// WriteJSON writes value as a JSON response with the given status and the
// shared content type.
func WriteJSON(writer http.ResponseWriter, status int, value any) {
	WriteError(writer, notImplemented())
}

// WriteError writes boundaryError as the shared ErrorEnvelope with its HTTP
// status. Adapters must route every failure through this instead of writing
// their own error body.
func WriteError(writer http.ResponseWriter, boundaryError *HTTPError) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusNotImplemented)
	_, _ = writer.Write([]byte(`{"error":{"code":"not_implemented","message":"this endpoint is not implemented"}}` + "\n"))
}

// MapError classifies a domain or service error into the HTTPError an
// adapter must return, logging unexpected failures without exposing them to
// the caller.
func MapError(err error, logger *slog.Logger) *HTTPError {
	return notImplemented()
}

// MethodNotAllowed builds the 405 error for a known path whose method is
// unsupported. The adapter remains responsible for writing allow to the
// response's Allow header.
func MethodNotAllowed(allow string) *HTTPError {
	return notImplemented()
}

// RouteNotFound builds the shared 404 error for an unmatched path.
func RouteNotFound() *HTTPError {
	return notImplemented()
}

func notImplemented() *HTTPError {
	return &HTTPError{
		Status: http.StatusNotImplemented, Code: "not_implemented",
		Message: "this endpoint is not implemented",
	}
}

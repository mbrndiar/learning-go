package api

import (
	"log/slog"
	"net/http"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

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

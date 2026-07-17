package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// HTTPError pairs an HTTP status with a machine-readable code and message so
// adapters can render any failure (decode, validation, domain, or panic)
// through the one WriteError path without knowing why it occurred.
type HTTPError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
}

// Error implements the error interface. It tolerates a nil receiver so a
// nil *HTTPError can still be passed through error-returning code paths.
func (e *HTTPError) Error() string {
	if e == nil {
		return "HTTP boundary error"
	}
	return e.Message
}

// WriteJSON encodes value as the JSON response body with the given status,
// centralizing the response Content-Type so every adapter's success and
// error paths look identical on the wire.
func WriteJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

// WriteError renders a boundary error as the shared error envelope. A nil
// boundaryError becomes a generic 500, which lets adapter-level panic
// recovery call WriteError(writer, nil) without constructing an HTTPError.
func WriteError(writer http.ResponseWriter, boundaryError *HTTPError) {
	if boundaryError == nil {
		boundaryError = &HTTPError{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "the server could not complete the request",
		}
	}
	WriteJSON(writer, boundaryError.Status, ErrorEnvelope{Error: Error{
		Code: boundaryError.Code, Message: boundaryError.Message, Details: boundaryError.Details,
	}})
}

// MapError translates a domain/service error into the HTTPError an adapter
// should render. Unrecognized errors are logged (if logger is non-nil) with
// their original detail and reported to the client only as a sanitized 500,
// so internals never leak through the wire.
func MapError(err error, logger *slog.Logger) *HTTPError {
	var validationError *task.ValidationError
	switch {
	case errors.As(err, &validationError):
		return validation(validationError.Field, validationError.Message)
	case errors.Is(err, task.ErrNotFound):
		var notFoundError *task.NotFoundError
		if errors.As(err, &notFoundError) {
			return &HTTPError{Status: http.StatusNotFound, Code: "not_found", Message: notFoundError.Error()}
		}
		return &HTTPError{Status: http.StatusNotFound, Code: "not_found", Message: "task was not found"}
	case errors.Is(err, task.ErrNotImplemented):
		return &HTTPError{Status: http.StatusNotImplemented, Code: "not_implemented", Message: "this endpoint is not implemented"}
	default:
		if logger != nil {
			logger.Error("task HTTP request failed", "error", err)
		}
		return &HTTPError{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "the server could not complete the request",
		}
	}
}

// MethodNotAllowed builds the shared 405 error for a known path with an
// unsupported method. allow is documentary only: it does not set the
// response Allow header, so callers must still set that header themselves
// before calling WriteError.
func MethodNotAllowed(allow string) *HTTPError {
	return &HTTPError{
		Status: http.StatusMethodNotAllowed, Code: "method_not_allowed",
		Message: "method is not allowed for this path",
	}
}

// RouteNotFound builds the shared 404 error for an unknown route.
func RouteNotFound() *HTTPError {
	return &HTTPError{Status: http.StatusNotFound, Code: "not_found", Message: "route was not found"}
}

func validation(field, message string) *HTTPError {
	return &HTTPError{
		Status: http.StatusUnprocessableEntity, Code: "validation_error", Message: message,
		Details: map[string]any{"field": field},
	}
}

func invalidJSON(message string) *HTTPError {
	return &HTTPError{Status: http.StatusBadRequest, Code: "invalid_json", Message: message}
}

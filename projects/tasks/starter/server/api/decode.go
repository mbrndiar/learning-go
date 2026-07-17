package api

import (
	"net/http"
	"net/url"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// MaxBodyBytes bounds how much of a request body a boundary decoder will read.
const MaxBodyBytes = 1 << 20

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

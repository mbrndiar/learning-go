// Package httpcontract contains the strict HTTP response policy shared by
// concrete Task client libraries.
package httpcontract

import (
	"io"
	"net/http"
	"net/url"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// MaxResponseBytes bounds how much of a response body a client will read.
const MaxResponseBytes = 1 << 20

// BuildURL joins baseURL with segments and query into one request URL, or
// reports an invalid baseURL.
func BuildURL(baseURL string, segments []string, query url.Values) (string, error) {
	return "", task.ErrNotImplemented
}

// EncodeJSON marshals value as a request body.
func EncodeJSON(value any) ([]byte, error) {
	return nil, task.ErrNotImplemented
}

// ReadResponse classifies status against successStatus and errorStatuses,
// then decodes body into target on success or returns a documented or
// unexpected-response error otherwise. Callers own reading and closing body.
func ReadResponse(
	status int,
	headers http.Header,
	body io.Reader,
	successStatus int,
	errorStatuses map[int]bool,
	target any,
) error {
	return task.ErrNotImplemented
}

// ConnectionFailure classifies err as a connection failure while preserving
// context.DeadlineExceeded so callers can distinguish a timeout.
func ConnectionFailure(err error) error {
	return task.ErrNotImplemented
}

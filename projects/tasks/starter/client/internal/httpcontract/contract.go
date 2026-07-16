// Package httpcontract contains the strict HTTP response policy shared by
// concrete Task client libraries.
package httpcontract

import (
	"io"
	"net/http"
	"net/url"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

const MaxResponseBytes = 1 << 20

func BuildURL(baseURL string, segments []string, query url.Values) (string, error) {
	return "", task.ErrNotImplemented
}

func EncodeJSON(value any) ([]byte, error) {
	return nil, task.ErrNotImplemented
}

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

func ConnectionFailure(err error) error {
	return task.ErrNotImplemented
}

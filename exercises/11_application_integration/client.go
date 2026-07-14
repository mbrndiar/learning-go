package taskapi

import (
	"context"
	"net/http"
)

// Client is a small HTTP client for the task API exposed by NewServer.
type Client struct {
	// BaseURL is the server's base URL, without a trailing slash, e.g.
	// "http://localhost:8080".
	BaseURL string
	// HTTPClient is used to send requests. If nil, http.DefaultClient is
	// used.
	HTTPClient *http.Client
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// CreateTask POSTs t as JSON to /tasks and returns the created task,
// including the ID assigned by the server. ctx governs the whole
// request/response round trip, including any deadline.
//
// TODO(task 11): implement CreateTask. Marshal t to JSON, build a request
// with http.NewRequestWithContext (method POST, URL c.BaseURL+"/tasks"),
// set the "Content-Type: application/json" header, send it with
// c.httpClient().Do, and close the response body when done. A non-201
// status must produce a descriptive error (read and include the response
// body); otherwise decode the JSON body into the returned Task.
func (c *Client) CreateTask(ctx context.Context, t Task) (Task, error) {
	panic("not implemented")
}

// GetTask GETs /tasks/{id} and returns the task. A 404 response must be
// translated into an error satisfying errors.Is(err, ErrNotFound).
//
// TODO(task 12): implement GetTask, mirroring CreateTask's request/response
// handling.
func (c *Client) GetTask(ctx context.Context, id int64) (Task, error) {
	panic("not implemented")
}

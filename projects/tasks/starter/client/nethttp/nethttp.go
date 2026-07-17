// Package nethttp implements the Task client with the standard net/http stack.
package nethttp

import (
	"context"
	"net/http"
	"net/url"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// Client implements client.Transport with the standard net/http stack.
type Client struct {
	baseURL *url.URL
	http    *http.Client
}

var _ client.Transport = (*Client)(nil)

// Close releases resources owned by Client, if any.
func (c *Client) Close() error {
	return task.ErrNotImplemented
}

// New builds a Client from config, validating it first.
func New(config client.Config) (*Client, error) {
	return nil, task.ErrNotImplemented
}

// NewWithHTTPClient builds a Client around httpClient. Apply the project
// timeout without unexpectedly mutating caller-owned configuration, and decide
// explicitly whether redirects are part of the accepted protocol.
func NewWithHTTPClient(config client.Config, httpClient *http.Client) (*Client, error) {
	return nil, task.ErrNotImplemented
}

// BuildURL joins baseURL with segments and query into one request URL.
func BuildURL(baseURL string, segments []string, query url.Values) (string, error) {
	return "", task.ErrNotImplemented
}

// Create sends one create request and validates the complete response.
func (c *Client) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

// List sends an optional completed filter and returns tasks in contract order.
func (c *Client) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	return nil, task.ErrNotImplemented
}

// Get returns one task or a typed API error.
func (c *Client) Get(ctx context.Context, id int64) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

// Update sends only the fields present in input and returns the resulting task.
func (c *Client) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

// Delete requires the API's exact empty 204 response.
func (c *Client) Delete(ctx context.Context, id int64) error {
	return task.ErrNotImplemented
}

// Package nethttp implements the Task client with the standard net/http stack.
package nethttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/client/internal/httpcontract"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// Client implements the Task transport contract with net/http.
type Client struct {
	baseURL *url.URL
	http    *http.Client
}

var _ client.Transport = (*Client)(nil)

// Close releases idle connections owned by the underlying HTTP transport.
func (c *Client) Close() error {
	if c != nil && c.http != nil {
		c.http.CloseIdleConnections()
	}
	return nil
}

// New constructs a Task transport with a contract-safe net/http client.
func New(config client.Config) (*Client, error) {
	return NewWithHTTPClient(config, nil)
}

// NewWithHTTPClient constructs a Task transport around httpClient. The
// client value is always copied so the caller's configuration is never
// mutated; other caller-supplied policies such as Transport and Jar remain
// intact. The project timeout is enforced as a cap, and redirect responses
// are always treated as part of the API contract: they must be validated by
// the caller rather than silently followed, regardless of any redirect
// policy the caller-supplied client may have configured.
func NewWithHTTPClient(config client.Config, httpClient *http.Client) (*Client, error) {
	validated, err := config.Validate()
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(validated.BaseURL)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{}
	} else {
		copy := *httpClient
		httpClient = &copy
	}
	if httpClient.Timeout <= 0 || httpClient.Timeout > validated.Timeout {
		httpClient.Timeout = validated.Timeout
	}
	// Redirect responses are part of the API contract and must be
	// validated rather than silently followed.
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &Client{baseURL: baseURL, http: httpClient}, nil
}

// BuildURL appends escaped path segments and an encoded query to baseURL.
func BuildURL(baseURL string, segments []string, query url.Values) (string, error) {
	return httpcontract.BuildURL(baseURL, segments, query)
}

// Create sends one create request and validates the complete HTTP response.
func (c *Client) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	var result task.Task
	err := c.exchange(ctx, http.MethodPost, []string{"tasks"}, nil, map[string]any{"title": input.Title},
		http.StatusCreated, map[int]bool{400: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

// List returns tasks matching filter after validating response order and shape.
func (c *Client) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	query := url.Values{}
	if filter.Completed != nil {
		query.Set("completed", strconv.FormatBool(*filter.Completed))
	}
	var result []task.Task
	err := c.exchange(ctx, http.MethodGet, []string{"tasks"}, query, nil,
		http.StatusOK, map[int]bool{405: true, 422: true, 500: true}, &result)
	if result == nil && err == nil {
		result = []task.Task{}
	}
	return result, err
}

// Get returns one task or a typed API error.
func (c *Client) Get(ctx context.Context, id int64) (task.Task, error) {
	var result task.Task
	err := c.exchange(ctx, http.MethodGet, []string{"tasks", strconv.FormatInt(id, 10)}, nil, nil,
		http.StatusOK, map[int]bool{404: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

// Update applies the fields present in input and returns the resulting task.
func (c *Client) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	body := make(map[string]any)
	if input.Title != nil {
		body["title"] = *input.Title
	}
	if input.Completed != nil {
		body["completed"] = *input.Completed
	}
	var result task.Task
	err := c.exchange(ctx, http.MethodPatch, []string{"tasks", strconv.FormatInt(id, 10)}, nil, body,
		http.StatusOK, map[int]bool{400: true, 404: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

// Delete removes one task and requires an exact empty 204 response.
func (c *Client) Delete(ctx context.Context, id int64) error {
	return c.exchange(ctx, http.MethodDelete, []string{"tasks", strconv.FormatInt(id, 10)}, nil, nil,
		http.StatusNoContent, map[int]bool{404: true, 405: true, 422: true, 500: true}, nil)
}

func (c *Client) exchange(
	ctx context.Context,
	method string,
	segments []string,
	query url.Values,
	jsonBody any,
	successStatus int,
	errorStatuses map[int]bool,
	target any,
) error {
	if c == nil || c.http == nil || c.baseURL == nil {
		return &client.ConnectionError{Err: errors.New("client is not initialized")}
	}
	requestURL, err := httpcontract.BuildURL(c.baseURL.String(), segments, query)
	if err != nil {
		return &client.ConnectionError{Err: err}
	}
	var body io.Reader
	if jsonBody != nil {
		encoded, encodeErr := httpcontract.EncodeJSON(jsonBody)
		if encodeErr != nil {
			return &client.ConnectionError{Err: encodeErr}
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return &client.ConnectionError{Err: err}
	}
	if jsonBody != nil {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return httpcontract.ConnectionFailure(err)
	}
	defer response.Body.Close()
	return httpcontract.ReadResponse(
		response.StatusCode,
		response.Header,
		response.Body,
		successStatus,
		errorStatuses,
		target,
	)
}

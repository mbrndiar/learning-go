// Package taskclient provides a reusable, typed HTTP client for the task API.
//
// The client owns transport concerns—base URL resolution, finite timeouts,
// JSON encoding and decoding, response validation, and error translation—so
// that callers work with Go values and sentinel errors instead of raw HTTP
// responses. It is safe to share a single Client across goroutines.
package taskclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultTimeout bounds every request when a caller does not supply an HTTP
// client of their own. A finite timeout keeps CLI commands responsive.
const DefaultTimeout = 10 * time.Second

// maxErrorBodyBytes caps how much of an error response body the client reads
// when building an APIError message, preventing unbounded memory use.
const maxErrorBodyBytes = 4 << 10

// Task is the wire representation of a task exchanged with the API.
type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Sentinel errors let callers branch with errors.Is instead of inspecting
// strings or status codes directly.
var (
	// ErrNotFound reports that the requested task does not exist. An APIError
	// with a 404 status reports true for errors.Is(err, ErrNotFound).
	ErrNotFound = errors.New("task not found")
	// ErrTimeout reports that a request exceeded its deadline.
	ErrTimeout = errors.New("task api request timed out")
	// ErrUnavailable reports that the API could not be reached.
	ErrUnavailable = errors.New("task api unavailable")
	// ErrInvalidResponse reports that the API returned a malformed or
	// unexpected payload.
	ErrInvalidResponse = errors.New("invalid task api response")
)

// APIError describes a non-2xx response from the API while preserving the
// HTTP status code so callers can react to specific failures.
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("task api error: status %d", e.StatusCode)
	}
	return fmt.Sprintf("task api error: status %d: %s", e.StatusCode, e.Message)
}

// Is lets errors.Is match an APIError against the ErrNotFound sentinel when the
// status is 404, without collapsing every status into one error value.
func (e *APIError) Is(target error) bool {
	return target == ErrNotFound && e.StatusCode == http.StatusNotFound
}

// Client is a typed API client. Construct it with New.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// Option configures a Client during construction. Functional options keep New's
// required inputs obvious and let callers select independent optional behavior
// without a constructor containing many positional parameters.
type Option func(*Client)

// WithHTTPClient overrides the HTTP client used for requests. It is primarily
// useful in tests, for example to inject an httptest server transport.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithTimeout sets the per-request timeout on the default HTTP client. It has
// no effect when combined with WithHTTPClient.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.httpClient.Timeout = timeout
		}
	}
}

// New builds a Client that talks to the API rooted at baseURL. The base URL
// must be absolute (include a scheme and host). Options are applied in order.
func New(baseURL string, opts ...Option) (*Client, error) {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return nil, fmt.Errorf("taskclient: base URL must not be empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("taskclient: parse base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("taskclient: base URL %q must be absolute", baseURL)
	}

	// Guarantee the base path ends in a slash so relative request references
	// join onto it (rather than replacing its final segment).
	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}

	client := &Client{
		baseURL:    parsed,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}

// List returns every task known to the API.
func (c *Client) List(ctx context.Context) ([]Task, error) {
	var tasks []Task
	if err := c.do(ctx, http.MethodGet, "/tasks", nil, &tasks); err != nil {
		return nil, err
	}
	for i, task := range tasks {
		if err := validateTask(task); err != nil {
			return nil, fmt.Errorf("%w: task %d: %w", ErrInvalidResponse, i, err)
		}
	}
	return tasks, nil
}

// Get returns the task with the given identifier. It reports ErrNotFound when
// the task does not exist.
func (c *Client) Get(ctx context.Context, id int64) (Task, error) {
	if id <= 0 {
		return Task{}, fmt.Errorf("taskclient: task id must be positive, got %d", id)
	}
	var task Task
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/tasks/%d", id), nil, &task); err != nil {
		return Task{}, err
	}
	if err := validateTask(task); err != nil {
		return Task{}, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}
	return task, nil
}

// Add creates a task with the given title and returns the stored task,
// including its server-assigned identifier.
func (c *Client) Add(ctx context.Context, title string) (Task, error) {
	body := struct {
		Title string `json:"title"`
	}{Title: title}

	var task Task
	if err := c.do(ctx, http.MethodPost, "/tasks", body, &task); err != nil {
		return Task{}, err
	}
	if err := validateTask(task); err != nil {
		return Task{}, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}
	return task, nil
}

// Complete marks the task with the given identifier as done and returns the
// updated task. It reports ErrNotFound when the task does not exist.
func (c *Client) Complete(ctx context.Context, id int64) (Task, error) {
	if id <= 0 {
		return Task{}, fmt.Errorf("taskclient: task id must be positive, got %d", id)
	}
	var task Task
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/tasks/%d/complete", id), nil, &task); err != nil {
		return Task{}, err
	}
	if err := validateTask(task); err != nil {
		return Task{}, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}
	return task, nil
}

// Remove deletes the task with the given identifier. It reports ErrNotFound
// when the task does not exist.
func (c *Client) Remove(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("taskclient: task id must be positive, got %d", id)
	}
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/tasks/%d", id), nil, nil)
}

// do performs a single request/response cycle. When body is non-nil it is
// JSON-encoded. When out is non-nil a 2xx response body is JSON-decoded into
// it; otherwise the response body is discarded.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	endpoint, err := c.resolve(path)
	if err != nil {
		return err
	}

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("taskclient: encode request: %w", err)
		}
		reader = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("taskclient: build request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return translateTransportError(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
	}()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &APIError{
			StatusCode: response.StatusCode,
			Message:    readErrorMessage(response),
		}
	}

	if out == nil {
		return nil
	}

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("%w: decode body: %w", ErrInvalidResponse, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: response body must contain a single JSON value", ErrInvalidResponse)
	}
	return nil
}

// resolve joins a request path onto the client base URL, preserving any base
// path prefix the caller configured. The path is treated as relative to the
// base so a configured prefix is not discarded.
func (c *Client) resolve(path string) (string, error) {
	ref, err := url.Parse(strings.TrimPrefix(path, "/"))
	if err != nil {
		return "", fmt.Errorf("taskclient: parse path %q: %w", path, err)
	}
	return c.baseURL.ResolveReference(ref).String(), nil
}

// translateTransportError maps low-level transport failures onto the package
// sentinels while preserving the original error for errors.Is chains.
func translateTransportError(err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return fmt.Errorf("%w: %w", ErrTimeout, err)
	}
	return fmt.Errorf("%w: %w", ErrUnavailable, err)
}

// readErrorMessage extracts a human-readable message from an error response,
// tolerating both structured {"error": "..."} bodies and plain text.
func readErrorMessage(response *http.Response) string {
	data, err := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes))
	if err != nil || len(data) == 0 {
		return ""
	}

	var structured struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(data, &structured) == nil && structured.Error != "" {
		return structured.Error
	}
	return strings.TrimSpace(string(data))
}

// validateTask enforces the invariants the client expects from the API so a
// malformed payload fails fast instead of propagating silently.
func validateTask(task Task) error {
	if task.ID <= 0 {
		return fmt.Errorf("id must be positive, got %d", task.ID)
	}
	if strings.TrimSpace(task.Title) == "" {
		return errors.New("title must not be empty")
	}
	return nil
}

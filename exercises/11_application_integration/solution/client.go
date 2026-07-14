package solution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a small HTTP client for the task API exposed by NewServer.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// CreateTask POSTs t as JSON to /tasks and returns the created task.
func (c *Client) CreateTask(ctx context.Context, t Task) (Task, error) {
	body, err := json.Marshal(t)
	if err != nil {
		return Task{}, fmt.Errorf("marshaling task: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/tasks", bytes.NewReader(body))
	if err != nil {
		return Task{}, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return Task{}, fmt.Errorf("creating task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Task{}, fmt.Errorf("creating task: unexpected status %d: %s", resp.StatusCode, readBody(resp.Body))
	}

	var created Task
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return Task{}, fmt.Errorf("decoding created task: %w", err)
	}
	return created, nil
}

// GetTask GETs /tasks/{id} and returns the task. A 404 response is
// translated into an error satisfying errors.Is(err, ErrNotFound).
func (c *Client) GetTask(ctx context.Context, id int64) (Task, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/tasks/%d", c.BaseURL, id), nil)
	if err != nil {
		return Task{}, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return Task{}, fmt.Errorf("getting task %d: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Task{}, fmt.Errorf("getting task %d: %w", id, ErrNotFound)
	}
	if resp.StatusCode != http.StatusOK {
		return Task{}, fmt.Errorf("getting task %d: unexpected status %d: %s", id, resp.StatusCode, readBody(resp.Body))
	}

	var t Task
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return Task{}, fmt.Errorf("decoding task %d: %w", id, err)
	}
	return t, nil
}

func readBody(r io.Reader) string {
	b, _ := io.ReadAll(r)
	return string(b)
}

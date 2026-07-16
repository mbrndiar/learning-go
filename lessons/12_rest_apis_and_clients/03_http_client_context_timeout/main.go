// Command 03_http_client_context_timeout implements a bounded HTTP client call.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type StatusClient struct {
	HTTP    *http.Client
	BaseURL string
}

func NewStatusClient(baseURL string) *StatusClient {
	return &StatusClient{HTTP: &http.Client{Timeout: 5 * time.Second}, BaseURL: baseURL}
}

func (c *StatusClient) Get(ctx context.Context, id string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/status/"+id, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("request status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("status service returned %d: %s", resp.StatusCode, body)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode status: %w", err)
	}
	return body.Status, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	status, err := NewStatusClient("http://127.0.0.1:8081").Get(ctx, "42")
	fmt.Println(status, err)
}

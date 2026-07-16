// Command 05_resty_client demonstrates Resty's shared client configuration,
// context propagation, JSON request bodies, and typed response decoding.
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type APIClient struct {
	client *resty.Client
}

type Widget struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewAPIClient(baseURL, token string) *APIClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json").
		SetTimeout(5 * time.Second)
	return &APIClient{client: client}
}

func (c *APIClient) CreateWidget(ctx context.Context, name string) (Widget, error) {
	var widget Widget
	response, err := c.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"name": name}).
		SetResult(&widget).
		Post("/widgets")
	if err != nil {
		return Widget{}, fmt.Errorf("create widget request: %w", err)
	}
	if response.StatusCode() != http.StatusCreated {
		return Widget{}, fmt.Errorf("create widget: status %d", response.StatusCode())
	}
	return widget, nil
}

func main() {
	fmt.Println("Resty client configured with a base URL, auth token, JSON, and timeout.")
}

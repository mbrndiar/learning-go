// Package resty implements the Task client with Resty.
package resty

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	restylib "github.com/go-resty/resty/v2"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/client/internal/httpcontract"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// Client implements the Task transport contract with Resty.
type Client struct {
	baseURL *url.URL
	resty   *restylib.Client
}

var _ client.Transport = (*Client)(nil)

// Close releases idle connections owned by Resty's underlying HTTP client.
func (c *Client) Close() error {
	if c != nil && c.resty != nil && c.resty.GetClient() != nil {
		c.resty.GetClient().CloseIdleConnections()
	}
	return nil
}

// New constructs a Task transport with a contract-safe Resty client.
func New(config client.Config) (*Client, error) {
	return NewWithRestyClient(config, nil)
}

// NewWithRestyClient constructs a Task transport around restyClient.
// A supplied Resty client is configured in place with the project timeout cap,
// no-retry policy, and strict redirect policy.
func NewWithRestyClient(config client.Config, restyClient *restylib.Client) (*Client, error) {
	validated, err := config.Validate()
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(validated.BaseURL)
	if err != nil {
		return nil, err
	}
	if restyClient == nil {
		restyClient = restylib.New()
	}
	if restyClient.GetClient() == nil {
		return nil, &client.ConfigError{Field: "client", Message: "Resty client must have an HTTP client"}
	}
	if timeout := restyClient.GetClient().Timeout; timeout <= 0 || timeout > validated.Timeout {
		restyClient.SetTimeout(validated.Timeout)
	}
	// Automatic retries would hide how many requests a command performs, while
	// redirects would bypass validation of the server's actual response.
	restyClient.SetRetryCount(0)
	restyClient.SetRedirectPolicy(restylib.NoRedirectPolicy())
	return &Client{baseURL: baseURL, resty: restyClient}, nil
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
	if c == nil || c.resty == nil || c.baseURL == nil {
		return &client.ConnectionError{Err: errors.New("client is not initialized")}
	}
	requestURL, err := httpcontract.BuildURL(c.baseURL.String(), segments, query)
	if err != nil {
		return &client.ConnectionError{Err: err}
	}
	request := c.resty.R().SetContext(ctx).SetDoNotParseResponse(true)
	if jsonBody != nil {
		encoded, encodeErr := httpcontract.EncodeJSON(jsonBody)
		if encodeErr != nil {
			return &client.ConnectionError{Err: encodeErr}
		}
		request.SetHeader("Content-Type", "application/json; charset=utf-8")
		request.SetBody(encoded)
	}
	response, err := request.Execute(method, requestURL)
	if err != nil {
		return httpcontract.ConnectionFailure(err)
	}
	if response == nil || response.RawResponse == nil || response.RawBody() == nil {
		return &client.ResponseError{Message: "response was not initialized"}
	}
	defer response.RawBody().Close()
	return httpcontract.ReadResponse(
		response.StatusCode(),
		response.Header(),
		response.RawBody(),
		successStatus,
		errorStatuses,
		target,
	)
}

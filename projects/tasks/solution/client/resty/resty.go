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

type Client struct {
	baseURL *url.URL
	resty   *restylib.Client
}

var _ client.Transport = (*Client)(nil)

func (c *Client) Close() error {
	if c != nil && c.resty != nil && c.resty.GetClient() != nil {
		c.resty.GetClient().CloseIdleConnections()
	}
	return nil
}

func New(config client.Config) (*Client, error) {
	return NewWithRestyClient(config, nil)
}

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
	restyClient.SetRetryCount(0)
	restyClient.SetRedirectPolicy(restylib.NoRedirectPolicy())
	return &Client{baseURL: baseURL, resty: restyClient}, nil
}

func BuildURL(baseURL string, segments []string, query url.Values) (string, error) {
	return httpcontract.BuildURL(baseURL, segments, query)
}

func (c *Client) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	var result task.Task
	err := c.exchange(ctx, http.MethodPost, []string{"tasks"}, nil, map[string]any{"title": input.Title},
		http.StatusCreated, map[int]bool{400: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

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

func (c *Client) Get(ctx context.Context, id int64) (task.Task, error) {
	var result task.Task
	err := c.exchange(ctx, http.MethodGet, []string{"tasks", strconv.FormatInt(id, 10)}, nil, nil,
		http.StatusOK, map[int]bool{404: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

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

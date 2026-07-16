// Package resty implements the Task client with Resty.
package resty

import (
	"context"
	"net/url"

	restylib "github.com/go-resty/resty/v2"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

type Client struct {
	baseURL *url.URL
	resty   *restylib.Client
}

var _ client.Transport = (*Client)(nil)

func (c *Client) Close() error {
	return task.ErrNotImplemented
}

func New(config client.Config) (*Client, error) {
	return nil, task.ErrNotImplemented
}

func NewWithRestyClient(config client.Config, restyClient *restylib.Client) (*Client, error) {
	return nil, task.ErrNotImplemented
}

func BuildURL(baseURL string, segments []string, query url.Values) (string, error) {
	return "", task.ErrNotImplemented
}

func (c *Client) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

func (c *Client) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	return nil, task.ErrNotImplemented
}

func (c *Client) Get(ctx context.Context, id int64) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

func (c *Client) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	return task.Task{}, task.ErrNotImplemented
}

func (c *Client) Delete(ctx context.Context, id int64) error {
	return task.ErrNotImplemented
}

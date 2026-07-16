// Package restapis contains focused REST boundary and HTTP client exercises.
package restapis

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrInvalid  = errors.New("invalid input")
	ErrNotFound = errors.New("not found")
	ErrUpstream = errors.New("upstream failure")
)

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Status struct {
	Status string `json:"status"`
}

type Store interface {
	Create(context.Context, string) (Task, error)
	List(context.Context, *bool) ([]Task, error)
	Get(context.Context, int64) (Task, error)
}

type StatusClient interface {
	Check(context.Context, string) (Status, error)
}

type API struct {
	store  Store
	client StatusClient
}

func NewAPI(store Store, client StatusClient) *API {
	return &API{store: store, client: client}
}

// TODO: register POST /tasks, GET /tasks, GET /tasks/{id}, and POST /checks.
// Enforce strict one-value JSON, validate values, parse done/id, call injected
// dependencies, and map sentinel errors to JSON responses.
func (a *API) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "TODO: implement Handler", http.StatusNotImplemented)
	})
}

type HTTPStatusClient struct{ HTTP *http.Client }

// TODO: issue a context-aware GET, always close the body, reject non-200
// responses with ErrUpstream, and reject malformed or incomplete JSON.
func (c *HTTPStatusClient) Check(context.Context, string) (Status, error) {
	return Status{}, errors.New("TODO: implement Check")
}

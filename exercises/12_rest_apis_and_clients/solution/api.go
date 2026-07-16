// Package solution is the reference implementation for the REST exercises.
package solution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func NewAPI(store Store, client StatusClient) *API { return &API{store: store, client: client} }

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", a.createTask)
	mux.HandleFunc("GET /tasks", a.listTasks)
	mux.HandleFunc("GET /tasks/{id}", a.getTask)
	mux.HandleFunc("POST /checks", a.checkStatus)
	return mux
}

func (a *API) createTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := decodeOne(r, &input); err != nil {
		writeError(w, fmt.Errorf("%w: %v", ErrInvalid, err))
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		writeError(w, fmt.Errorf("%w: title is required", ErrInvalid))
		return
	}
	task, err := a.store.Create(r.Context(), input.Title)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (a *API) listTasks(w http.ResponseWriter, r *http.Request) {
	var filter *bool
	if raw, ok := r.URL.Query()["done"]; ok {
		if len(raw) != 1 {
			writeError(w, fmt.Errorf("%w: done must appear once", ErrInvalid))
			return
		}
		value, err := strconv.ParseBool(raw[0])
		if err != nil {
			writeError(w, fmt.Errorf("%w: done must be true or false", ErrInvalid))
			return
		}
		filter = &value
	}
	tasks, err := a.store.List(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (a *API) getTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, fmt.Errorf("%w: id must be a positive integer", ErrInvalid))
		return
	}
	task, err := a.store.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (a *API) checkStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL string `json:"url"`
	}
	if err := decodeOne(r, &input); err != nil {
		writeError(w, fmt.Errorf("%w: %v", ErrInvalid, err))
		return
	}
	input.URL = strings.TrimSpace(input.URL)
	if input.URL == "" {
		writeError(w, fmt.Errorf("%w: url is required", ErrInvalid))
		return
	}
	status, err := a.client.Check(r.Context(), input.URL)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func decodeOne(r *http.Request, dst any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain exactly one JSON value")
	}
	return nil
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "internal error"
	switch {
	case errors.Is(err, ErrInvalid):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, ErrNotFound):
		status, message = http.StatusNotFound, "not found"
	case errors.Is(err, ErrUpstream):
		status, message = http.StatusBadGateway, "upstream failure"
	}
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

type HTTPStatusClient struct{ HTTP *http.Client }

func (c *HTTPStatusClient) Check(ctx context.Context, url string) (Status, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Status{}, fmt.Errorf("create status request: %w", err)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Status{}, fmt.Errorf("check status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return Status{}, fmt.Errorf("status %d: %w", resp.StatusCode, ErrUpstream)
	}
	var status Status
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&status); err != nil {
		return Status{}, fmt.Errorf("decode status: %w", err)
	}
	if strings.TrimSpace(status.Status) == "" {
		return Status{}, fmt.Errorf("empty status: %w", ErrUpstream)
	}
	return status, nil
}

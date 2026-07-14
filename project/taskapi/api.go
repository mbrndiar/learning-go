package taskapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// DefaultMaxBodyBytes bounds request bodies so a client cannot exhaust server
// memory with an oversized payload.
const DefaultMaxBodyBytes = 1 << 20

// Store is the persistence contract the API depends on. It is defined here,
// where it is consumed, so the API can be tested with an in-memory fake and so
// SQLiteStore stays free of HTTP concerns.
type Store interface {
	List(ctx context.Context) ([]Task, error)
	Get(ctx context.Context, id int64) (Task, error)
	Add(ctx context.Context, title string) (Task, error)
	Complete(ctx context.Context, id int64) (Task, error)
	Remove(ctx context.Context, id int64) error
}

// API adapts a Store to HTTP. Construct it with NewAPI.
type API struct {
	store        Store
	logger       *slog.Logger
	maxBodyBytes int64
}

// APIOption configures an API during construction.
type APIOption func(*API)

// WithLogger sets the structured logger used for server-side error logging.
func WithLogger(logger *slog.Logger) APIOption {
	return func(a *API) {
		if logger != nil {
			a.logger = logger
		}
	}
}

// WithMaxBodyBytes overrides the maximum accepted request body size.
func WithMaxBodyBytes(limit int64) APIOption {
	return func(a *API) {
		if limit > 0 {
			a.maxBodyBytes = limit
		}
	}
}

// NewAPI builds an API backed by the given store.
func NewAPI(store Store, opts ...APIOption) (*API, error) {
	if store == nil {
		return nil, errors.New("taskapi: store must not be nil")
	}
	api := &API{
		store:        store,
		logger:       slog.Default(),
		maxBodyBytes: DefaultMaxBodyBytes,
	}
	for _, opt := range opts {
		opt(api)
	}
	return api, nil
}

// createTaskRequest is the accepted body for creating a task.
type createTaskRequest struct {
	Title string `json:"title"`
}

// errorResponse is the structured error body returned for every failure.
type errorResponse struct {
	Error string `json:"error"`
}

// Handler returns an http.Handler with method-aware routes for the API. The
// returned handler is safe for concurrent use.
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", a.handleList)
	mux.HandleFunc("POST /tasks", a.handleAdd)
	mux.HandleFunc("GET /tasks/{id}", a.handleGet)
	mux.HandleFunc("POST /tasks/{id}/complete", a.handleComplete)
	mux.HandleFunc("DELETE /tasks/{id}", a.handleRemove)
	return mux
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	tasks, err := a.store.List(r.Context())
	if err != nil {
		a.writeStoreError(w, r, err)
		return
	}
	a.writeJSON(w, r, http.StatusOK, tasks)
}

func (a *API) handleAdd(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := a.decodeJSON(w, r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	task, err := a.store.Add(r.Context(), req.Title)
	if err != nil {
		if errors.Is(err, ErrEmptyTitle) ||
			errors.Is(err, ErrTitleTooLong) ||
			errors.Is(err, ErrInvalidTitle) {
			a.writeError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		a.writeStoreError(w, r, err)
		return
	}
	a.writeJSON(w, r, http.StatusCreated, task)
}

func (a *API) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r)
	if err != nil {
		a.writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	task, err := a.store.Get(r.Context(), id)
	if err != nil {
		a.writeStoreError(w, r, err)
		return
	}
	a.writeJSON(w, r, http.StatusOK, task)
}

func (a *API) handleComplete(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r)
	if err != nil {
		a.writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	task, err := a.store.Complete(r.Context(), id)
	if err != nil {
		a.writeStoreError(w, r, err)
		return
	}
	a.writeJSON(w, r, http.StatusOK, task)
}

func (a *API) handleRemove(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r)
	if err != nil {
		a.writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.store.Remove(r.Context(), id); err != nil {
		a.writeStoreError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// decodeJSON reads and validates a JSON request body, enforcing the body-size
// limit and rejecting unknown fields and trailing data.
func (a *API) decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return fmt.Errorf("request body exceeds %d bytes", a.maxBodyBytes)
		}
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return fmt.Errorf("request body exceeds %d bytes", a.maxBodyBytes)
		}
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

// writeStoreError maps store failures onto HTTP status codes without comparing
// error strings, then delegates to writeError.
func (a *API) writeStoreError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		a.writeError(w, r, http.StatusNotFound, "task not found")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		a.writeError(w, r, http.StatusRequestTimeout, "request cancelled")
	default:
		a.logger.ErrorContext(r.Context(), "task store failure",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
		a.writeError(w, r, http.StatusInternalServerError, "internal server error")
	}
}

// writeError writes a structured JSON error with the given status code.
func (a *API) writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	a.writeJSON(w, r, status, errorResponse{Error: message})
}

// writeJSON encodes the payload as JSON with the given status code.
func (a *API) writeJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		a.logger.ErrorContext(r.Context(), "encode response failed",
			slog.String("error", err.Error()),
		)
	}
}

// parsePathID extracts and validates the {id} path segment.
func parsePathID(r *http.Request) (int64, error) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid task id %q", raw)
	}
	if id <= 0 {
		return 0, fmt.Errorf("task id must be positive, got %d", id)
	}
	return id, nil
}

// NewServer builds an http.Server with conservative finite timeouts. The
// caller owns starting and shutting the server down.
func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

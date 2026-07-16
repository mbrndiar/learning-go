// Package api constructs the monitor's HTTP handler.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
)

// State tracks whether the service is accepting normal API work.
type State struct {
	stopping atomic.Bool
}

// NewState constructs a running lifecycle state.
func NewState() *State {
	return &State{}
}

// Stop moves the service into its terminal stopping state.
func (state *State) Stop() {
	state.stopping.Store(true)
}

// Stopping reports whether shutdown has begun.
func (state *State) Stopping() bool {
	return state != nil && state.stopping.Load()
}

// Options customizes the HTTP handler's observable dependencies.
type Options struct {
	HistoryLimit int
	State        *State
	Logger       *slog.Logger
}

type handler struct {
	store        history.Store
	targets      []domain.Target
	configured   map[string]struct{}
	historyLimit int
	state        *State
	logger       *slog.Logger
}

// NewHandler returns an ordinary http.Handler for the monitor API.
func NewHandler(store history.Store, targets []domain.Target) http.Handler {
	limit := 1
	if bounded, ok := store.(interface{ Limit() int }); ok {
		limit = bounded.Limit()
	}
	return NewHandlerWithOptions(store, targets, Options{HistoryLimit: limit})
}

// NewHandlerWithOptions returns a configured monitor API handler.
func NewHandlerWithOptions(store history.Store, targets []domain.Target, options Options) http.Handler {
	if options.HistoryLimit < 1 {
		options.HistoryLimit = 1
	}
	if options.State == nil {
		options.State = NewState()
	}
	if options.Logger == nil {
		options.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	configured := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		configured[target.Name] = struct{}{}
	}
	return &handler{
		store:        store,
		targets:      append([]domain.Target(nil), targets...),
		configured:   configured,
		historyLimit: options.HistoryLimit,
		state:        options.State,
		logger:       options.Logger,
	}
}

func (handler *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()
	status := http.StatusOK
	defer func() {
		handler.logger.InfoContext(
			request.Context(),
			"http request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", status,
			"duration_ms", max(time.Since(start).Milliseconds(), int64(0)),
		)
	}()

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch {
	case request.URL.Path == "/healthz":
		if request.Method != http.MethodGet {
			status = methodNotAllowed(writer)
			return
		}
		if handler.state.Stopping() {
			status = http.StatusServiceUnavailable
			writeJSON(writer, status, map[string]string{"status": "stopping"})
			return
		}
		writeJSON(writer, status, map[string]string{"status": "ok"})
	case request.URL.Path == "/v1/targets":
		if request.Method != http.MethodGet {
			status = methodNotAllowed(writer)
			return
		}
		if handler.state.Stopping() {
			status = serviceStopping(writer)
			return
		}
		status = handler.serveTargets(writer)
	case strings.HasPrefix(request.URL.Path, "/v1/history/"):
		if request.Method != http.MethodGet {
			status = methodNotAllowed(writer)
			return
		}
		if handler.state.Stopping() {
			status = serviceStopping(writer)
			return
		}
		status = handler.serveHistory(writer, request)
	default:
		status = http.StatusNotFound
		writeError(writer, status, "not_found", "route was not found")
	}
}

func (handler *handler) serveTargets(writer http.ResponseWriter) int {
	if handler.store == nil {
		writeError(writer, http.StatusInternalServerError, "history_error", "current state is unavailable")
		return http.StatusInternalServerError
	}
	current := handler.store.Current()
	byTarget := make(map[string]domain.Observation, len(current))
	for _, observation := range current {
		byTarget[observation.Target] = observation
	}
	response := domain.TargetsResponse{Targets: make([]domain.TargetState, 0, len(handler.targets))}
	for _, target := range handler.targets {
		state := domain.TargetState{Target: target.Name, Status: domain.StatusUnknown}
		if observation, exists := byTarget[target.Name]; exists {
			observationCopy := observation
			state.Status = observation.Status
			state.Observation = &observationCopy
		}
		response.Targets = append(response.Targets, state)
	}
	writeJSON(writer, http.StatusOK, response)
	return http.StatusOK
}

func (handler *handler) serveHistory(writer http.ResponseWriter, request *http.Request) int {
	escapedName := strings.TrimPrefix(request.URL.EscapedPath(), "/v1/history/")
	name, err := url.PathUnescape(escapedName)
	if err != nil || name == "" || strings.Contains(name, "/") {
		writeError(writer, http.StatusNotFound, "not_found", "route was not found")
		return http.StatusNotFound
	}
	if _, exists := handler.configured[name]; !exists {
		writeError(
			writer,
			http.StatusNotFound,
			"target_not_found",
			fmt.Sprintf("target %q was not configured", name),
		)
		return http.StatusNotFound
	}
	limit, err := handler.parseLimit(request)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_limit", err.Error())
		return http.StatusBadRequest
	}
	if handler.store == nil {
		writeError(writer, http.StatusInternalServerError, "history_error", "history is unavailable")
		return http.StatusInternalServerError
	}
	observations, err := handler.store.History(name, limit)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "history_error", "history is unavailable")
		return http.StatusInternalServerError
	}
	if observations == nil {
		observations = []domain.Observation{}
	}
	writeJSON(writer, http.StatusOK, domain.HistoryResponse{
		Target:       name,
		Observations: observations,
	})
	return http.StatusOK
}

func (handler *handler) parseLimit(request *http.Request) (int, error) {
	values, present := request.URL.Query()["limit"]
	if !present {
		return handler.historyLimit, nil
	}
	if len(values) != 1 || values[0] == "" {
		return 0, fmt.Errorf("limit must be one integer between 1 and %d", handler.historyLimit)
	}
	limit, err := strconv.Atoi(values[0])
	if err != nil || limit < 1 || limit > handler.historyLimit {
		return 0, fmt.Errorf("limit must be an integer between 1 and %d", handler.historyLimit)
	}
	return limit, nil
}

func methodNotAllowed(writer http.ResponseWriter) int {
	writer.Header().Set("Allow", http.MethodGet)
	writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed", "method must be GET")
	return http.StatusMethodNotAllowed
}

func serviceStopping(writer http.ResponseWriter) int {
	writeError(writer, http.StatusServiceUnavailable, "stopping", "service is stopping")
	return http.StatusServiceUnavailable
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, domain.ErrorResponse{
		Error: domain.APIError{Code: code, Message: message},
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

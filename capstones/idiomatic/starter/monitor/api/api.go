// Package api constructs the monitor's HTTP handler.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/history"
)

const placeholderMessage = "TODO: implement monitor HTTP API"

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
	return state.stopping.Load()
}

// Options customizes the HTTP handler's observable dependencies.
type Options struct {
	HistoryLimit int
	State        *State
	Logger       *slog.Logger
}

type handler struct{}

// NewHandler returns an ordinary http.Handler for the monitor API.
func NewHandler(store history.Store, targets []domain.Target) http.Handler {
	return NewHandlerWithOptions(store, targets, Options{})
}

// NewHandlerWithOptions returns a configured monitor API handler.
func NewHandlerWithOptions(store history.Store, targets []domain.Target, options Options) http.Handler {
	return &handler{}
}

func (*handler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(writer).Encode(domain.ErrorResponse{
		Error: domain.APIError{Code: "not_implemented", Message: placeholderMessage},
	})
}

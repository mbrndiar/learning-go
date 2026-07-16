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

type State struct {
	stopping atomic.Bool
}

func NewState() *State {
	return &State{}
}

func (state *State) Stop() {
	state.stopping.Store(true)
}

func (state *State) Stopping() bool {
	return state.stopping.Load()
}

type Options struct {
	HistoryLimit int
	State        *State
	Logger       *slog.Logger
}

type handler struct{}

func NewHandler(store history.Store, targets []domain.Target) http.Handler {
	return NewHandlerWithOptions(store, targets, Options{})
}

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

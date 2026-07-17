package nethttp

import (
	"log/slog"
	"net/http"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/api"
)

// Handler routes Task HTTP requests with the standard net/http mux.
type Handler struct {
	service api.Service
	logger  *slog.Logger
	mux     *http.ServeMux
}

// New is the exercise boundary for constructing a Task Handler. Wire service,
// logger, and the mux before registering routes; every route must use the api
// package's shared decode, DTO, and error helpers.
func New(service api.Service, logger *slog.Logger) http.Handler {
	return &Handler{}
}

// ServeHTTP implements http.Handler by dispatching to the configured mux.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	api.WriteError(writer, api.MapError(nil, nil))
}

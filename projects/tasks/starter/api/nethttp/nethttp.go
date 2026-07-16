package nethttp

import (
	"log/slog"
	"net/http"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/api"
)

type Handler struct {
	service api.Service
	logger  *slog.Logger
	mux     *http.ServeMux
}

func New(service api.Service, logger *slog.Logger) http.Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	api.WriteError(writer, api.MapError(nil, nil))
}

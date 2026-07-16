package chi

import (
	"log/slog"
	"net/http"

	chilib "github.com/go-chi/chi/v5"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/api"
)

type Handler struct {
	service api.Service
	logger  *slog.Logger
	router  chilib.Router
}

func New(service api.Service, logger *slog.Logger) http.Handler {
	handler := &Handler{service: service, logger: logger, router: chilib.NewRouter()}
	placeholder := func(writer http.ResponseWriter, _ *http.Request) {
		api.WriteError(writer, api.MapError(nil, nil))
	}
	handler.router.Get("/health", placeholder)
	handler.router.Get("/tasks", placeholder)
	handler.router.Post("/tasks", placeholder)
	handler.router.Get("/tasks/{id}", placeholder)
	handler.router.Patch("/tasks/{id}", placeholder)
	handler.router.Delete("/tasks/{id}", placeholder)
	handler.router.NotFound(placeholder)
	handler.router.MethodNotAllowed(placeholder)
	return handler
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.router.ServeHTTP(writer, request)
}

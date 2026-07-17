package chi

import (
	"log/slog"
	"net/http"

	chilib "github.com/go-chi/chi/v5"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/api"
)

// Handler routes Task HTTP requests through the Chi router.
type Handler struct {
	service api.Service
	logger  *slog.Logger
	router  chilib.Router
}

// New builds a Task Handler backed by service, using logger for boundary
// failures. Registered routes currently point at a shared placeholder; each
// must eventually reach the api package's decode, DTO, and error helpers
// instead of reimplementing them.
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

// ServeHTTP implements http.Handler by dispatching to the configured router.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.router.ServeHTTP(writer, request)
}

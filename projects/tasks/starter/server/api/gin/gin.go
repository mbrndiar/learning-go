package gin

import (
	"log/slog"
	"net/http"

	ginlib "github.com/gin-gonic/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/server/api"
)

// Handler routes Task HTTP requests through the Gin router.
type Handler struct {
	service api.Service
	logger  *slog.Logger
	engine  *ginlib.Engine
}

// New builds a Task Handler backed by service, using logger for boundary
// failures. Registered routes currently point at a shared placeholder; each
// must eventually reach the api package's decode, DTO, and error helpers
// instead of reimplementing them.
func New(service api.Service, logger *slog.Logger) http.Handler {
	handler := &Handler{service: service, logger: logger, engine: ginlib.New()}
	handler.engine.RedirectTrailingSlash = false
	handler.engine.RedirectFixedPath = false
	handler.engine.RemoveExtraSlash = false
	handler.engine.HandleMethodNotAllowed = true
	placeholder := func(context *ginlib.Context) {
		api.WriteError(context.Writer, api.MapError(nil, nil))
	}
	handler.engine.GET("/health", placeholder)
	tasks := handler.engine.Group("/tasks")
	tasks.GET("", placeholder)
	tasks.POST("", placeholder)
	items := tasks.Group("/:id")
	items.GET("", placeholder)
	items.PATCH("", placeholder)
	items.DELETE("", placeholder)
	handler.engine.NoRoute(placeholder)
	handler.engine.NoMethod(placeholder)
	return handler
}

// ServeHTTP implements http.Handler by dispatching to the configured engine.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.engine.ServeHTTP(writer, request)
}

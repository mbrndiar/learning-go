package gin

import (
	"log/slog"
	"net/http"

	ginlib "github.com/gin-gonic/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/api"
)

type Handler struct {
	service api.Service
	logger  *slog.Logger
	engine  *ginlib.Engine
}

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

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.engine.ServeHTTP(writer, request)
}

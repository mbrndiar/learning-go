package gin

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	ginlib "github.com/gin-gonic/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/api"
)

type Handler struct {
	service api.Service
	logger  *slog.Logger
	engine  *ginlib.Engine
}

func New(service api.Service, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	handler := &Handler{service: service, logger: logger, engine: ginlib.New()}
	handler.engine.RedirectTrailingSlash = false
	handler.engine.RedirectFixedPath = false
	handler.engine.RemoveExtraSlash = false
	handler.engine.HandleMethodNotAllowed = true
	handler.engine.Use(handler.recovery())

	handler.engine.GET("/health", handler.health)
	tasks := handler.engine.Group("/tasks")
	tasks.GET("", handler.list)
	tasks.POST("", handler.create)
	items := tasks.Group("/:id")
	items.GET("", handler.get)
	items.PATCH("", handler.update)
	items.DELETE("", handler.delete)

	handler.engine.NoRoute(func(context *ginlib.Context) {
		api.WriteError(context.Writer, api.RouteNotFound())
	})
	handler.engine.NoMethod(func(context *ginlib.Context) {
		allow := allowedMethods(context.Request.URL.Path)
		if allow == "" {
			api.WriteError(context.Writer, api.RouteNotFound())
			return
		}
		context.Header("Allow", allow)
		api.WriteError(context.Writer, api.MethodNotAllowed(allow))
	})
	return handler
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.engine.ServeHTTP(writer, request)
}

func (h *Handler) recovery() ginlib.HandlerFunc {
	return func(context *ginlib.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				h.logger.Error("task HTTP handler panicked",
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
				context.Abort()
				if !context.Writer.Written() {
					clearHeaders(context.Writer.Header())
					api.WriteError(context.Writer, nil)
				}
			}
		}()
		context.Next()
	}
}

func (h *Handler) health(context *ginlib.Context) {
	if boundaryError := api.ValidateNoQuery(context.Request.URL.Query()); boundaryError != nil {
		api.WriteError(context.Writer, boundaryError)
		return
	}
	api.WriteJSON(context.Writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) list(context *ginlib.Context) {
	filter, boundaryError := api.ParseListFilter(context.Request.URL.Query())
	if boundaryError != nil {
		api.WriteError(context.Writer, boundaryError)
		return
	}
	values, err := h.service.List(context.Request.Context(), filter)
	if err != nil {
		api.WriteError(context.Writer, api.MapError(err, h.logger))
		return
	}
	api.WriteJSON(context.Writer, http.StatusOK, api.TaskDTOs(values))
}

func (h *Handler) create(context *ginlib.Context) {
	if boundaryError := api.ValidateNoQuery(context.Request.URL.Query()); boundaryError != nil {
		api.WriteError(context.Writer, boundaryError)
		return
	}
	input, boundaryError := api.DecodeCreate(context.Request)
	if boundaryError != nil {
		api.WriteError(context.Writer, boundaryError)
		return
	}
	value, err := h.service.Create(context.Request.Context(), input)
	if err != nil {
		api.WriteError(context.Writer, api.MapError(err, h.logger))
		return
	}
	api.WriteJSON(context.Writer, http.StatusCreated, api.TaskDTO(value))
}

func (h *Handler) get(context *ginlib.Context) {
	h.withID(context, func(id int64) {
		value, err := h.service.Get(context.Request.Context(), id)
		if err != nil {
			api.WriteError(context.Writer, api.MapError(err, h.logger))
			return
		}
		api.WriteJSON(context.Writer, http.StatusOK, api.TaskDTO(value))
	})
}

func (h *Handler) update(context *ginlib.Context) {
	h.withID(context, func(id int64) {
		input, boundaryError := api.DecodeUpdate(context.Request)
		if boundaryError != nil {
			api.WriteError(context.Writer, boundaryError)
			return
		}
		value, err := h.service.Update(context.Request.Context(), id, input)
		if err != nil {
			api.WriteError(context.Writer, api.MapError(err, h.logger))
			return
		}
		api.WriteJSON(context.Writer, http.StatusOK, api.TaskDTO(value))
	})
}

func (h *Handler) delete(context *ginlib.Context) {
	h.withID(context, func(id int64) {
		if err := h.service.Delete(context.Request.Context(), id); err != nil {
			api.WriteError(context.Writer, api.MapError(err, h.logger))
			return
		}
		context.Writer.WriteHeader(http.StatusNoContent)
	})
}

func (h *Handler) withID(context *ginlib.Context, action func(int64)) {
	if boundaryError := api.ValidateNoQuery(context.Request.URL.Query()); boundaryError != nil {
		api.WriteError(context.Writer, boundaryError)
		return
	}
	id, boundaryError := api.ParseID(context.Param("id"))
	if boundaryError != nil {
		api.WriteError(context.Writer, boundaryError)
		return
	}
	action(id)
}

func allowedMethods(requestPath string) string {
	switch requestPath {
	case "/health":
		return "GET"
	case "/tasks":
		return "GET, POST"
	}
	if strings.HasPrefix(requestPath, "/tasks/") &&
		!strings.Contains(strings.TrimPrefix(requestPath, "/tasks/"), "/") {
		return "GET, PATCH, DELETE"
	}
	return ""
}

func clearHeaders(headers http.Header) {
	for name := range headers {
		headers.Del(name)
	}
}

package gin

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	ginlib "github.com/gin-gonic/gin"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/api"
)

// Handler serves the Task HTTP contract through a Gin engine.
type Handler struct {
	service api.Service
	logger  *slog.Logger
	engine  *ginlib.Engine
}

// New constructs the strict Task HTTP adapter implemented with Gin.
func New(service api.Service, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	handler := &Handler{service: service, logger: logger, engine: ginlib.New()}
	// Gin normally redirects trailing-slash and case/cleaned-path variants
	// (e.g. "/tasks/" -> "/tasks") with a 3xx. Redirects are not part of the
	// project's HTTP contract, so all three are disabled and such requests
	// fall through to NoRoute's 404 instead.
	handler.engine.RedirectTrailingSlash = false
	handler.engine.RedirectFixedPath = false
	handler.engine.RemoveExtraSlash = false
	// Without this, Gin answers 404 for a registered path with a method it
	// doesn't recognize instead of running NoMethod below, so unsupported
	// methods would be indistinguishable from unknown routes.
	handler.engine.HandleMethodNotAllowed = true
	// gin.New() (rather than gin.Default()) skips Gin's built-in
	// Logger/Recovery middleware, which write plain-text output; this
	// custom recovery instead logs via slog and renders the shared JSON
	// error envelope, matching the other adapters.
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
		// allowedMethods recomputes the Allow set from the fixed route
		// table because Gin's NoMethod handler isn't given the methods
		// registered for the matched path.
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

// ServeHTTP dispatches a request through the configured Gin engine.
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
				// Gin's ResponseWriter already tracks whether a response
				// was started (Written()), unlike net/http's, so no custom
				// wrapper is needed here as in the nethttp/chi adapters.
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

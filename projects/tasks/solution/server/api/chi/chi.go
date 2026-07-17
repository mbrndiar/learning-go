package chi

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	chilib "github.com/go-chi/chi/v5"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api"
)

// Handler serves the Task HTTP contract through a Chi router.
type Handler struct {
	service api.Service
	logger  *slog.Logger
	router  chilib.Router
}

// New constructs the strict Task HTTP adapter implemented with Chi.
func New(service api.Service, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	handler := &Handler{service: service, logger: logger, router: chilib.NewRouter()}
	// Chi's own middleware.Recoverer writes a plain-text 500 body; replace
	// it so panics render the same JSON error envelope as every other
	// failure path.
	handler.router.Use(recovery(logger))
	handler.router.Get("/health", handler.health)
	handler.router.Get("/tasks", handler.list)
	handler.router.Post("/tasks", handler.create)
	handler.router.Get("/tasks/{id}", handler.get)
	handler.router.Patch("/tasks/{id}", handler.update)
	handler.router.Delete("/tasks/{id}", handler.delete)
	handler.router.NotFound(func(writer http.ResponseWriter, _ *http.Request) {
		api.WriteError(writer, api.RouteNotFound())
	})
	handler.router.MethodNotAllowed(func(writer http.ResponseWriter, request *http.Request) {
		// Chi's MethodNotAllowed handler is not given the set of methods
		// registered for the matched path, so recompute it from the fixed
		// route table below to populate the Allow header and error body.
		allow := allowedMethods(request.URL.Path)
		writer.Header().Set("Allow", allow)
		api.WriteError(writer, api.MethodNotAllowed(allow))
	})
	return handler
}

// ServeHTTP dispatches a request through the configured Chi router.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.router.ServeHTTP(writer, request)
}

func (h *Handler) health(writer http.ResponseWriter, request *http.Request) {
	if boundaryError := api.ValidateNoQuery(request.URL.Query()); boundaryError != nil {
		api.WriteError(writer, boundaryError)
		return
	}
	api.WriteJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) list(writer http.ResponseWriter, request *http.Request) {
	filter, boundaryError := api.ParseListFilter(request.URL.Query())
	if boundaryError != nil {
		api.WriteError(writer, boundaryError)
		return
	}
	values, err := h.service.List(request.Context(), filter)
	if err != nil {
		api.WriteError(writer, api.MapError(err, h.logger))
		return
	}
	api.WriteJSON(writer, http.StatusOK, api.TaskDTOs(values))
}

func (h *Handler) create(writer http.ResponseWriter, request *http.Request) {
	if boundaryError := api.ValidateNoQuery(request.URL.Query()); boundaryError != nil {
		api.WriteError(writer, boundaryError)
		return
	}
	input, boundaryError := api.DecodeCreate(request)
	if boundaryError != nil {
		api.WriteError(writer, boundaryError)
		return
	}
	value, err := h.service.Create(request.Context(), input)
	if err != nil {
		api.WriteError(writer, api.MapError(err, h.logger))
		return
	}
	api.WriteJSON(writer, http.StatusCreated, api.TaskDTO(value))
}

func (h *Handler) get(writer http.ResponseWriter, request *http.Request) {
	h.withID(writer, request, func(id int64) {
		value, err := h.service.Get(request.Context(), id)
		if err != nil {
			api.WriteError(writer, api.MapError(err, h.logger))
			return
		}
		api.WriteJSON(writer, http.StatusOK, api.TaskDTO(value))
	})
}

func (h *Handler) update(writer http.ResponseWriter, request *http.Request) {
	h.withID(writer, request, func(id int64) {
		input, boundaryError := api.DecodeUpdate(request)
		if boundaryError != nil {
			api.WriteError(writer, boundaryError)
			return
		}
		value, err := h.service.Update(request.Context(), id, input)
		if err != nil {
			api.WriteError(writer, api.MapError(err, h.logger))
			return
		}
		api.WriteJSON(writer, http.StatusOK, api.TaskDTO(value))
	})
}

func (h *Handler) delete(writer http.ResponseWriter, request *http.Request) {
	h.withID(writer, request, func(id int64) {
		if err := h.service.Delete(request.Context(), id); err != nil {
			api.WriteError(writer, api.MapError(err, h.logger))
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
}

func (h *Handler) withID(writer http.ResponseWriter, request *http.Request, action func(int64)) {
	if boundaryError := api.ValidateNoQuery(request.URL.Query()); boundaryError != nil {
		api.WriteError(writer, boundaryError)
		return
	}
	id, boundaryError := api.ParseID(chilib.URLParam(request, "id"))
	if boundaryError != nil {
		api.WriteError(writer, boundaryError)
		return
	}
	action(id)
}

func allowedMethods(path string) string {
	switch path {
	case "/health":
		return "GET"
	case "/tasks":
		return "GET, POST"
	default:
		return "GET, PATCH, DELETE"
	}
}

func recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			state := &responseState{ResponseWriter: writer}
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("task HTTP handler panicked",
						"panic", recovered,
						"stack", string(debug.Stack()),
					)
					// Only fall back to a clean error body if downstream
					// handlers had not already written a response.
					if !state.written {
						clearHeaders(state.Header())
						api.WriteError(state, nil)
					}
				}
			}()
			next.ServeHTTP(state, request)
		})
	}
}

// responseState wraps http.ResponseWriter to record whether a response has
// already been started, since ResponseWriter itself exposes no way to ask.
// recovery needs this to know whether it is still safe to clear headers and
// write a fresh error body after a panic.
type responseState struct {
	http.ResponseWriter
	written bool
}

func (writer *responseState) WriteHeader(status int) {
	writer.written = true
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *responseState) Write(content []byte) (int, error) {
	writer.written = true
	return writer.ResponseWriter.Write(content)
}

func clearHeaders(headers http.Header) {
	for name := range headers {
		headers.Del(name)
	}
}

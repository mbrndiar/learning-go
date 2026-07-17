package nethttp

import (
	"log/slog"
	"net/http"
	"path"
	"runtime/debug"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api"
)

// Handler serves the Task HTTP contract with the standard library router.
type Handler struct {
	service api.Service
	logger  *slog.Logger
	mux     *http.ServeMux
}

// New constructs the strict Task HTTP adapter implemented with net/http.
func New(service api.Service, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	handler := &Handler{service: service, logger: logger, mux: http.NewServeMux()}
	handler.mux.HandleFunc("GET /health", handler.health)
	// Go 1.22+ ServeMux treats a "GET pattern" as also matching HEAD, and
	// the more specific "HEAD pattern" below overrides that so HEAD -- not
	// part of the project's HTTP contract -- gets 405 like any other
	// unsupported method instead of silently running the GET handler.
	handler.mux.HandleFunc("HEAD /health", methodFallback("GET"))
	handler.mux.HandleFunc("GET /tasks", handler.list)
	handler.mux.HandleFunc("HEAD /tasks", methodFallback("GET, POST"))
	handler.mux.HandleFunc("POST /tasks", handler.create)
	handler.mux.HandleFunc("GET /tasks/{id}", handler.get)
	handler.mux.HandleFunc("HEAD /tasks/{id}", methodFallback("GET, PATCH, DELETE"))
	handler.mux.HandleFunc("PATCH /tasks/{id}", handler.update)
	handler.mux.HandleFunc("DELETE /tasks/{id}", handler.delete)
	// These method-less patterns are the catch-all for any other method on
	// a known path (e.g. PUT /tasks): without them ServeMux would fall
	// through to the "/" pattern and answer 404 instead of 405.
	handler.mux.HandleFunc("/health", methodFallback("GET"))
	handler.mux.HandleFunc("/tasks", methodFallback("GET, POST"))
	handler.mux.HandleFunc("/tasks/{id}", methodFallback("GET, PATCH, DELETE"))
	handler.mux.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		api.WriteError(writer, api.RouteNotFound())
	})
	return handler
}

// ServeHTTP applies strict path handling and dispatches through the mux.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	state := &responseState{ResponseWriter: writer}
	defer func() {
		if recovered := recover(); recovered != nil {
			h.logger.Error("task HTTP handler panicked",
				"panic", recovered,
				"stack", string(debug.Stack()),
			)
			// Only fall back to a clean error body if the handler had not
			// already written a response; otherwise the client has already
			// received a partial or committed response we cannot retract.
			if !state.written {
				clearHeaders(state.Header())
				api.WriteError(state, nil)
			}
		}
	}()
	// ServeMux would otherwise redirect non-canonical paths (repeated
	// slashes, "." or ".." segments) to a cleaned URL. Redirects are not
	// part of the project's HTTP contract, so treat any such path as an
	// unknown route instead of letting the mux issue a 3xx.
	if request.URL.Path != "/" && path.Clean(request.URL.Path) != request.URL.Path {
		api.WriteError(state, api.RouteNotFound())
		return
	}
	h.mux.ServeHTTP(state, request)
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
	id, boundaryError := api.ParseID(request.PathValue("id"))
	if boundaryError != nil {
		api.WriteError(writer, boundaryError)
		return
	}
	action(id)
}

func methodFallback(allow string) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Allow", allow)
		api.WriteError(writer, api.MethodNotAllowed(allow))
	}
}

// responseState wraps http.ResponseWriter to record whether a response has
// already been started, since ResponseWriter itself exposes no way to ask.
// The panic-recovery deferred in ServeHTTP needs this to know whether it is
// still safe to clear headers and write a fresh error body.
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

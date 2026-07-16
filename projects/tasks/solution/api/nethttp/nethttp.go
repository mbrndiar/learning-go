package nethttp

import (
	"log/slog"
	"net/http"
	"path"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/api"
)

type Handler struct {
	service api.Service
	logger  *slog.Logger
	mux     *http.ServeMux
}

func New(service api.Service, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	handler := &Handler{service: service, logger: logger, mux: http.NewServeMux()}
	handler.mux.HandleFunc("GET /health", handler.health)
	handler.mux.HandleFunc("HEAD /health", methodFallback("GET"))
	handler.mux.HandleFunc("GET /tasks", handler.list)
	handler.mux.HandleFunc("HEAD /tasks", methodFallback("GET, POST"))
	handler.mux.HandleFunc("POST /tasks", handler.create)
	handler.mux.HandleFunc("GET /tasks/{id}", handler.get)
	handler.mux.HandleFunc("HEAD /tasks/{id}", methodFallback("GET, PATCH, DELETE"))
	handler.mux.HandleFunc("PATCH /tasks/{id}", handler.update)
	handler.mux.HandleFunc("DELETE /tasks/{id}", handler.delete)
	handler.mux.HandleFunc("/health", methodFallback("GET"))
	handler.mux.HandleFunc("/tasks", methodFallback("GET, POST"))
	handler.mux.HandleFunc("/tasks/{id}", methodFallback("GET, PATCH, DELETE"))
	handler.mux.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		api.WriteError(writer, api.RouteNotFound())
	})
	return handler
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" && path.Clean(request.URL.Path) != request.URL.Path {
		api.WriteError(writer, api.RouteNotFound())
		return
	}
	h.mux.ServeHTTP(writer, request)
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

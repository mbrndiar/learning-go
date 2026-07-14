package solution

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// NewServer returns an http.Handler exposing store over HTTP.
func NewServer(store TaskStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", handleCreateTask(store))
	mux.HandleFunc("GET /tasks", handleListTasks(store))
	mux.HandleFunc("GET /tasks/{id}", handleGetTask(store))
	return mux
}

func handleCreateTask(store TaskStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
			return
		}
		if err := t.Validate(time.Now()); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		created, err := store.Create(r.Context(), t)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "creating task: "+err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func handleListTasks(store TaskStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := store.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "listing tasks: "+err.Error())
			return
		}
		if tasks == nil {
			tasks = []Task{}
		}
		writeJSON(w, http.StatusOK, tasks)
	}
}

func handleGetTask(store TaskStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		task, err := store.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "task not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "getting task: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, task)
	}
}

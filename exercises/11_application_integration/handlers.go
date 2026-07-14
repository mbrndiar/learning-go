package taskapi

import (
	"encoding/json"
	"net/http"
)

// errorResponse is the JSON body written for 4xx/5xx responses:
// {"error": "message"}.
type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON encodes v as the response body with the given status code and
// Content-Type: application/json.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes {"error": msg} with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// NewServer returns an http.Handler exposing store over HTTP:
//
//	POST /tasks       create a task from a JSON body, 201 with the created task
//	GET  /tasks       list every task, 200 with a JSON array
//	GET  /tasks/{id}  get one task by ID, 200 with the task or 404 if missing
//
// TODO(task 5): register the three routes above on a *http.ServeMux using
// Go's method-and-path patterns (e.g. mux.HandleFunc("POST /tasks", ...)),
// wiring each to the handler functions you implement below.
func NewServer(store TaskStore) http.Handler {
	panic("not implemented")
}

// handleCreateTask decodes a Task from the request body, validates it
// against the current time, and on success calls store.Create and writes
// the result with status 201. An invalid or unparsable body must produce a
// 400 response with a JSON error body via writeError, not a panic.
//
// TODO(task 6): implement handleCreateTask.
func handleCreateTask(store TaskStore) http.HandlerFunc {
	panic("not implemented")
}

// handleListTasks calls store.List and writes the result as a JSON array
// with status 200.
//
// TODO(task 7): implement handleListTasks.
func handleListTasks(store TaskStore) http.HandlerFunc {
	panic("not implemented")
}

// handleGetTask parses the "id" path value (r.PathValue("id")) as an int64,
// calls store.Get, and writes the task with status 200. A malformed id
// yields a 400; an id for which store.Get returns ErrNotFound yields a 404.
// Both must use writeError, not a panic.
//
// TODO(task 8): implement handleGetTask.
func handleGetTask(store TaskStore) http.HandlerFunc {
	panic("not implemented")
}

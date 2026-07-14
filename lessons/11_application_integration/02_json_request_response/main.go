// Command 02_json_request_response shows how to decode JSON request
// bodies safely and encode JSON responses correctly: strict decoding,
// explicit status codes and content types, and a consistent error
// envelope.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// createTaskRequest is the shape we expect a client to send. Keeping
// request/response types separate from any internal model (even when they
// look similar today) means the wire format can evolve independently of
// internal representations later.
type createTaskRequest struct {
	Title string `json:"title"`
}

// taskResponse is the shape we promise to send back.
type taskResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// errorResponse is a small, consistent envelope for reporting failures as
// JSON instead of plain text, so every client can parse errors the same way
// success responses are parsed.
type errorResponse struct {
	Error string `json:"error"`
}

// decodeStrict decodes exactly one JSON value from body into dst and
// rejects two common problems: fields the server does not recognize (a
// sign the client is out of sync with the API), and trailing data after
// the JSON value (a sign of a malformed or concatenated body).
func decodeStrict(body []byte, dst any) error {
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	if decoder.More() {
		return errors.New("decode json: unexpected trailing data")
	}
	return nil
}

// writeJSON encodes v as the response body, setting the content type and
// status code before writing anything else. Headers must be set before the
// first write to the body; net/http silently ignores header changes made
// after that.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error envelope with the given status.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

// maxRequestBody caps how many bytes a handler will read from a request
// body, so a client cannot exhaust server memory with an unbounded body.
const maxRequestBody = 1 << 20 // 1 MiB

// createTaskHandler demonstrates the full request/response boundary:
// validate the Content-Type, cap and decode the body strictly, validate
// field values, then encode a typed response.
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	raw, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read body")
		return
	}
	if len(raw) > maxRequestBody {
		writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}

	var body createTaskRequest
	if err := decodeStrict(raw, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(body.Title) == "" {
		writeError(w, http.StatusUnprocessableEntity, "title must not be empty")
		return
	}

	writeJSON(w, http.StatusCreated, taskResponse{ID: 1, Title: body.Title, Done: false})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", createTaskHandler)

	fmt.Println("registered POST /tasks; see main_test.go for example requests")
}

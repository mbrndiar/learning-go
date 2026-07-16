// Package api defines transport-neutral HTTP boundary policy for Task routers.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

const MaxBodyBytes = 1 << 20

type Service interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context, task.ListFilter) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Delete(context.Context, int64) error
}

type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error Error `json:"error"`
}

type HTTPError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "HTTP boundary error"
	}
	return e.Message
}

func TaskDTO(value task.Task) Task {
	return Task{ID: value.ID, Title: value.Title, Completed: value.Completed}
}

func TaskDTOs(values []task.Task) []Task {
	result := make([]Task, len(values))
	for index, value := range values {
		result[index] = TaskDTO(value)
	}
	return result
}

func ValidateNoQuery(query url.Values) *HTTPError {
	if len(query) == 0 {
		return nil
	}
	keys := sortedKeys(query)
	return validation(keys[0], "unknown query parameter: "+keys[0])
}

func ParseListFilter(query url.Values) (task.ListFilter, *HTTPError) {
	for _, key := range sortedKeys(query) {
		if key != "completed" {
			return task.ListFilter{}, validation(key, "unknown query parameter: "+key)
		}
	}
	values, present := query["completed"]
	if !present {
		return task.ListFilter{}, nil
	}
	if len(values) != 1 || (values[0] != "true" && values[0] != "false") {
		return task.ListFilter{}, validation("completed", "completed filter must be true or false")
	}
	completed := values[0] == "true"
	return task.ListFilter{Completed: &completed}, nil
}

func ParseID(raw string) (int64, *HTTPError) {
	if raw == "" {
		return 0, validation("id", "task ID must be a positive integer")
	}
	for _, value := range []byte(raw) {
		if value < '0' || value > '9' {
			return 0, validation("id", "task ID must be a positive integer")
		}
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, validation("id", "task ID must be a positive integer")
	}
	return id, nil
}

func DecodeCreate(request *http.Request) (task.CreateInput, *HTTPError) {
	object, boundaryError := decodeObject(request)
	if boundaryError != nil {
		return task.CreateInput{}, boundaryError
	}
	if field := firstUnknown(object, "title"); field != "" {
		return task.CreateInput{}, validation(field, "unknown property: "+field)
	}
	raw, ok := object["title"]
	if !ok {
		return task.CreateInput{}, validation("title", "missing property: title")
	}
	var title string
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &title) != nil {
		return task.CreateInput{}, validation("title", "title must be a string")
	}
	normalized, err := task.NormalizeTitle(title)
	if err != nil {
		return task.CreateInput{}, MapError(err, nil)
	}
	return task.CreateInput{Title: normalized}, nil
}

func DecodeUpdate(request *http.Request) (task.UpdateInput, *HTTPError) {
	object, boundaryError := decodeObject(request)
	if boundaryError != nil {
		return task.UpdateInput{}, boundaryError
	}
	if field := firstUnknown(object, "completed", "title"); field != "" {
		return task.UpdateInput{}, validation(field, "unknown property: "+field)
	}
	var input task.UpdateInput
	if raw, ok := object["title"]; ok {
		var title string
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &title) != nil {
			return task.UpdateInput{}, validation("title", "title must be a string")
		}
		input.Title = &title
	}
	if raw, ok := object["completed"]; ok {
		var completed bool
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &completed) != nil {
			return task.UpdateInput{}, validation("completed", "completed must be a Boolean")
		}
		input.Completed = &completed
	}
	normalized, err := task.NormalizeUpdate(input)
	if err != nil {
		return task.UpdateInput{}, MapError(err, nil)
	}
	return normalized, nil
}

func WriteJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func WriteError(writer http.ResponseWriter, boundaryError *HTTPError) {
	if boundaryError == nil {
		boundaryError = &HTTPError{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "the server could not complete the request",
		}
	}
	WriteJSON(writer, boundaryError.Status, ErrorEnvelope{Error: Error{
		Code: boundaryError.Code, Message: boundaryError.Message, Details: boundaryError.Details,
	}})
}

func MapError(err error, logger *slog.Logger) *HTTPError {
	var validationError *task.ValidationError
	switch {
	case errors.As(err, &validationError):
		return validation(validationError.Field, validationError.Message)
	case errors.Is(err, task.ErrNotFound):
		var notFoundError *task.NotFoundError
		if errors.As(err, &notFoundError) {
			return &HTTPError{Status: http.StatusNotFound, Code: "not_found", Message: notFoundError.Error()}
		}
		return &HTTPError{Status: http.StatusNotFound, Code: "not_found", Message: "task was not found"}
	case errors.Is(err, task.ErrNotImplemented):
		return &HTTPError{Status: http.StatusNotImplemented, Code: "not_implemented", Message: "this endpoint is not implemented"}
	default:
		if logger != nil {
			logger.Error("task HTTP request failed", "error", err)
		}
		return &HTTPError{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "the server could not complete the request",
		}
	}
}

func MethodNotAllowed(allow string) *HTTPError {
	return &HTTPError{
		Status: http.StatusMethodNotAllowed, Code: "method_not_allowed",
		Message: "method is not allowed for this path",
	}
}

func RouteNotFound() *HTTPError {
	return &HTTPError{Status: http.StatusNotFound, Code: "not_found", Message: "route was not found"}
}

func validation(field, message string) *HTTPError {
	return &HTTPError{
		Status: http.StatusUnprocessableEntity, Code: "validation_error", Message: message,
		Details: map[string]any{"field": field},
	}
}

func invalidJSON(message string) *HTTPError {
	return &HTTPError{Status: http.StatusBadRequest, Code: "invalid_json", Message: message}
}

func decodeObject(request *http.Request) (map[string]json.RawMessage, *HTTPError) {
	contentType := request.Header.Get("Content-Type")
	mediaType, parameters, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return nil, invalidJSON("request Content-Type must be application/json")
	}
	if charset, ok := parameters["charset"]; ok && !strings.EqualFold(charset, "utf-8") {
		return nil, invalidJSON("request JSON charset must be UTF-8")
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, MaxBodyBytes+1))
	if err != nil || len(body) > MaxBodyBytes || !utf8.Valid(body) || validateJSON(body) != nil {
		return nil, invalidJSON("request body must be valid JSON")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil || object == nil {
		return nil, validation("body", "request body must be a JSON object")
	}
	return object, nil
}

func validateJSON(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := consumeJSON(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func consumeJSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate property %q", key)
			}
			seen[key] = struct{}{}
			if err := consumeJSON(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return errors.New("unterminated object")
		}
	case '[':
		for decoder.More() {
			if err := consumeJSON(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return errors.New("unterminated array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func firstUnknown(object map[string]json.RawMessage, allowed ...string) string {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	keys := make([]string, 0)
	for key := range object {
		if _, ok := allowedSet[key]; !ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

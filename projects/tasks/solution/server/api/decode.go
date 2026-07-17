package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// MaxBodyBytes bounds decoded request bodies; decodeObject reads one byte
// past this limit so oversized bodies are rejected instead of silently
// truncated.
const MaxBodyBytes = 1 << 20

// ValidateNoQuery rejects any query parameter for endpoints that accept
// none, so adapters do not need a per-endpoint allowlist for query keys.
func ValidateNoQuery(query url.Values) *HTTPError {
	if len(query) == 0 {
		return nil
	}
	keys := sortedKeys(query)
	return validation(keys[0], "unknown query parameter: "+keys[0])
}

// ParseListFilter validates the /tasks query string and translates it into
// a domain filter, rejecting any key other than completed or a value other
// than the literal strings "true"/"false".
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

// ParseID validates a path ID as a base-10 positive integer so every
// adapter rejects the same malformed IDs the same way, independent of each
// router's own path-value extraction.
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

// DecodeCreate decodes and validates a create request body, sharing one
// strict JSON object decoder (reject unknown/duplicate keys, wrong types,
// null) across every adapter.
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

// DecodeUpdate decodes and validates a partial update request body using
// the same strict object decoder as DecodeCreate.
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

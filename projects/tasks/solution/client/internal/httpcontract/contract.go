// Package httpcontract contains the strict HTTP response policy shared by
// concrete Task client libraries.
package httpcontract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

const MaxResponseBytes = 1 << 20

func BuildURL(baseURL string, segments []string, query url.Values) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	escaped := strings.TrimRight(base.EscapedPath(), "/")
	path := strings.TrimRight(base.Path, "/")
	for _, segment := range segments {
		path += "/" + segment
		escaped += "/" + url.PathEscape(segment)
	}
	base.Path = path
	base.RawPath = escaped
	base.RawQuery = strings.ReplaceAll(query.Encode(), "+", "%20")
	return base.String(), nil
}

func EncodeJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func ReadResponse(
	status int,
	headers http.Header,
	body io.Reader,
	successStatus int,
	errorStatuses map[int]bool,
	target any,
) error {
	if status != successStatus && !errorStatuses[status] {
		return responseError(status, fmt.Sprintf("unexpected HTTP status: %d", status), nil)
	}
	responseBody, err := io.ReadAll(io.LimitReader(body, MaxResponseBytes+1))
	if err != nil {
		return responseError(status, "could not read response body", err)
	}
	if len(responseBody) > MaxResponseBytes {
		return responseError(status, "response body is too large", nil)
	}
	if status == http.StatusNoContent {
		if len(responseBody) != 0 {
			return responseError(status, "204 response body must be empty", nil)
		}
		if headers.Get("Content-Type") != "" {
			return responseError(status, "204 response must not have a Content-Type", nil)
		}
		return nil
	}
	if err := validateContentType(headers.Get("Content-Type")); err != nil {
		return responseError(status, err.Error(), err)
	}
	if !utf8.Valid(responseBody) || validateJSON(responseBody) != nil {
		return responseError(status, "response body must be valid JSON", nil)
	}
	if status != successStatus {
		return decodeAPIError(status, responseBody)
	}
	if err := decodeSuccess(responseBody, target); err != nil {
		return responseError(status, "invalid success response", err)
	}
	return nil
}

func ConnectionFailure(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &client.ConnectionError{Err: context.DeadlineExceeded}
	}
	return &client.ConnectionError{Err: errors.New("connection failed")}
}

type responseTask struct {
	ID        *int64  `json:"id"`
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

func decodeSuccess(body []byte, target any) error {
	switch value := target.(type) {
	case *task.Task:
		var wire responseTask
		if err := decodeStrict(body, &wire); err != nil {
			return err
		}
		decoded, err := validatedTask(wire)
		if err != nil {
			return err
		}
		*value = decoded
		return nil
	case *[]task.Task:
		var wire []responseTask
		if err := decodeStrict(body, &wire); err != nil || wire == nil {
			if err != nil {
				return err
			}
			return errors.New("task list must be a JSON array")
		}
		result := make([]task.Task, len(wire))
		var previous int64
		for index, item := range wire {
			decoded, err := validatedTask(item)
			if err != nil {
				return err
			}
			if decoded.ID <= previous {
				return errors.New("task list must have unique ascending IDs")
			}
			result[index] = decoded
			previous = decoded.ID
		}
		*value = result
		return nil
	default:
		return decodeStrict(body, target)
	}
}

func validatedTask(value responseTask) (task.Task, error) {
	if value.ID == nil || value.Title == nil || value.Completed == nil {
		return task.Task{}, errors.New("task response is missing a required field")
	}
	result := task.Task{ID: *value.ID, Title: *value.Title, Completed: *value.Completed}
	if err := task.ValidateTask(result); err != nil {
		return task.Task{}, err
	}
	return result, nil
}

func validateContentType(raw string) error {
	mediaType, parameters, err := mime.ParseMediaType(raw)
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return errors.New("response Content-Type must be application/json")
	}
	if charset, ok := parameters["charset"]; ok && !strings.EqualFold(charset, "utf-8") {
		return errors.New("response JSON charset must be UTF-8")
	}
	return nil
}

func decodeAPIError(status int, body []byte) error {
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if err := decodeStrict(body, &envelope); err != nil || len(envelope.Error) == 0 {
		return responseError(status, "invalid API error response", err)
	}
	var raw struct {
		Code    *string         `json:"code"`
		Message *string         `json:"message"`
		Details json.RawMessage `json:"details,omitempty"`
	}
	if err := decodeStrict(envelope.Error, &raw); err != nil || raw.Code == nil || raw.Message == nil ||
		*raw.Code == "" || *raw.Message == "" {
		return responseError(status, "invalid API error response", err)
	}
	if !validErrorCode(status, *raw.Code) {
		return responseError(status, "invalid API error code", nil)
	}
	var details map[string]any
	if len(raw.Details) != 0 {
		if bytes.Equal(bytes.TrimSpace(raw.Details), []byte("null")) ||
			json.Unmarshal(raw.Details, &details) != nil || details == nil {
			return responseError(status, "invalid API error details", nil)
		}
	}
	return &client.APIError{Status: status, Code: *raw.Code, Message: *raw.Message, Details: details}
}

func validErrorCode(status int, code string) bool {
	expected := map[int]string{
		400: "invalid_json", 404: "not_found", 405: "method_not_allowed",
		422: "validation_error", 500: "internal_error",
	}
	return expected[status] == code
}

func decodeStrict(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return errors.New("trailing JSON value")
	}
	return nil
}

func validateJSON(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := consumeJSON(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON value")
	}
	return nil
}

func consumeJSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			if !ok || seen[key] {
				return errors.New("duplicate JSON property")
			}
			seen[key] = true
			if err := consumeJSON(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
	case '[':
		for decoder.More() {
			if err := consumeJSON(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
	default:
		return errors.New("invalid JSON delimiter")
	}
	return nil
}

func responseError(status int, message string, err error) error {
	return &client.ResponseError{Status: status, Message: message, Err: err}
}

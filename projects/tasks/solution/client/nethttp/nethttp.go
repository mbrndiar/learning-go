// Package nethttp implements the Task client with the standard net/http stack.
package nethttp

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
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

const maxResponseBytes = 1 << 20

type Client struct {
	baseURL *url.URL
	http    *http.Client
}

var _ client.Transport = (*Client)(nil)

func (c *Client) Close() error {
	if c != nil && c.http != nil {
		c.http.CloseIdleConnections()
	}
	return nil
}

func New(config client.Config) (*Client, error) {
	return NewWithHTTPClient(config, nil)
}

func NewWithHTTPClient(config client.Config, httpClient *http.Client) (*Client, error) {
	validated, err := config.Validate()
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(validated.BaseURL)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: validated.Timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	} else if httpClient.Timeout <= 0 || httpClient.Timeout > validated.Timeout {
		copy := *httpClient
		copy.Timeout = validated.Timeout
		httpClient = &copy
	}
	return &Client{baseURL: baseURL, http: httpClient}, nil
}

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
	base.RawQuery = encodeQuery(query)
	return base.String(), nil
}

func encodeQuery(query url.Values) string {
	encoded := query.Encode()
	return strings.ReplaceAll(encoded, "+", "%20")
}

func (c *Client) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	var result task.Task
	err := c.exchange(ctx, http.MethodPost, []string{"tasks"}, nil, map[string]any{"title": input.Title},
		http.StatusCreated, map[int]bool{400: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

func (c *Client) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	query := url.Values{}
	if filter.Completed != nil {
		query.Set("completed", strconv.FormatBool(*filter.Completed))
	}
	var result []task.Task
	err := c.exchange(ctx, http.MethodGet, []string{"tasks"}, query, nil,
		http.StatusOK, map[int]bool{405: true, 422: true, 500: true}, &result)
	if result == nil && err == nil {
		result = []task.Task{}
	}
	return result, err
}

func (c *Client) Get(ctx context.Context, id int64) (task.Task, error) {
	var result task.Task
	err := c.exchange(ctx, http.MethodGet, []string{"tasks", strconv.FormatInt(id, 10)}, nil, nil,
		http.StatusOK, map[int]bool{404: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

func (c *Client) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	body := make(map[string]any)
	if input.Title != nil {
		body["title"] = *input.Title
	}
	if input.Completed != nil {
		body["completed"] = *input.Completed
	}
	var result task.Task
	err := c.exchange(ctx, http.MethodPatch, []string{"tasks", strconv.FormatInt(id, 10)}, nil, body,
		http.StatusOK, map[int]bool{400: true, 404: true, 405: true, 422: true, 500: true}, &result)
	return result, err
}

func (c *Client) Delete(ctx context.Context, id int64) error {
	return c.exchange(ctx, http.MethodDelete, []string{"tasks", strconv.FormatInt(id, 10)}, nil, nil,
		http.StatusNoContent, map[int]bool{404: true, 405: true, 422: true, 500: true}, nil)
}

func (c *Client) exchange(
	ctx context.Context,
	method string,
	segments []string,
	query url.Values,
	jsonBody any,
	successStatus int,
	errorStatuses map[int]bool,
	target any,
) error {
	if c == nil || c.http == nil || c.baseURL == nil {
		return &client.ConnectionError{Err: errors.New("client is not initialized")}
	}
	requestURL, err := BuildURL(c.baseURL.String(), segments, query)
	if err != nil {
		return &client.ConnectionError{Err: err}
	}
	var body io.Reader
	if jsonBody != nil {
		encoded, encodeErr := json.Marshal(jsonBody)
		if encodeErr != nil {
			return &client.ConnectionError{Err: encodeErr}
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return &client.ConnectionError{Err: err}
	}
	if jsonBody != nil {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	response, err := c.http.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return &client.ConnectionError{Err: context.DeadlineExceeded}
		}
		return &client.ConnectionError{Err: errors.New("connection failed")}
	}
	defer response.Body.Close()

	if response.StatusCode != successStatus && !errorStatuses[response.StatusCode] {
		return responseError(response.StatusCode, fmt.Sprintf("unexpected HTTP status: %d", response.StatusCode), nil)
	}
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return responseError(response.StatusCode, "could not read response body", err)
	}
	if len(responseBody) > maxResponseBytes {
		return responseError(response.StatusCode, "response body is too large", nil)
	}
	if response.StatusCode == http.StatusNoContent {
		if len(responseBody) != 0 {
			return responseError(response.StatusCode, "204 response body must be empty", nil)
		}
		if response.Header.Get("Content-Type") != "" {
			return responseError(response.StatusCode, "204 response must not have a Content-Type", nil)
		}
		return nil
	}
	if err := validateContentType(response.Header.Get("Content-Type")); err != nil {
		return responseError(response.StatusCode, err.Error(), err)
	}
	if !utf8.Valid(responseBody) || validateJSON(responseBody) != nil {
		return responseError(response.StatusCode, "response body must be valid JSON", nil)
	}
	if response.StatusCode != successStatus {
		return decodeAPIError(response.StatusCode, responseBody)
	}
	if err := decodeSuccess(responseBody, target); err != nil {
		return responseError(response.StatusCode, "invalid success response", err)
	}
	return nil
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

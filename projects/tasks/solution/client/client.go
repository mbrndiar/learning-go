// Package client defines the library-independent remote Task boundary.
package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

const (
	// DefaultTimeout is the finite default used by Task clients.
	DefaultTimeout = 5 * time.Second
)

var (
	// ErrAPI classifies documented errors returned by a Task server.
	ErrAPI = errors.New("task client: API error")
	// ErrUnexpectedResponse classifies an invalid or undocumented response.
	ErrUnexpectedResponse = errors.New("task client: unexpected response")
	// ErrConnection classifies failures reaching a Task server.
	ErrConnection = errors.New("task client: connection failure")
	// ErrInvalidConfiguration classifies invalid client configuration.
	ErrInvalidConfiguration = errors.New("task client: invalid configuration")
)

// Config contains shared client configuration.
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// Validate normalizes and validates the configuration.
func (c Config) Validate() (Config, error) {
	baseURL, err := NormalizeBaseURL(c.BaseURL)
	if err != nil {
		return Config{}, err
	}
	if c.Timeout <= 0 {
		return Config{}, &ConfigError{Field: "timeout", Message: "timeout must be positive and finite"}
	}
	c.BaseURL = baseURL
	return c, nil
}

// Transport is the remote Task capability consumed by command policy.
type Transport interface {
	Create(context.Context, task.CreateInput) (task.Task, error)
	List(context.Context, task.ListFilter) ([]task.Task, error)
	Get(context.Context, int64) (task.Task, error)
	Update(context.Context, int64, task.UpdateInput) (task.Task, error)
	Delete(context.Context, int64) error
}

// APIError is a documented error returned by the server.
type APIError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
}

// Error implements error.
func (e *APIError) Error() string {
	if e == nil {
		return ErrAPI.Error()
	}
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap classifies the error as ErrAPI.
func (e *APIError) Unwrap() error {
	return ErrAPI
}

// ResponseError describes an unexpected server response.
type ResponseError struct {
	Status  int
	Message string
	Err     error
}

// Error implements error.
func (e *ResponseError) Error() string {
	if e == nil {
		return ErrUnexpectedResponse.Error()
	}
	if e.Status > 0 {
		return fmt.Sprintf("unexpected response status %d: %s", e.Status, e.Message)
	}
	return fmt.Sprintf("unexpected response: %s", e.Message)
}

// Unwrap returns the underlying decoding or validation failure.
func (e *ResponseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is classifies every ResponseError as ErrUnexpectedResponse.
func (e *ResponseError) Is(target error) bool {
	return target == ErrUnexpectedResponse
}

// ConnectionError preserves a transport failure.
type ConnectionError struct {
	Err error
}

// Error implements error.
func (e *ConnectionError) Error() string {
	if e == nil || e.Err == nil {
		return ErrConnection.Error()
	}
	return fmt.Sprintf("%s: %v", ErrConnection, e.Err)
}

// Unwrap returns the underlying transport failure.
func (e *ConnectionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is classifies every ConnectionError as ErrConnection.
func (e *ConnectionError) Is(target error) bool {
	return target == ErrConnection
}

// ConfigError describes one invalid configuration field.
type ConfigError struct {
	Field   string
	Message string
}

// Error implements error.
func (e *ConfigError) Error() string {
	if e == nil {
		return ErrInvalidConfiguration.Error()
	}
	return e.Message
}

// Unwrap classifies the error as ErrInvalidConfiguration.
func (e *ConfigError) Unwrap() error {
	return ErrInvalidConfiguration
}

// NormalizeBaseURL validates a server base URL and removes a trailing slash.
func NormalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", &ConfigError{Field: "base-url", Message: "base URL must be an absolute HTTP URL"}
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", &ConfigError{Field: "base-url", Message: "base URL must be an absolute HTTP URL"}
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return parsed.String(), nil
}

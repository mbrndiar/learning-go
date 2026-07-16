// Package domain defines monitor configuration and observable JSON values.
package domain

import (
	"errors"
	"io"
	"time"
)

// ErrNotImplemented marks an intentional harness placeholder.
var ErrNotImplemented = errors.New("health monitor: not implemented")

// Implemented reports whether the harness placeholders have been replaced.
const Implemented = false

// Status is a target's classified health state.
type Status string

const (
	StatusUnknown   Status = "unknown"
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// ErrorCode classifies an unhealthy probe result.
type ErrorCode string

const (
	ErrorTimeout      ErrorCode = "timeout"
	ErrorCancelled    ErrorCode = "cancelled"
	ErrorTransport    ErrorCode = "transport_error"
	ErrorBodyRead     ErrorCode = "body_read_error"
	ErrorBodyTooLarge ErrorCode = "body_too_large"
)

// Config is the validated monitor configuration.
type Config struct {
	SchemaVersion  int      `json:"schema_version"`
	MaxConcurrency int      `json:"max_concurrency"`
	HistoryLimit   int      `json:"history_limit"`
	Targets        []Target `json:"targets"`
}

// Target describes one configured HTTP probe.
type Target struct {
	Name              string `json:"name"`
	URL               string `json:"url"`
	IntervalMS        int    `json:"interval_ms"`
	TimeoutMS         int    `json:"timeout_ms"`
	ExpectedStatusMin int    `json:"expected_status_min"`
	ExpectedStatusMax int    `json:"expected_status_max"`
	MaxBodyBytes      int64  `json:"max_body_bytes"`
}

// Observation is one committed probe result.
type Observation struct {
	Sequence       int64      `json:"sequence"`
	Target         string     `json:"target"`
	CheckedAt      time.Time  `json:"checked_at"`
	DurationMS     int64      `json:"duration_ms"`
	Status         Status     `json:"status"`
	PreviousStatus Status     `json:"previous_status"`
	Transition     bool       `json:"transition"`
	HTTPStatus     *int       `json:"http_status"`
	ErrorCode      *ErrorCode `json:"error_code"`
	Message        string     `json:"message"`
}

// Summary counts observations by classified status.
type Summary struct {
	Healthy   int `json:"healthy"`
	Degraded  int `json:"degraded"`
	Unhealthy int `json:"unhealthy"`
}

// CheckReport is the one-shot command response.
type CheckReport struct {
	CheckedAt time.Time     `json:"checked_at"`
	Summary   Summary       `json:"summary"`
	Results   []Observation `json:"results"`
}

// TargetState is the API representation of current target health.
type TargetState struct {
	Target      string       `json:"target"`
	Status      Status       `json:"status"`
	Observation *Observation `json:"observation"`
}

// TargetsResponse is returned by GET /v1/targets.
type TargetsResponse struct {
	Targets []TargetState `json:"targets"`
}

// HistoryResponse is returned by GET /v1/history/{name}.
type HistoryResponse struct {
	Target       string        `json:"target"`
	Observations []Observation `json:"observations"`
}

// APIError describes one HTTP API failure.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the HTTP API error envelope.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// LoadConfig decodes and validates one monitor configuration.
func LoadConfig(io.Reader) (Config, error) {
	return Config{}, ErrNotImplemented
}

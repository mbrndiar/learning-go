// Package domain defines monitor configuration and observable JSON values.
package domain

import (
	"encoding/json"
	"errors"
	"io"
	"time"
)

var (
	// ErrNotImplemented marks the incomplete boundary used by the starter harness.
	// The complete solution retains the symbol for API parity but never returns it.
	ErrNotImplemented = errors.New("health monitor: not implemented")
	// ErrInvalidConfig identifies malformed or invalid configuration.
	ErrInvalidConfig = errors.New("invalid monitor configuration")
	// ErrUnsupportedSchema identifies a configuration schema newer or older than version 1.
	ErrUnsupportedSchema = errors.New("unsupported configuration schema")
	// ErrDuplicateTarget identifies two configured targets with the same name.
	ErrDuplicateTarget = errors.New("duplicate target")
	// ErrTargetNotFound identifies a target name that was not configured.
	ErrTargetNotFound = errors.New("target not found")
	// ErrConfigIO identifies an error reading a configuration source.
	ErrConfigIO = errors.New("configuration I/O")
	// ErrInvalidLimit identifies a history limit outside the configured range.
	ErrInvalidLimit = errors.New("invalid history limit")
	// ErrHistory identifies an internal history-store failure.
	ErrHistory = errors.New("history error")
	// ErrCancelled identifies an operation cancelled before it could complete.
	ErrCancelled = errors.New("monitor operation cancelled")
)

// Implemented reports whether the guided starter has been completed.
const Implemented = false

// Status is a target's classified health state.
type Status string

const (
	StatusUnknown   Status = "unknown"
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Valid reports whether status is an observable probe classification.
func (status Status) Valid() bool {
	return status == StatusHealthy || status == StatusDegraded || status == StatusUnhealthy
}

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
	CheckedAt      time.Time  `json:"-"`
	DurationMS     int64      `json:"duration_ms"`
	Status         Status     `json:"status"`
	PreviousStatus Status     `json:"previous_status"`
	Transition     bool       `json:"transition"`
	HTTPStatus     *int       `json:"http_status"`
	ErrorCode      *ErrorCode `json:"error_code"`
	Message        string     `json:"message"`
}

// MarshalJSON emits checked_at in UTC with millisecond precision.
func (observation Observation) MarshalJSON() ([]byte, error) {
	type wire struct {
		Sequence       int64      `json:"sequence"`
		Target         string     `json:"target"`
		CheckedAt      string     `json:"checked_at"`
		DurationMS     int64      `json:"duration_ms"`
		Status         Status     `json:"status"`
		PreviousStatus Status     `json:"previous_status"`
		Transition     bool       `json:"transition"`
		HTTPStatus     *int       `json:"http_status"`
		ErrorCode      *ErrorCode `json:"error_code"`
		Message        string     `json:"message"`
	}
	return json.Marshal(wire{
		Sequence: observation.Sequence, Target: observation.Target, CheckedAt: FormatTime(observation.CheckedAt),
		DurationMS: observation.DurationMS, Status: observation.Status, PreviousStatus: observation.PreviousStatus,
		Transition: observation.Transition, HTTPStatus: observation.HTTPStatus, ErrorCode: observation.ErrorCode,
		Message: observation.Message,
	})
}

// UnmarshalJSON accepts the same millisecond RFC3339 representation emitted by MarshalJSON.
func (observation *Observation) UnmarshalJSON(data []byte) error {
	type wire struct {
		Sequence       int64      `json:"sequence"`
		Target         string     `json:"target"`
		CheckedAt      string     `json:"checked_at"`
		DurationMS     int64      `json:"duration_ms"`
		Status         Status     `json:"status"`
		PreviousStatus Status     `json:"previous_status"`
		Transition     bool       `json:"transition"`
		HTTPStatus     *int       `json:"http_status"`
		ErrorCode      *ErrorCode `json:"error_code"`
		Message        string     `json:"message"`
	}
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	checkedAt, err := time.Parse(time.RFC3339Nano, decoded.CheckedAt)
	if err != nil {
		return err
	}
	*observation = Observation{
		Sequence: decoded.Sequence, Target: decoded.Target, CheckedAt: checkedAt, DurationMS: decoded.DurationMS,
		Status: decoded.Status, PreviousStatus: decoded.PreviousStatus, Transition: decoded.Transition,
		HTTPStatus: decoded.HTTPStatus, ErrorCode: decoded.ErrorCode, Message: decoded.Message,
	}
	return nil
}

// Summary counts observations by classified status.
type Summary struct {
	Healthy   int `json:"healthy"`
	Degraded  int `json:"degraded"`
	Unhealthy int `json:"unhealthy"`
}

// CheckReport is the one-shot command response.
type CheckReport struct {
	CheckedAt time.Time     `json:"-"`
	Summary   Summary       `json:"summary"`
	Results   []Observation `json:"results"`
}

// MarshalJSON emits checked_at in UTC with millisecond precision.
func (report CheckReport) MarshalJSON() ([]byte, error) {
	type wire struct {
		CheckedAt string        `json:"checked_at"`
		Summary   Summary       `json:"summary"`
		Results   []Observation `json:"results"`
	}
	results := report.Results
	if results == nil {
		results = []Observation{}
	}
	return json.Marshal(wire{CheckedAt: FormatTime(report.CheckedAt), Summary: report.Summary, Results: results})
}

// UnmarshalJSON accepts the same millisecond RFC3339 representation emitted by MarshalJSON.
func (report *CheckReport) UnmarshalJSON(data []byte) error {
	type wire struct {
		CheckedAt string        `json:"checked_at"`
		Summary   Summary       `json:"summary"`
		Results   []Observation `json:"results"`
	}
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	checkedAt, err := time.Parse(time.RFC3339Nano, decoded.CheckedAt)
	if err != nil {
		return err
	}
	*report = CheckReport{CheckedAt: checkedAt, Summary: decoded.Summary, Results: decoded.Results}
	return nil
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

// FormatTime normalizes a timestamp to UTC millisecond precision.
func FormatTime(value time.Time) string {
	return value.UTC().Truncate(time.Millisecond).Format("2006-01-02T15:04:05.000Z")
}

// Summarize counts all classified observations.
func Summarize(observations []Observation) Summary {
	var summary Summary
	for _, observation := range observations {
		switch observation.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusDegraded:
			summary.Degraded++
		case StatusUnhealthy:
			summary.Unhealthy++
		}
	}
	return summary
}

// LoadConfig is Milestone 1: strictly decode and validate the configuration.
func LoadConfig(reader io.Reader) (Config, error) {
	return Config{}, ErrNotImplemented
}

// TargetByName returns a target and whether it was configured.
func TargetByName(targets []Target, name string) (Target, bool) {
	return Target{}, false
}

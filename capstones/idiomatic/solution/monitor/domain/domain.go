// Package domain defines monitor configuration and observable JSON values.
package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"time"
	"unicode/utf8"
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

// Implemented reports whether the harness placeholders have been replaced.
const Implemented = true

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
	type wireObservation struct {
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
	return json.Marshal(wireObservation{
		Sequence:       observation.Sequence,
		Target:         observation.Target,
		CheckedAt:      FormatTime(observation.CheckedAt),
		DurationMS:     observation.DurationMS,
		Status:         observation.Status,
		PreviousStatus: observation.PreviousStatus,
		Transition:     observation.Transition,
		HTTPStatus:     observation.HTTPStatus,
		ErrorCode:      observation.ErrorCode,
		Message:        observation.Message,
	})
}

// UnmarshalJSON accepts the same millisecond RFC3339 representation emitted by MarshalJSON.
func (observation *Observation) UnmarshalJSON(data []byte) error {
	type wireObservation struct {
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
	var wire wireObservation
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	checkedAt, err := time.Parse(time.RFC3339Nano, wire.CheckedAt)
	if err != nil {
		return fmt.Errorf("parse checked_at: %w", err)
	}
	*observation = Observation{
		Sequence:       wire.Sequence,
		Target:         wire.Target,
		CheckedAt:      checkedAt,
		DurationMS:     wire.DurationMS,
		Status:         wire.Status,
		PreviousStatus: wire.PreviousStatus,
		Transition:     wire.Transition,
		HTTPStatus:     wire.HTTPStatus,
		ErrorCode:      wire.ErrorCode,
		Message:        wire.Message,
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
	type wireReport struct {
		CheckedAt string        `json:"checked_at"`
		Summary   Summary       `json:"summary"`
		Results   []Observation `json:"results"`
	}
	results := report.Results
	if results == nil {
		results = []Observation{}
	}
	return json.Marshal(wireReport{
		CheckedAt: FormatTime(report.CheckedAt),
		Summary:   report.Summary,
		Results:   results,
	})
}

// UnmarshalJSON accepts the same millisecond RFC3339 representation emitted by MarshalJSON.
func (report *CheckReport) UnmarshalJSON(data []byte) error {
	type wireReport struct {
		CheckedAt string        `json:"checked_at"`
		Summary   Summary       `json:"summary"`
		Results   []Observation `json:"results"`
	}
	var wire wireReport
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	checkedAt, err := time.Parse(time.RFC3339Nano, wire.CheckedAt)
	if err != nil {
		return fmt.Errorf("parse checked_at: %w", err)
	}
	*report = CheckReport{CheckedAt: checkedAt, Summary: wire.Summary, Results: wire.Results}
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

type rawConfig struct {
	SchemaVersion  *int         `json:"schema_version"`
	MaxConcurrency *int         `json:"max_concurrency"`
	HistoryLimit   *int         `json:"history_limit"`
	Targets        *[]rawTarget `json:"targets"`
}

type rawTarget struct {
	Name              *string `json:"name"`
	URL               *string `json:"url"`
	IntervalMS        *int    `json:"interval_ms"`
	TimeoutMS         *int    `json:"timeout_ms"`
	ExpectedStatusMin *int    `json:"expected_status_min"`
	ExpectedStatusMax *int    `json:"expected_status_max"`
	MaxBodyBytes      *int64  `json:"max_body_bytes"`
}

var targetNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// LoadConfig decodes and validates one monitor configuration.
func LoadConfig(reader io.Reader) (Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return Config{}, fmt.Errorf("%w: read configuration: %w", ErrConfigIO, err)
	}
	if !utf8.Valid(data) {
		return Config{}, fmt.Errorf("%w: configuration is not valid UTF-8", ErrInvalidConfig)
	}
	if err := rejectDuplicateJSONFields(data); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var raw rawConfig
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, fmt.Errorf("%w: decode JSON: %w", ErrInvalidConfig, err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}
	if raw.SchemaVersion == nil || raw.MaxConcurrency == nil || raw.HistoryLimit == nil || raw.Targets == nil {
		return Config{}, fmt.Errorf("%w: top-level fields are required", ErrInvalidConfig)
	}
	if *raw.SchemaVersion != 1 {
		return Config{}, fmt.Errorf("%w: schema_version must be 1", ErrUnsupportedSchema)
	}
	if *raw.MaxConcurrency < 1 || *raw.MaxConcurrency > 32 {
		return Config{}, fmt.Errorf("%w: max_concurrency must be between 1 and 32", ErrInvalidConfig)
	}
	if *raw.HistoryLimit < 1 || *raw.HistoryLimit > 1000 {
		return Config{}, fmt.Errorf("%w: history_limit must be between 1 and 1000", ErrInvalidConfig)
	}
	if len(*raw.Targets) < 1 || len(*raw.Targets) > 100 {
		return Config{}, fmt.Errorf("%w: targets must contain between 1 and 100 entries", ErrInvalidConfig)
	}

	config := Config{
		SchemaVersion:  *raw.SchemaVersion,
		MaxConcurrency: *raw.MaxConcurrency,
		HistoryLimit:   *raw.HistoryLimit,
		Targets:        make([]Target, 0, len(*raw.Targets)),
	}
	names := make(map[string]struct{}, len(*raw.Targets))
	for index, rawTarget := range *raw.Targets {
		target, err := validateTarget(rawTarget)
		if err != nil {
			return Config{}, fmt.Errorf("target %d: %w", index, err)
		}
		if _, exists := names[target.Name]; exists {
			return Config{}, fmt.Errorf("%w: target %q is repeated", ErrDuplicateTarget, target.Name)
		}
		names[target.Name] = struct{}{}
		config.Targets = append(config.Targets, target)
	}
	return config, nil
}

func validateTarget(raw rawTarget) (Target, error) {
	if raw.Name == nil || raw.URL == nil || raw.IntervalMS == nil || raw.TimeoutMS == nil ||
		raw.ExpectedStatusMin == nil || raw.ExpectedStatusMax == nil || raw.MaxBodyBytes == nil {
		return Target{}, fmt.Errorf("%w: every target field is required", ErrInvalidConfig)
	}
	target := Target{
		Name:              *raw.Name,
		URL:               *raw.URL,
		IntervalMS:        *raw.IntervalMS,
		TimeoutMS:         *raw.TimeoutMS,
		ExpectedStatusMin: *raw.ExpectedStatusMin,
		ExpectedStatusMax: *raw.ExpectedStatusMax,
		MaxBodyBytes:      *raw.MaxBodyBytes,
	}
	if !targetNamePattern.MatchString(target.Name) {
		return Target{}, fmt.Errorf("%w: name %q is invalid", ErrInvalidConfig, target.Name)
	}
	parsedURL, err := url.Parse(target.URL)
	if err != nil {
		return Target{}, fmt.Errorf("%w: url is invalid: %w", ErrInvalidConfig, err)
	}
	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Hostname() == "" ||
		parsedURL.User != nil || parsedURL.Fragment != "" {
		return Target{}, fmt.Errorf("%w: url must be absolute http or https without user information or fragment", ErrInvalidConfig)
	}
	if target.IntervalMS < 100 || target.IntervalMS > 86_400_000 {
		return Target{}, fmt.Errorf("%w: interval_ms must be between 100 and 86400000", ErrInvalidConfig)
	}
	if target.TimeoutMS < 1 || target.TimeoutMS > target.IntervalMS {
		return Target{}, fmt.Errorf("%w: timeout_ms must be between 1 and interval_ms", ErrInvalidConfig)
	}
	if target.ExpectedStatusMin < 100 || target.ExpectedStatusMin > 599 ||
		target.ExpectedStatusMax < 100 || target.ExpectedStatusMax > 599 ||
		target.ExpectedStatusMin > target.ExpectedStatusMax {
		return Target{}, fmt.Errorf("%w: expected status range must be ordered within 100..599", ErrInvalidConfig)
	}
	if target.MaxBodyBytes < 0 || target.MaxBodyBytes > 1_048_576 {
		return Target{}, fmt.Errorf("%w: max_body_bytes must be between 0 and 1048576", ErrInvalidConfig)
	}
	return target, nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("configuration contains trailing JSON values")
	}
	return fmt.Errorf("decode trailing JSON: %w", err)
}

func rejectDuplicateJSONFields(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if err := scanJSONToken(decoder, token, "$"); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("configuration contains trailing JSON values")
		}
		return err
	}
	return nil
}

func scanJSONToken(decoder *json.Decoder, token json.Token, path string) error {
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		fields := make(map[string]struct{})
		for decoder.More() {
			nameToken, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := nameToken.(string)
			if !ok {
				return fmt.Errorf("object key at %s is not a string", path)
			}
			if _, duplicate := fields[name]; duplicate {
				return fmt.Errorf("duplicate JSON field %q at %s", name, path)
			}
			fields[name] = struct{}{}
			valueToken, err := decoder.Token()
			if err != nil {
				return err
			}
			if err := scanJSONToken(decoder, valueToken, path+"."+name); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return fmt.Errorf("object at %s was not closed", path)
		}
	case '[':
		index := 0
		for decoder.More() {
			valueToken, err := decoder.Token()
			if err != nil {
				return err
			}
			if err := scanJSONToken(decoder, valueToken, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
			index++
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return fmt.Errorf("array at %s was not closed", path)
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	return nil
}

// TargetByName returns a target and whether it was configured.
func TargetByName(targets []Target, name string) (Target, bool) {
	for _, target := range targets {
		if target.Name == name {
			return target, true
		}
	}
	return Target{}, false
}

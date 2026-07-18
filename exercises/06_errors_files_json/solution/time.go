package contacts

import (
	"errors"
	"fmt"
	"time"
)

// ErrEndBeforeStart reports an interval whose end precedes its start.
var ErrEndBeforeStart = errors.New("end before start")

// ParseTimestamp parses one RFC 3339 timestamp while preserving parse errors
// for errors.As.
func ParseTimestamp(raw string) (time.Time, error) {
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp: %w", err)
	}
	return value, nil
}

// FormatTimestampUTC formats value as an RFC 3339 timestamp normalized to UTC.
func FormatTimestampUTC(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

// Elapsed returns the duration from start to end.
func Elapsed(start, end time.Time) (time.Duration, error) {
	if end.Before(start) {
		return 0, ErrEndBeforeStart
	}
	return end.Sub(start), nil
}

package contacts

import (
	"errors"
	"time"
)

// ErrEndBeforeStart reports an interval whose end precedes its start.
var ErrEndBeforeStart = errors.New("end before start")

// ParseTimestamp parses one RFC 3339 timestamp while preserving parse errors
// for errors.As.
//
// TODO(task 7): implement ParseTimestamp.
func ParseTimestamp(raw string) (time.Time, error) {
	panic("not implemented")
}

// FormatTimestampUTC formats value as an RFC 3339 timestamp normalized to UTC.
//
// TODO(task 8): implement FormatTimestampUTC.
func FormatTimestampUTC(value time.Time) string {
	panic("not implemented")
}

// Elapsed returns the duration from start to end.
//
// TODO(task 9): return ErrEndBeforeStart when end precedes start.
func Elapsed(start, end time.Time) (time.Duration, error) {
	panic("not implemented")
}

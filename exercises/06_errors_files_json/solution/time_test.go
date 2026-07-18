package contacts

import (
	"errors"
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	t.Run("valid offset", func(t *testing.T) {
		got, err := ParseTimestamp("2026-07-18T17:45:00+02:00")
		if err != nil {
			t.Fatalf("ParseTimestamp() error = %v", err)
		}
		want := time.Date(2026, time.July, 18, 15, 45, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("ParseTimestamp() = %s, want instant %s", got, want)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := ParseTimestamp("18 July 2026")
		var parseErr *time.ParseError
		if !errors.As(err, &parseErr) {
			t.Fatalf("ParseTimestamp() error = %v, want *time.ParseError", err)
		}
	})
}

func TestFormatTimestampUTC(t *testing.T) {
	value := time.Date(
		2026, time.July, 18, 17, 45, 0, 0,
		time.FixedZone("UTC+2", 2*60*60),
	)
	if got, want := FormatTimestampUTC(value), "2026-07-18T15:45:00Z"; got != want {
		t.Fatalf("FormatTimestampUTC() = %q, want %q", got, want)
	}
}

func TestElapsed(t *testing.T) {
	start := time.Date(2026, time.July, 18, 15, 0, 0, 0, time.UTC)

	t.Run("positive", func(t *testing.T) {
		got, err := Elapsed(start, start.Add(90*time.Minute))
		if err != nil {
			t.Fatalf("Elapsed() error = %v", err)
		}
		if got != 90*time.Minute {
			t.Fatalf("Elapsed() = %s, want %s", got, 90*time.Minute)
		}
	})

	t.Run("same instant in another location", func(t *testing.T) {
		got, err := Elapsed(start, start.In(time.FixedZone("UTC+2", 2*60*60)))
		if err != nil || got != 0 {
			t.Fatalf("Elapsed() = (%s, %v), want (0s, nil)", got, err)
		}
	})

	t.Run("end before start", func(t *testing.T) {
		_, err := Elapsed(start, start.Add(-time.Nanosecond))
		if !errors.Is(err, ErrEndBeforeStart) {
			t.Fatalf("Elapsed() error = %v, want ErrEndBeforeStart", err)
		}
	})
}

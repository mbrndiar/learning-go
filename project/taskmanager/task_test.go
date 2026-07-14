package taskmanager

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"trims surrounding space", "  buy milk  ", "buy milk", nil},
		{"keeps interior space", "buy milk", "buy milk", nil},
		{"empty", "", "", ErrEmptyTitle},
		{"whitespace only", "   \t\n", "", ErrEmptyTitle},
		{"too long", strings.Repeat("a", MaxTitleLength+1), "", ErrTitleTooLong},
		{"invalid utf8", "\xff\xfe", "", ErrInvalidTitle},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeTitle(test.input)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("NormalizeTitle(%q) error = %v, want %v", test.input, err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeTitle(%q) error = %v, want nil", test.input, err)
			}
			if got != test.want {
				t.Fatalf("NormalizeTitle(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr error
	}{
		{"valid", Task{ID: 1, Title: "ok"}, nil},
		{"zero id", Task{ID: 0, Title: "ok"}, ErrInvalidID},
		{"negative id", Task{ID: -3, Title: "ok"}, ErrInvalidID},
		{"empty title", Task{ID: 1, Title: "  "}, ErrEmptyTitle},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.task.Validate()
			if test.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

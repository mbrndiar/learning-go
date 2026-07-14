package solution

import (
	"testing"
	"unicode/utf8"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"leading and trailing space", "  Hello World  ", "hello world"},
		{"multiple interior spaces", "Hello    World", "hello world"},
		{"mixed case", "HeLLo WoRLD", "hello world"},
		{"empty string", "", ""},
		{"tabs and newlines", "Hello\t\nWorld", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.in); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestWordFrequency(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[string]int
	}{
		{"normal sentence", "the quick brown fox", map[string]int{"the": 1, "quick": 1, "brown": 1, "fox": 1}},
		{"repeated words", "go go go", map[string]int{"go": 3}},
		{"empty string", "", map[string]int{}},
		{"mixed case counted together", "Go go GO", map[string]int{"go": 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordFrequency(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("WordFrequency(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for word, count := range tt.want {
				if got[word] != count {
					t.Errorf("WordFrequency(%q)[%q] = %d, want %d", tt.in, word, got[word], count)
				}
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"single rune", "a", "a"},
		{"ascii word", "hello", "olleh"},
		{"multi-byte runes", "héllo", "olléh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reverse(tt.in); got != tt.want {
				t.Errorf("Reverse(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSafeSlice(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		start   int
		length  int
		want    string
		wantErr bool
	}{
		{"success ascii", "hello world", 6, 5, "world", false},
		{"success multi-byte", "héllo", 1, 1, "é", false},
		{"negative start", "hello", -1, 2, "", true},
		{"negative length", "hello", 0, -1, "", true},
		{"start beyond rune length", "hi", 5, 1, "", true},
		{"end beyond rune length", "hi", 1, 5, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeSlice(tt.in, tt.start, tt.length)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("SafeSlice(%q, %d, %d) = %q, nil, want error", tt.in, tt.start, tt.length, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("SafeSlice(%q, %d, %d) unexpected error: %v", tt.in, tt.start, tt.length, err)
			}
			if got != tt.want {
				t.Errorf("SafeSlice(%q, %d, %d) = %q, want %q", tt.in, tt.start, tt.length, got, tt.want)
			}
		})
	}
}

func BenchmarkWordFrequency(b *testing.B) {
	input := "the quick brown fox jumps over the lazy dog the fox runs away quickly"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WordFrequency(input)
	}
}

func FuzzSafeSlice(f *testing.F) {
	f.Add("hello world", 0, 5)
	f.Add("héllo", 1, 1)
	f.Add("", 0, 0)
	f.Add("hi", 5, 1)
	f.Add("hi", 1, -3)

	f.Fuzz(func(t *testing.T, s string, start, length int) {
		got, err := SafeSlice(s, start, length)
		if err != nil {
			return
		}
		if n := utf8.RuneCountInString(got); n != length {
			t.Fatalf("SafeSlice(%q, %d, %d) returned %q with %d runes, want %d", s, start, length, got, n, length)
		}
	})
}

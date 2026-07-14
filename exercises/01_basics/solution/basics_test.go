package solution

import "testing"

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		name string
		c    float64
		want float64
	}{
		{"freezing", 0, 32},
		{"boiling", 100, 212},
		{"negative", -40, -40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CelsiusToFahrenheit(tt.c); got != tt.want {
				t.Errorf("CelsiusToFahrenheit(%v) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestFahrenheitToCelsius(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		want float64
	}{
		{"freezing", 32, 0},
		{"boiling", 212, 100},
		{"negative", -40, -40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FahrenheitToCelsius(tt.f); got != tt.want {
				t.Errorf("FahrenheitToCelsius(%v) = %v, want %v", tt.f, got, tt.want)
			}
		})
	}
}

func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		name string
		s    string
		def  int
		want int
	}{
		{"valid positive", "42", -1, 42},
		{"valid negative", "-7", 0, -7},
		{"empty string uses default", "", 9, 9},
		{"non numeric uses default", "abc", 5, 5},
		{"whitespace uses default", " 12", 3, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseIntOrDefault(tt.s, tt.def); got != tt.want {
				t.Errorf("ParseIntOrDefault(%q, %v) = %v, want %v", tt.s, tt.def, got, tt.want)
			}
		})
	}
}

func TestReverseString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"single rune", "a", "a"},
		{"ascii word", "hello", "olleh"},
		{"multi-byte runes", "héllo", "olléh"},
		{"palindrome unaffected", "level", "level"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReverseString(tt.s); got != tt.want {
				t.Errorf("ReverseString(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestCountVowels(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"all consonants", "rhythm", 0},
		{"mixed case", "Education", 5},
		{"only vowels", "aeiouAEIOU", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountVowels(tt.s); got != tt.want {
				t.Errorf("CountVowels(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"empty is palindrome", "", true},
		{"single char", "a", true},
		{"simple palindrome", "level", true},
		{"mixed case palindrome", "Level", true},
		{"not a palindrome", "hello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPalindrome(tt.s); got != tt.want {
				t.Errorf("IsPalindrome(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestByteAndRuneLen(t *testing.T) {
	tests := []struct {
		name        string
		s           string
		wantByteLen int
		wantRuneLen int
	}{
		{"empty", "", 0, 0},
		{"ascii", "go", 2, 2},
		{"multi-byte rune", "é", 2, 1},
		{"mixed", "café", 5, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotByte, gotRune := ByteAndRuneLen(tt.s)
			if gotByte != tt.wantByteLen || gotRune != tt.wantRuneLen {
				t.Errorf("ByteAndRuneLen(%q) = (%v, %v), want (%v, %v)",
					tt.s, gotByte, gotRune, tt.wantByteLen, tt.wantRuneLen)
			}
		})
	}
}

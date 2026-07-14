package solution

import (
	"errors"
	"testing"
)

func TestClassifyNumber(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"negative", -5, "negative"},
		{"zero", 0, "zero"},
		{"positive", 5, "positive"},
		{"boundary negative one", -1, "negative"},
		{"boundary one", 1, "positive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyNumber(tt.n); got != tt.want {
				t.Errorf("ClassifyNumber(%v) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestGrade(t *testing.T) {
	tests := []struct {
		name    string
		score   int
		want    string
		wantErr bool
	}{
		{"top of A", 100, "A", false},
		{"bottom of A", 90, "A", false},
		{"top of B", 89, "B", false},
		{"bottom of D", 60, "D", false},
		{"top of F", 59, "F", false},
		{"bottom of F", 0, "F", false},
		{"below range", -1, "", true},
		{"above range", 101, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Grade(tt.score)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Grade(%v) error = nil, want error", tt.score)
				}
				if !errors.Is(err, ErrInvalidScore) {
					t.Errorf("Grade(%v) error = %v, want ErrInvalidScore", tt.score, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Grade(%v) unexpected error: %v", tt.score, err)
			}
			if got != tt.want {
				t.Errorf("Grade(%v) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}

func TestSumRange(t *testing.T) {
	tests := []struct {
		name       string
		start, end int
		want       int
	}{
		{"ascending range", 1, 5, 15},
		{"single value", 3, 3, 3},
		{"start greater than end", 5, 1, 0},
		{"negative range", -2, 2, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SumRange(tt.start, tt.end); got != tt.want {
				t.Errorf("SumRange(%v, %v) = %v, want %v", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []string
	}{
		{"n <= 0 returns empty", 0, []string{}},
		{"n <= 0 negative returns empty", -3, []string{}},
		{"basic 15", 15, []string{
			"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz",
			"11", "Fizz", "13", "14", "FizzBuzz",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FizzBuzz(tt.n)
			if got == nil {
				t.Fatalf("FizzBuzz(%v) returned nil, want non-nil slice", tt.n)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("FizzBuzz(%v) len = %v, want %v", tt.n, len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("FizzBuzz(%v)[%d] = %q, want %q", tt.n, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCountDigits(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"zero", 0, 1},
		{"single digit", 7, 1},
		{"negative single digit", -7, 1},
		{"multi digit", 12345, 5},
		{"negative multi digit", -12345, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountDigits(tt.n); got != tt.want {
				t.Errorf("CountDigits(%v) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestLinearSearch(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int
	}{
		{"found middle", []int{5, 3, 8, 1}, 8, 2},
		{"found first", []int{5, 3, 8, 1}, 5, 0},
		{"not found", []int{5, 3, 8, 1}, 99, -1},
		{"empty slice", []int{}, 1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinearSearch(tt.nums, tt.target); got != tt.want {
				t.Errorf("LinearSearch(%v, %v) = %v, want %v", tt.nums, tt.target, got, tt.want)
			}
		})
	}
}

func TestBinarySearch(t *testing.T) {
	sorted := []int{1, 3, 5, 7, 9, 11}
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int
	}{
		{"found first", sorted, 1, 0},
		{"found last", sorted, 11, 5},
		{"found middle", sorted, 7, 3},
		{"not found", sorted, 4, -1},
		{"empty slice", []int{}, 1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BinarySearch(tt.nums, tt.target); got != tt.want {
				t.Errorf("BinarySearch(%v, %v) = %v, want %v", tt.nums, tt.target, got, tt.want)
			}
		})
	}
}

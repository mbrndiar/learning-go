package solution

import (
	"errors"
	"reflect"
	"testing"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{"empty", []int{}, 0},
		{"nil", nil, 0},
		{"single value", []int{7}, 7},
		{"multiple values", []int{1, 2, 3, 4}, 10},
		{"with negatives", []int{-5, 5, 10}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum(tt.nums); got != tt.want {
				t.Errorf("Sum(%v) = %v, want %v", tt.nums, got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want []int
	}{
		{"empty", []int{}, []int{}},
		{"no duplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"duplicates preserve first order", []int{3, 1, 3, 2, 1}, []int{3, 1, 2}},
		{"all same", []int{5, 5, 5}, []int{5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unique(tt.nums)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unique(%v) = %v, want %v", tt.nums, got, tt.want)
			}
		})
	}
}

func TestWordFrequency(t *testing.T) {
	tests := []struct {
		name string
		text string
		want map[string]int
	}{
		{"empty", "", map[string]int{}},
		{"single word", "go", map[string]int{"go": 1}},
		{"repeated words", "go is fun go is great", map[string]int{
			"go": 2, "is": 2, "fun": 1, "great": 1,
		}},
		{"extra whitespace", "  go   go  ", map[string]int{"go": 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordFrequency(tt.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WordFrequency(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestMergeCounts(t *testing.T) {
	a := map[string]int{"x": 1, "y": 2}
	b := map[string]int{"y": 3, "z": 4}
	aCopy := map[string]int{"x": 1, "y": 2}
	bCopy := map[string]int{"y": 3, "z": 4}

	got := MergeCounts(a, b)
	want := map[string]int{"x": 1, "y": 5, "z": 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MergeCounts(%v, %v) = %v, want %v", a, b, got, want)
	}

	if !reflect.DeepEqual(a, aCopy) {
		t.Errorf("MergeCounts mutated input a: got %v, want %v", a, aCopy)
	}
	if !reflect.DeepEqual(b, bCopy) {
		t.Errorf("MergeCounts mutated input b: got %v, want %v", b, bCopy)
	}
}

func TestSortDescending(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want []int
	}{
		{"empty", []int{}, []int{}},
		{"already descending", []int{5, 3, 1}, []int{5, 3, 1}},
		{"ascending input", []int{1, 2, 3}, []int{3, 2, 1}},
		{"with duplicates", []int{2, 5, 2, 8}, []int{8, 5, 2, 2}},
		{"with negatives", []int{-1, 3, -5, 0}, []int{3, 0, -1, -5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := make([]int, len(tt.nums))
			copy(original, tt.nums)
			got := SortDescending(tt.nums)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortDescending(%v) = %v, want %v", original, got, tt.want)
			}
			if !reflect.DeepEqual(tt.nums, original) {
				t.Errorf("SortDescending mutated input: got %v, want %v", tt.nums, original)
			}
		})
	}
}

func TestRemoveAt(t *testing.T) {
	tests := []struct {
		name    string
		nums    []int
		idx     int
		want    []int
		wantErr bool
	}{
		{"remove first", []int{1, 2, 3}, 0, []int{2, 3}, false},
		{"remove middle", []int{1, 2, 3}, 1, []int{1, 3}, false},
		{"remove last", []int{1, 2, 3}, 2, []int{1, 2}, false},
		{"negative index", []int{1, 2, 3}, -1, nil, true},
		{"index equal to length", []int{1, 2, 3}, 3, nil, true},
		{"empty slice", []int{}, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := make([]int, len(tt.nums))
			copy(original, tt.nums)
			got, err := RemoveAt(tt.nums, tt.idx)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("RemoveAt(%v, %v) error = nil, want error", tt.nums, tt.idx)
				}
				if !errors.Is(err, ErrIndexOutOfRange) {
					t.Errorf("RemoveAt(%v, %v) error = %v, want ErrIndexOutOfRange", tt.nums, tt.idx, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("RemoveAt(%v, %v) unexpected error: %v", tt.nums, tt.idx, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveAt(%v, %v) = %v, want %v", original, tt.idx, got, tt.want)
			}
			if !reflect.DeepEqual(tt.nums, original) {
				t.Errorf("RemoveAt mutated input: got %v, want %v", tt.nums, original)
			}
		})
	}
}

func TestCloneInts(t *testing.T) {
	original := []int{1, 2, 3}
	clone := CloneInts(original)

	if !reflect.DeepEqual(clone, original) {
		t.Fatalf("CloneInts(%v) = %v, want equal contents", original, clone)
	}

	// Mutating the clone must not affect the original.
	clone[0] = 99
	if original[0] == 99 {
		t.Error("mutating clone affected original: CloneInts did not copy independently")
	}

	// Mutating the original after cloning must not affect the clone.
	original2 := []int{4, 5, 6}
	clone2 := CloneInts(original2)
	original2[0] = -1
	if clone2[0] == -1 {
		t.Error("mutating original after clone affected clone: CloneInts did not copy independently")
	}
}

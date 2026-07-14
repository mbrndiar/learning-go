package solution

import (
	"errors"
	"testing"
)

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
	}{
		{"simple division", 10, 2, 5, false},
		{"negative dividend", -9, 3, -3, false},
		{"non-integer result", 1, 4, 0.25, false},
		{"division by zero", 5, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Divide(%v, %v) error = nil, want error", tt.a, tt.b)
				}
				if !errors.Is(err, ErrDivideByZero) {
					t.Errorf("Divide(%v, %v) error = %v, want ErrDivideByZero", tt.a, tt.b, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
			}
			if got != tt.want {
				t.Errorf("Divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMinMax(t *testing.T) {
	t.Run("no values returns error", func(t *testing.T) {
		_, _, err := MinMax()
		if err == nil {
			t.Fatal("MinMax() error = nil, want error")
		}
		if !errors.Is(err, ErrNoValues) {
			t.Errorf("MinMax() error = %v, want ErrNoValues", err)
		}
	})

	tests := []struct {
		name    string
		nums    []int
		wantMin int
		wantMax int
	}{
		{"single value", []int{5}, 5, 5},
		{"already sorted", []int{1, 2, 3}, 1, 3},
		{"reverse sorted", []int{9, 5, 1}, 1, 9},
		{"with negatives", []int{-3, 0, 7, -10}, -10, 7},
		{"duplicates", []int{4, 4, 4}, 4, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax, err := MinMax(tt.nums...)
			if err != nil {
				t.Fatalf("MinMax(%v) unexpected error: %v", tt.nums, err)
			}
			if gotMin != tt.wantMin || gotMax != tt.wantMax {
				t.Errorf("MinMax(%v) = (%v, %v), want (%v, %v)", tt.nums, gotMin, gotMax, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{"no arguments", nil, 0},
		{"single value", []int{5}, 5},
		{"multiple values", []int{1, 2, 3, 4}, 10},
		{"with negatives", []int{-5, 5, 10}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum(tt.nums...); got != tt.want {
				t.Errorf("Sum(%v) = %v, want %v", tt.nums, got, tt.want)
			}
		})
	}
}

func TestCounter(t *testing.T) {
	c := Counter()
	for i := 1; i <= 3; i++ {
		if got := c(); got != i {
			t.Errorf("call %d: Counter()() = %v, want %v", i, got, i)
		}
	}

	// Independent counters must not share state.
	c2 := Counter()
	if got := c2(); got != 1 {
		t.Errorf("new Counter()() = %v, want 1 (independent state)", got)
	}
	if got := c(); got != 4 {
		t.Errorf("original counter continued call = %v, want 4", got)
	}
}

func TestAccumulator(t *testing.T) {
	add := Accumulator(10)
	tests := []struct {
		add  int
		want int
	}{
		{5, 15},
		{-3, 12},
		{0, 12},
	}
	for _, tt := range tests {
		if got := add(tt.add); got != tt.want {
			t.Errorf("add(%v) = %v, want %v", tt.add, got, tt.want)
		}
	}

	// Independent accumulators must not share state.
	other := Accumulator(0)
	if got := other(1); got != 1 {
		t.Errorf("new Accumulator(0)(1) = %v, want 1 (independent state)", got)
	}
}

func TestIncrement(t *testing.T) {
	t.Run("increments value", func(t *testing.T) {
		n := 5
		Increment(&n)
		if n != 6 {
			t.Errorf("after Increment, n = %v, want 6", n)
		}
	})
	t.Run("nil pointer does not panic", func(t *testing.T) {
		Increment(nil)
	})
}

func TestSwapInts(t *testing.T) {
	t.Run("swaps values", func(t *testing.T) {
		a, b := 1, 2
		SwapInts(&a, &b)
		if a != 2 || b != 1 {
			t.Errorf("after SwapInts, a = %v, b = %v, want a = 2, b = 1", a, b)
		}
	})
	t.Run("nil pointer does not panic", func(t *testing.T) {
		a := 1
		SwapInts(&a, nil)
		SwapInts(nil, &a)
	})
}

package collections

import (
	"reflect"
	"strconv"
	"testing"
)

func TestSum(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		tests := []struct {
			name   string
			values []int
			want   int
		}{
			{"empty", nil, 0},
			{"single", []int{5}, 5},
			{"several", []int{1, 2, 3, 4}, 10},
			{"negative", []int{-1, 1, -1}, -1},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := Sum(tt.values); got != tt.want {
					t.Errorf("Sum(%v) = %v, want %v", tt.values, got, tt.want)
				}
			})
		}
	})

	t.Run("float64s", func(t *testing.T) {
		tests := []struct {
			name   string
			values []float64
			want   float64
		}{
			{"empty", nil, 0},
			{"several", []float64{1.5, 2.5, 1}, 5},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := Sum(tt.values); got != tt.want {
					t.Errorf("Sum(%v) = %v, want %v", tt.values, got, tt.want)
				}
			})
		}
	})
}

func TestMap(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   []string
	}{
		{"empty", nil, []string{}},
		{"several", []int{1, 2, 3}, []string{"1", "2", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Map(tt.values, func(v int) string { return strconv.Itoa(v) })
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	isEven := func(v int) bool { return v%2 == 0 }

	tests := []struct {
		name   string
		values []int
		want   []int
	}{
		{"empty", nil, []int{}},
		{"mixed", []int{1, 2, 3, 4, 5, 6}, []int{2, 4, 6}},
		{"none match", []int{1, 3, 5}, []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.values, isEven)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestReduce(t *testing.T) {
	t.Run("sum", func(t *testing.T) {
		got := Reduce([]int{1, 2, 3, 4}, 0, func(acc, v int) int { return acc + v })
		if got != 10 {
			t.Errorf("Reduce sum = %v, want 10", got)
		}
	})

	t.Run("concatenate", func(t *testing.T) {
		got := Reduce([]string{"a", "b", "c"}, "", func(acc, v string) string { return acc + v })
		if got != "abc" {
			t.Errorf("Reduce concatenate = %q, want %q", got, "abc")
		}
	})

	t.Run("empty values returns initial", func(t *testing.T) {
		got := Reduce([]int(nil), 42, func(acc, v int) int { return acc + v })
		if got != 42 {
			t.Errorf("Reduce empty = %v, want 42", got)
		}
	})
}

func TestStack(t *testing.T) {
	var s Stack[int]

	if _, ok := s.Pop(); ok {
		t.Fatal("Pop() on empty stack ok = true, want false")
	}
	if _, ok := s.Peek(); ok {
		t.Fatal("Peek() on empty stack ok = true, want false")
	}
	if got := s.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}

	s.Push(1)
	s.Push(2)
	s.Push(3)

	if got := s.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}

	if v, ok := s.Peek(); !ok || v != 3 {
		t.Fatalf("Peek() = (%v, %v), want (3, true)", v, ok)
	}

	wantOrder := []int{3, 2, 1}
	for _, want := range wantOrder {
		v, ok := s.Pop()
		if !ok || v != want {
			t.Fatalf("Pop() = (%v, %v), want (%v, true)", v, ok, want)
		}
	}

	if got := s.Len(); got != 0 {
		t.Fatalf("Len() after draining = %d, want 0", got)
	}
	if _, ok := s.Pop(); ok {
		t.Fatal("Pop() on drained stack ok = true, want false")
	}
}

func TestQueue(t *testing.T) {
	var q Queue[string]

	if _, ok := q.Dequeue(); ok {
		t.Fatal("Dequeue() on empty queue ok = true, want false")
	}
	if got := q.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}

	q.Enqueue("a")
	q.Enqueue("b")
	q.Enqueue("c")

	if got := q.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}

	wantOrder := []string{"a", "b", "c"}
	for _, want := range wantOrder {
		v, ok := q.Dequeue()
		if !ok || v != want {
			t.Fatalf("Dequeue() = (%v, %v), want (%v, true)", v, ok, want)
		}
	}

	if got := q.Len(); got != 0 {
		t.Fatalf("Len() after draining = %d, want 0", got)
	}
	if _, ok := q.Dequeue(); ok {
		t.Fatal("Dequeue() on drained queue ok = true, want false")
	}
}

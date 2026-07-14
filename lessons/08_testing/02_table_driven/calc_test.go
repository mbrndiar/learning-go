package tabledriven

import "testing"

// TestAdd is the classic table-driven pattern: define the cases as data,
// then loop over them, running each as its own subtest with t.Run. Subtests
// give independent pass/fail reporting and can be selected individually with
// `go test -run TestAdd/negative`.
func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{name: "positive", a: 2, b: 3, want: 5},
		{name: "zero", a: 0, b: 0, want: 0},
		{name: "negative", a: -2, b: -3, want: -5},
		{name: "mixed signs", a: -2, b: 5, want: 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Add(test.a, test.b); got != test.want {
				t.Errorf("Add(%d, %d) = %d, want %d", test.a, test.b, got, test.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
	}{
		{name: "even division", a: 10, b: 2, want: 5},
		{name: "fractional result", a: 1, b: 4, want: 0.25},
		{name: "division by zero", a: 1, b: 0, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Divide(test.a, test.b)

			if test.wantErr {
				if err == nil {
					t.Fatalf("Divide(%g, %g) = %g, nil; want an error", test.a, test.b, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Divide(%g, %g) returned unexpected error: %v", test.a, test.b, err)
			}
			if got != test.want {
				t.Errorf("Divide(%g, %g) = %g, want %g", test.a, test.b, got, test.want)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	tests := map[string]struct {
		input int
		want  int
	}{
		"positive": {input: 3, want: 3},
		"negative": {input: -3, want: 3},
		"zero":     {input: 0, want: 0},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Abs(test.input); got != test.want {
				t.Errorf("Abs(%d) = %d, want %d", test.input, got, test.want)
			}
		})
	}
}

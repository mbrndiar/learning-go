package coverage

import "testing"

// These tests deliberately do NOT cover every branch of Classify: the "D"
// grade (60-69) case has no test below. Run
//
//	go test -coverprofile=coverage.out ./lessons/08_testing/07_coverage
//	go tool cover -func=coverage.out
//	go tool cover -html=coverage.out
//
// to see Classify report less than 100% coverage, then find the missing
// line with -html before reading the source again.
func TestClassify(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{
		{name: "A grade", score: 95, want: "A"},
		{name: "B grade", score: 85, want: "B"},
		{name: "C grade", score: 72, want: "C"},
		{name: "F grade", score: 40, want: "F"},
		{name: "A boundary", score: 90, want: "A"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Classify(test.score)
			if err != nil {
				t.Fatalf("Classify(%d) returned unexpected error: %v", test.score, err)
			}
			if got != test.want {
				t.Errorf("Classify(%d) = %q, want %q", test.score, got, test.want)
			}
		})
	}
}

func TestClassifyOutOfRange(t *testing.T) {
	for _, score := range []int{-1, 101} {
		if _, err := Classify(score); err == nil {
			t.Errorf("Classify(%d) = nil error, want an error", score)
		}
	}
}

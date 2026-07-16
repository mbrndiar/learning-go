package capstones

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/testsupport"
)

func TestStarterAndSolutionExportedBoundariesMatch(t *testing.T) {
	tests := []struct {
		name     string
		starter  string
		solution string
	}{
		{
			name:     "comparative",
			starter:  "comparative/starter/kvstore",
			solution: "comparative/solution/kvstore",
		},
		{
			name:     "idiomatic",
			starter:  "idiomatic/starter/monitor",
			solution: "idiomatic/solution/monitor",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := testsupport.CompareExportedSurface(test.starter, test.solution); err != nil {
				t.Fatal(err)
			}
		})
	}
}

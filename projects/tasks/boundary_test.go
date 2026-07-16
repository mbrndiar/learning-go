package tasks

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/testsupport"
)

func TestStarterAndSolutionExportedBoundariesMatch(t *testing.T) {
	if err := testsupport.CompareExportedSurface("starter", "solution"); err != nil {
		t.Fatal(err)
	}
}

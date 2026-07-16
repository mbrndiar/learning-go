package monitor_test

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
)

func TestHarness(t *testing.T) {
	if !domain.Implemented {
		t.Fatal("solution must advertise complete behavior")
	}
}

// Package m2 contains shared comparative Milestone 2 assertions.
package m2

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/comparative/tests/contract"
)

// Run validates exact CLI and envelope behavior.
func Run(t *testing.T, program string) {
	t.Helper()
	contract.RunMilestone2(t, program)
}

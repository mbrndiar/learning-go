// Package m4 contains shared comparative Milestone 4 assertions.
package m4

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/comparative/tests/contract"
)

// Run validates revisions, expectations, ordering, and boundaries.
func Run(t *testing.T, program string) {
	t.Helper()
	contract.RunMilestone4(t, program)
}

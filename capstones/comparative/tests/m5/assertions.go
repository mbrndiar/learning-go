// Package m5 contains shared comparative Milestone 5 assertions.
package m5

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/comparative/tests/contract"
)

// Run validates real child-process concurrency and busy behavior.
func Run(t *testing.T, program string) {
	t.Helper()
	contract.RunMilestone5(t, program)
}

// Package m1 contains shared comparative Milestone 1 assertions.
package m1

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/comparative/tests/contract"
)

// Run validates one domain implementation against the frozen fixtures.
func Run(t *testing.T, subject contract.DomainSubject) {
	t.Helper()
	contract.RunMilestone1(t, subject)
}

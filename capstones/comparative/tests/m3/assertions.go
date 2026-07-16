// Package m3 contains shared comparative Milestone 3 assertions.
package m3

import (
	"testing"

	"github.com/mbrndiar/learning-go/capstones/comparative/tests/contract"
)

// Run validates SQLite initialization, storage, and migration.
func Run(t *testing.T, program string) {
	t.Helper()
	contract.RunMilestone3(t, program)
}

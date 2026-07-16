// Package history defines observation persistence consumed by the monitor.
package history

import (
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"
)

// Store records current and historical observations.
type Store interface {
	Record(domain.Observation) error
	Current() []domain.Observation
	History(string, int) ([]domain.Observation, error)
}

// MemoryStore is the future bounded, race-safe history implementation.
type MemoryStore struct{}

// NewMemoryStore constructs the in-memory history boundary.
func NewMemoryStore(_ int) *MemoryStore {
	return &MemoryStore{}
}

// Record is intentionally incomplete in the harness.
func (*MemoryStore) Record(domain.Observation) error {
	return domain.ErrNotImplemented
}

// Current returns no observations until history is implemented.
func (*MemoryStore) Current() []domain.Observation {
	return nil
}

// History is intentionally incomplete in the harness.
func (*MemoryStore) History(string, int) ([]domain.Observation, error) {
	return nil, domain.ErrNotImplemented
}

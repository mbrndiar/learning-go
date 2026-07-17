// Package history defines observation persistence consumed by the monitor.
package history

import "github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"

// Store records current and historical observations.
type Store interface {
	Record(domain.Observation) error
	Current() []domain.Observation
	History(string, int) ([]domain.Observation, error)
}

// MemoryStore is a bounded, race-safe observation store.
type MemoryStore struct {
	limit int
}

// NewMemoryStore constructs an in-memory store with a global observation limit.
func NewMemoryStore(limit int) *MemoryStore {
	return &MemoryStore{limit: limit}
}

// Limit returns the configured global history bound.
func (store *MemoryStore) Limit() int {
	return store.limit
}

// Record commits an observation, assigning its sequence and transition fields.
func (store *MemoryStore) Record(observation domain.Observation) error {
	return domain.ErrNotImplemented
}

// Current returns the latest observation for every observed target in first-seen order.
func (store *MemoryStore) Current() []domain.Observation {
	return nil
}

// History returns up to limit recent target observations in ascending sequence order.
func (store *MemoryStore) History(target string, limit int) ([]domain.Observation, error) {
	return nil, domain.ErrNotImplemented
}

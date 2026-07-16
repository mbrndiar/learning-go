// Package history defines observation persistence consumed by the monitor.
package history

import "github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"

type Store interface {
	Record(domain.Observation) error
	Current() []domain.Observation
	History(string, int) ([]domain.Observation, error)
}

type MemoryStore struct {
	limit int
}

func NewMemoryStore(limit int) *MemoryStore {
	return &MemoryStore{limit: limit}
}

func (store *MemoryStore) Limit() int {
	return store.limit
}

func (store *MemoryStore) Record(observation domain.Observation) error {
	return domain.ErrNotImplemented
}

func (store *MemoryStore) Current() []domain.Observation {
	return nil
}

func (store *MemoryStore) History(target string, limit int) ([]domain.Observation, error) {
	return nil, domain.ErrNotImplemented
}

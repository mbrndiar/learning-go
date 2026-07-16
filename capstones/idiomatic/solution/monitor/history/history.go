// Package history defines observation persistence consumed by the monitor.
package history

import (
	"fmt"
	"sync"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
)

// Store records current and historical observations.
type Store interface {
	Record(domain.Observation) error
	Current() []domain.Observation
	History(string, int) ([]domain.Observation, error)
}

// MemoryStore is a bounded, race-safe observation store.
type MemoryStore struct {
	mu       sync.RWMutex
	limit    int
	next     int64
	entries  []domain.Observation
	current  map[string]domain.Observation
	order    []string
	observed map[string]struct{}
}

// NewMemoryStore constructs an in-memory store with a global observation limit.
func NewMemoryStore(limit int) *MemoryStore {
	if limit < 1 {
		limit = 1
	}
	return &MemoryStore{
		limit:    limit,
		current:  make(map[string]domain.Observation),
		observed: make(map[string]struct{}),
	}
}

// Limit returns the configured global history bound.
func (store *MemoryStore) Limit() int {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.limit
}

// Record commits an observation, assigning its sequence and transition fields.
func (store *MemoryStore) Record(observation domain.Observation) error {
	if observation.Target == "" {
		return fmt.Errorf("%w: observation target is required", domain.ErrHistory)
	}
	if !observation.Status.Valid() {
		return fmt.Errorf("%w: observation status %q is invalid", domain.ErrHistory, observation.Status)
	}
	if observation.DurationMS < 0 {
		observation.DurationMS = 0
	}
	observation.CheckedAt = observation.CheckedAt.UTC().Truncate(1_000_000)

	store.mu.Lock()
	defer store.mu.Unlock()

	previous := domain.StatusUnknown
	if current, exists := store.current[observation.Target]; exists {
		previous = current.Status
	}
	store.next++
	observation.Sequence = store.next
	observation.PreviousStatus = previous
	observation.Transition = observation.Status != previous
	observation = cloneObservation(observation)

	if _, exists := store.observed[observation.Target]; !exists {
		store.observed[observation.Target] = struct{}{}
		store.order = append(store.order, observation.Target)
	}
	store.current[observation.Target] = observation
	store.entries = append(store.entries, observation)
	if overflow := len(store.entries) - store.limit; overflow > 0 {
		copy(store.entries, store.entries[overflow:])
		store.entries = store.entries[:store.limit]
	}
	return nil
}

// Current returns the latest observation for every observed target in first-seen order.
func (store *MemoryStore) Current() []domain.Observation {
	store.mu.RLock()
	defer store.mu.RUnlock()

	current := make([]domain.Observation, 0, len(store.order))
	for _, name := range store.order {
		current = append(current, cloneObservation(store.current[name]))
	}
	return current
}

// History returns up to limit recent target observations in ascending sequence order.
func (store *MemoryStore) History(target string, limit int) ([]domain.Observation, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if limit < 1 || limit > store.limit {
		return nil, fmt.Errorf("%w: limit must be between 1 and %d", domain.ErrInvalidLimit, store.limit)
	}

	start := 0
	count := 0
	for index := len(store.entries) - 1; index >= 0; index-- {
		if store.entries[index].Target != target {
			continue
		}
		count++
		start = index
		if count == limit {
			break
		}
	}
	if count == 0 {
		return []domain.Observation{}, nil
	}
	observations := make([]domain.Observation, 0, count)
	for index := start; index < len(store.entries); index++ {
		if store.entries[index].Target == target {
			observations = append(observations, cloneObservation(store.entries[index]))
		}
	}
	return observations, nil
}

func cloneObservation(observation domain.Observation) domain.Observation {
	if observation.HTTPStatus != nil {
		status := *observation.HTTPStatus
		observation.HTTPStatus = &status
	}
	if observation.ErrorCode != nil {
		code := *observation.ErrorCode
		observation.ErrorCode = &code
	}
	return observation
}

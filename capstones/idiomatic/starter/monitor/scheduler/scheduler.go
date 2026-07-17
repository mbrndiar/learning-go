// Package scheduler defines periodic probe ownership and lifecycle boundaries.
package scheduler

import (
	"context"
	"errors"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/probe"
)

var (
	// ErrAlreadyStarted identifies a second Start call.
	ErrAlreadyStarted = errors.New("scheduler already started")
	// ErrNotStarted identifies Wait before Start.
	ErrNotStarted = errors.New("scheduler not started")
	// ErrInvalidScheduler identifies invalid scheduler dependencies.
	ErrInvalidScheduler = errors.New("invalid scheduler")
)

// Trigger supplies deterministic cycle notifications.
type Trigger interface {
	Wait(context.Context) error
}

// TargetSelector optionally restricts the targets due after a trigger.
type TargetSelector interface {
	Select([]domain.Target) []domain.Target
}

// WallClock supplies deterministic scheduler time.
type WallClock interface {
	Now() time.Time
}

// ManualTrigger is a deterministic trigger controlled by tests or callers.
type ManualTrigger struct {
	placeholder struct{}
}

// NewManualTrigger constructs a trigger with one queued-event slot.
func NewManualTrigger() *ManualTrigger {
	return &ManualTrigger{}
}

// Fire queues one cycle notification.
func (trigger *ManualTrigger) Fire(ctx context.Context) error {
	return domain.ErrNotImplemented
}

// Wait blocks until Fire is called or the context ends.
func (trigger *ManualTrigger) Wait(ctx context.Context) error {
	return domain.ErrNotImplemented
}

// IntervalTrigger selects targets according to their configured intervals.
type IntervalTrigger struct {
	placeholder struct{}
}

// NewIntervalTrigger constructs a real-time trigger. Its first cycle is immediate.
func NewIntervalTrigger() *IntervalTrigger {
	return &IntervalTrigger{}
}

// NewIntervalTriggerWithClock constructs an interval trigger with a wall-clock seam.
func NewIntervalTriggerWithClock(clock WallClock) *IntervalTrigger {
	return &IntervalTrigger{}
}

// Wait waits until the earliest configured due time. The first notification is immediate.
func (trigger *IntervalTrigger) Wait(ctx context.Context) error {
	return domain.ErrNotImplemented
}

// Select returns due targets in configuration order and advances their due times.
func (trigger *IntervalTrigger) Select(targets []domain.Target) []domain.Target {
	return nil
}

// Scheduler owns serialized, bounded probe cycles.
type Scheduler struct {
	placeholder struct{}
}

// New constructs a scheduler.
func New(
	prober probe.Prober,
	store history.Store,
	targets []domain.Target,
	maxConcurrency int,
	trigger Trigger,
) *Scheduler {
	return &Scheduler{}
}

// Start starts the scheduler's single owner goroutine.
func (scheduler *Scheduler) Start(ctx context.Context) error {
	return domain.ErrNotImplemented
}

// Wait joins every goroutine owned by the scheduler.
func (scheduler *Scheduler) Wait() error {
	return domain.ErrNotImplemented
}

// RunCycle synchronously probes and commits every configured target.
func (scheduler *Scheduler) RunCycle(ctx context.Context) ([]domain.Observation, error) {
	return nil, domain.ErrNotImplemented
}

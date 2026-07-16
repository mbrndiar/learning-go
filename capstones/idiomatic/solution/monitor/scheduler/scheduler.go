// Package scheduler defines periodic probe ownership and lifecycle boundaries.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/probe"
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

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// ManualTrigger is a deterministic trigger controlled by tests or callers.
type ManualTrigger struct {
	events chan struct{}
}

// NewManualTrigger constructs a trigger with one queued-event slot.
func NewManualTrigger() *ManualTrigger {
	return &ManualTrigger{events: make(chan struct{}, 1)}
}

// Fire queues one cycle notification.
func (trigger *ManualTrigger) Fire(ctx context.Context) error {
	select {
	case trigger.events <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Wait blocks until Fire is called or the context ends.
func (trigger *ManualTrigger) Wait(ctx context.Context) error {
	select {
	case <-trigger.events:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IntervalTrigger selects targets according to their configured intervals.
type IntervalTrigger struct {
	mu      sync.Mutex
	clock   WallClock
	next    map[string]time.Time
	last    time.Time
	started bool
}

// NewIntervalTrigger constructs a real-time trigger. Its first cycle is immediate.
func NewIntervalTrigger() *IntervalTrigger {
	return NewIntervalTriggerWithClock(realClock{})
}

// NewIntervalTriggerWithClock constructs an interval trigger with a wall-clock seam.
func NewIntervalTriggerWithClock(clock WallClock) *IntervalTrigger {
	if clock == nil {
		clock = realClock{}
	}
	return &IntervalTrigger{clock: clock, next: make(map[string]time.Time)}
}

// Wait waits until the earliest configured due time. The first notification is immediate.
func (trigger *IntervalTrigger) Wait(ctx context.Context) error {
	trigger.mu.Lock()
	if !trigger.started {
		trigger.started = true
		trigger.last = trigger.clock.Now()
		trigger.mu.Unlock()
		return nil
	}
	now := trigger.clock.Now()
	var earliest time.Time
	for _, due := range trigger.next {
		if earliest.IsZero() || due.Before(earliest) {
			earliest = due
		}
	}
	if earliest.IsZero() || !earliest.After(now) {
		trigger.last = now
		trigger.mu.Unlock()
		return nil
	}
	delay := earliest.Sub(now)
	trigger.mu.Unlock()

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		trigger.mu.Lock()
		trigger.last = trigger.clock.Now()
		trigger.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Select returns due targets in configuration order and advances their due times.
func (trigger *IntervalTrigger) Select(targets []domain.Target) []domain.Target {
	trigger.mu.Lock()
	defer trigger.mu.Unlock()
	now := trigger.last
	if now.IsZero() {
		now = trigger.clock.Now()
	}
	selected := make([]domain.Target, 0, len(targets))
	for _, target := range targets {
		due, exists := trigger.next[target.Name]
		if exists && due.After(now) {
			continue
		}
		selected = append(selected, target)
		interval := time.Duration(target.IntervalMS) * time.Millisecond
		next := due
		if next.IsZero() {
			next = now
		}
		for !next.After(now) {
			next = next.Add(interval)
		}
		trigger.next[target.Name] = next
	}
	return selected
}

// Scheduler owns serialized, bounded probe cycles.
type Scheduler struct {
	prober         probe.Prober
	store          history.Store
	targets        []domain.Target
	maxConcurrency int
	trigger        Trigger
	cycleToken     chan struct{}

	mu      sync.Mutex
	started bool
	done    chan struct{}
	err     error
}

// New constructs a scheduler.
func New(
	prober probe.Prober,
	store history.Store,
	targets []domain.Target,
	maxConcurrency int,
	trigger Trigger,
) *Scheduler {
	if trigger == nil {
		trigger = NewIntervalTrigger()
	}
	cycleToken := make(chan struct{}, 1)
	cycleToken <- struct{}{}
	return &Scheduler{
		prober:         prober,
		store:          store,
		targets:        append([]domain.Target(nil), targets...),
		maxConcurrency: maxConcurrency,
		trigger:        trigger,
		cycleToken:     cycleToken,
		done:           make(chan struct{}),
	}
}

// Start starts the scheduler's single owner goroutine.
func (scheduler *Scheduler) Start(ctx context.Context) error {
	if scheduler.prober == nil || scheduler.store == nil || len(scheduler.targets) == 0 ||
		scheduler.maxConcurrency < 1 || scheduler.maxConcurrency > 32 {
		return ErrInvalidScheduler
	}
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	if scheduler.started {
		return ErrAlreadyStarted
	}
	scheduler.started = true
	go scheduler.run(ctx)
	return nil
}

// Wait joins every goroutine owned by the scheduler.
func (scheduler *Scheduler) Wait() error {
	scheduler.mu.Lock()
	if !scheduler.started {
		scheduler.mu.Unlock()
		return ErrNotStarted
	}
	done := scheduler.done
	scheduler.mu.Unlock()
	<-done
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	return scheduler.err
}

// RunCycle synchronously probes and commits every configured target.
func (scheduler *Scheduler) RunCycle(ctx context.Context) ([]domain.Observation, error) {
	if scheduler.prober == nil || scheduler.store == nil || len(scheduler.targets) == 0 ||
		scheduler.maxConcurrency < 1 || scheduler.maxConcurrency > 32 {
		return nil, ErrInvalidScheduler
	}
	return scheduler.runTargets(ctx, scheduler.targets)
}

func (scheduler *Scheduler) run(ctx context.Context) {
	defer close(scheduler.done)
	for {
		if err := scheduler.trigger.Wait(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			scheduler.setError(fmt.Errorf("wait for scheduler trigger: %w", err))
			return
		}
		targets := scheduler.targets
		if selector, ok := scheduler.trigger.(TargetSelector); ok {
			targets = selector.Select(scheduler.targets)
		}
		if len(targets) == 0 {
			continue
		}
		if _, err := scheduler.runTargets(ctx, targets); err != nil {
			if errors.Is(err, domain.ErrCancelled) && ctx.Err() != nil {
				return
			}
			scheduler.setError(err)
			return
		}
	}
}

func (scheduler *Scheduler) runTargets(
	ctx context.Context,
	targets []domain.Target,
) ([]domain.Observation, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %w", domain.ErrCancelled, ctx.Err())
	case <-scheduler.cycleToken:
	}
	defer func() { scheduler.cycleToken <- struct{}{} }()

	type job struct {
		index  int
		target domain.Target
	}
	jobs := make(chan job, len(targets))
	results := make([]domain.Observation, len(targets))
	for index, target := range targets {
		jobs <- job{index: index, target: target}
	}
	close(jobs)

	workerCount := min(scheduler.maxConcurrency, len(targets))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for {
				var next job
				var ok bool
				select {
				case <-ctx.Done():
					return
				case next, ok = <-jobs:
					if !ok {
						return
					}
				}
				probeContext, cancel := context.WithTimeout(
					ctx,
					time.Duration(next.target.TimeoutMS)*time.Millisecond,
				)
				results[next.index] = scheduler.prober.Probe(probeContext, next.target)
				cancel()
			}
		}()
	}
	workers.Wait()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrCancelled, ctx.Err())
	}

	for index, observation := range results {
		if err := scheduler.store.Record(observation); err != nil {
			return nil, fmt.Errorf("%w: record target %q: %w", domain.ErrHistory, targets[index].Name, err)
		}
		committed, err := scheduler.store.History(targets[index].Name, 1)
		if err != nil {
			return nil, fmt.Errorf("%w: read target %q after record: %w", domain.ErrHistory, targets[index].Name, err)
		}
		if len(committed) != 1 {
			return nil, fmt.Errorf("%w: target %q was not retained after record", domain.ErrHistory, targets[index].Name)
		}
		results[index] = committed[0]
	}
	return results, nil
}

func (scheduler *Scheduler) setError(err error) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	scheduler.err = err
}

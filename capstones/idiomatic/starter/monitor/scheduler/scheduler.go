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

type Trigger interface {
	Wait(context.Context) error
}

type TargetSelector interface {
	Select([]domain.Target) []domain.Target
}

type WallClock interface {
	Now() time.Time
}

type ManualTrigger struct {
	placeholder struct{}
}

func NewManualTrigger() *ManualTrigger {
	return &ManualTrigger{}
}

func (trigger *ManualTrigger) Fire(ctx context.Context) error {
	return domain.ErrNotImplemented
}

func (trigger *ManualTrigger) Wait(ctx context.Context) error {
	return domain.ErrNotImplemented
}

type IntervalTrigger struct {
	placeholder struct{}
}

func NewIntervalTrigger() *IntervalTrigger {
	return &IntervalTrigger{}
}

func NewIntervalTriggerWithClock(clock WallClock) *IntervalTrigger {
	return &IntervalTrigger{}
}

func (trigger *IntervalTrigger) Wait(ctx context.Context) error {
	return domain.ErrNotImplemented
}

func (trigger *IntervalTrigger) Select(targets []domain.Target) []domain.Target {
	return nil
}

type Scheduler struct {
	placeholder struct{}
}

func New(
	prober probe.Prober,
	store history.Store,
	targets []domain.Target,
	maxConcurrency int,
	trigger Trigger,
) *Scheduler {
	return &Scheduler{}
}

func (scheduler *Scheduler) Start(ctx context.Context) error {
	return domain.ErrNotImplemented
}

func (scheduler *Scheduler) Wait() error {
	return domain.ErrNotImplemented
}

func (scheduler *Scheduler) RunCycle(ctx context.Context) ([]domain.Observation, error) {
	return nil, domain.ErrNotImplemented
}

// Package scheduler defines periodic probe ownership and lifecycle boundaries.
package scheduler

import (
	"context"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/probe"
)

// Trigger supplies deterministic cycle notifications.
type Trigger interface {
	Wait(context.Context) error
}

// Scheduler owns future probe-cycle goroutines.
type Scheduler struct{}

// New constructs the scheduler boundary.
func New(
	_ probe.Prober,
	_ history.Store,
	_ []domain.Target,
	_ int,
	_ Trigger,
) *Scheduler {
	return &Scheduler{}
}

// Start is intentionally incomplete and starts no goroutines.
func (*Scheduler) Start(context.Context) error {
	return domain.ErrNotImplemented
}

// Wait is intentionally incomplete and has no goroutines to join.
func (*Scheduler) Wait() error {
	return domain.ErrNotImplemented
}

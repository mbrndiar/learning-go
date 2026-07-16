package scheduler_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/scheduler"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/m3"
)

func TestRunCycleIsBoundedAndConfigurationOrdered(t *testing.T) {
	prober := newGatedProber()
	targets := []domain.Target{target("a"), target("b"), target("c"), target("d")}
	monitor := scheduler.New(prober, history.NewMemoryStore(10), targets, 2, scheduler.NewManualTrigger())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan cycleResult, 1)
	go func() {
		results, err := monitor.RunCycle(ctx)
		done <- cycleResult{results: results, err: err}
	}()

	m3.RequireSignal(t, ctx, prober.entered, "first worker")
	m3.RequireSignal(t, ctx, prober.entered, "second worker")
	if prober.maximum.Load() != 2 {
		t.Fatalf("maximum active probes = %d", prober.maximum.Load())
	}
	select {
	case <-prober.entered:
		t.Fatal("third probe started before a worker was released")
	default:
	}
	for range targets {
		prober.release <- struct{}{}
	}
	result := <-done
	if result.err != nil {
		t.Fatal(result.err)
	}
	for index, observation := range result.results {
		if observation.Target != targets[index].Name || observation.Sequence != int64(index+1) {
			t.Fatalf("results = %+v", result.results)
		}
	}
}

func TestRunCycleCommitOrderIgnoresCompletionOrder(t *testing.T) {
	prober := &orderedProber{
		entered: make(chan string, 3),
		release: map[string]chan struct{}{
			"a": make(chan struct{}), "b": make(chan struct{}), "c": make(chan struct{}),
		},
	}
	targets := []domain.Target{target("a"), target("b"), target("c")}
	monitor := scheduler.New(prober, history.NewMemoryStore(10), targets, 3, scheduler.NewManualTrigger())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan cycleResult, 1)
	go func() {
		results, err := monitor.RunCycle(ctx)
		done <- cycleResult{results: results, err: err}
	}()
	for range targets {
		select {
		case <-prober.entered:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	close(prober.release["c"])
	close(prober.release["b"])
	close(prober.release["a"])
	result := <-done
	if result.err != nil {
		t.Fatal(result.err)
	}
	for index, observation := range result.results {
		if observation.Target != targets[index].Name || observation.Sequence != int64(index+1) {
			t.Fatalf("results = %+v", result.results)
		}
	}
}

func TestRunCycleCancellationAndSerialization(t *testing.T) {
	prober := newGatedProber()
	monitor := scheduler.New(prober, history.NewMemoryStore(10), []domain.Target{target("a")}, 1, scheduler.NewManualTrigger())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := monitor.RunCycle(ctx)
		done <- err
	}()
	<-prober.entered
	cancel()
	err := <-done
	if !errors.Is(err, domain.ErrCancelled) {
		t.Fatalf("RunCycle() error = %v", err)
	}

	serialProber := newGatedProber()
	serial := scheduler.New(serialProber, history.NewMemoryStore(10), []domain.Target{target("a")}, 1, scheduler.NewManualTrigger())
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	twoDone := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := serial.RunCycle(ctx)
			twoDone <- err
		}()
	}
	<-serialProber.entered
	if serialProber.maximum.Load() != 1 {
		t.Fatalf("maximum active probes = %d", serialProber.maximum.Load())
	}
	serialProber.release <- struct{}{}
	<-serialProber.entered
	serialProber.release <- struct{}{}
	for range 2 {
		if err := <-twoDone; err != nil {
			t.Fatal(err)
		}
	}
}

func TestStartTriggerWaitAndLifecycle(t *testing.T) {
	store := &signallingStore{
		MemoryStore: history.NewMemoryStore(5),
		recorded:    make(chan struct{}, 1),
	}
	trigger := scheduler.NewManualTrigger()
	monitor := scheduler.New(immediateProber{}, store, []domain.Target{target("a")}, 1, trigger)
	if err := monitor.Wait(); !errors.Is(err, scheduler.ErrNotStarted) {
		t.Fatalf("Wait() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := monitor.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := monitor.Start(ctx); !errors.Is(err, scheduler.ErrAlreadyStarted) {
		t.Fatalf("second Start() error = %v", err)
	}
	if err := trigger.Fire(ctx); err != nil {
		t.Fatal(err)
	}
	waitContext, stopWaiting := context.WithTimeout(context.Background(), time.Second)
	defer stopWaiting()
	m3.RequireSignal(t, waitContext, store.recorded, "record")
	cancel()
	if err := monitor.Wait(); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if len(store.Current()) != 1 {
		t.Fatalf("current = %+v", store.Current())
	}
}

func TestManualAndIntervalTriggers(t *testing.T) {
	manual := scheduler.NewManualTrigger()
	ctx, cancel := context.WithCancel(context.Background())
	if err := manual.Fire(ctx); err != nil {
		t.Fatal(err)
	}
	cancelledContext, cancelImmediately := context.WithCancel(context.Background())
	cancelImmediately()
	if err := manual.Fire(cancelledContext); !errors.Is(err, context.Canceled) {
		t.Fatalf("Fire() error = %v", err)
	}
	if err := manual.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := manual.Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait() error = %v", err)
	}

	clock := &wallClock{now: time.Unix(0, 0)}
	interval := scheduler.NewIntervalTriggerWithClock(clock)
	targets := []domain.Target{targetWithInterval("a", 100), targetWithInterval("b", 200)}
	if err := interval.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if selected := interval.Select(targets); len(selected) != 2 {
		t.Fatalf("initial selected = %+v", selected)
	}
	clock.now = clock.now.Add(100 * time.Millisecond)
	if err := interval.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	selected := interval.Select(targets)
	if len(selected) != 1 || selected[0].Name != "a" {
		t.Fatalf("100ms selected = %+v", selected)
	}
	clock.now = clock.now.Add(100 * time.Millisecond)
	if err := interval.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	selected = interval.Select(targets)
	if len(selected) != 2 {
		t.Fatalf("200ms selected = %+v", selected)
	}

	cancelledContext, cancelImmediately = context.WithCancel(context.Background())
	cancelImmediately()
	clock.now = clock.now.Add(-50 * time.Millisecond)
	if err := interval.Wait(cancelledContext); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled interval Wait() error = %v", err)
	}
	if scheduler.NewIntervalTrigger() == nil {
		t.Fatal("NewIntervalTrigger() returned nil")
	}
}

func TestSchedulerValidationAndStoreFailure(t *testing.T) {
	invalid := scheduler.New(nil, nil, nil, 0, scheduler.NewManualTrigger())
	if _, err := invalid.RunCycle(context.Background()); !errors.Is(err, scheduler.ErrInvalidScheduler) {
		t.Fatalf("RunCycle() error = %v", err)
	}
	if err := invalid.Start(context.Background()); !errors.Is(err, scheduler.ErrInvalidScheduler) {
		t.Fatalf("Start() error = %v", err)
	}
	monitor := scheduler.New(immediateProber{}, failingStore{}, []domain.Target{target("a")}, 1, scheduler.NewManualTrigger())
	if _, err := monitor.RunCycle(context.Background()); !errors.Is(err, domain.ErrHistory) {
		t.Fatalf("RunCycle() error = %v", err)
	}
	monitor = scheduler.New(immediateProber{}, historyErrorStore{}, []domain.Target{target("a")}, 1, scheduler.NewManualTrigger())
	if _, err := monitor.RunCycle(context.Background()); !errors.Is(err, domain.ErrHistory) {
		t.Fatalf("RunCycle() history error = %v", err)
	}

	triggerFailure := errors.New("fixture trigger failure")
	monitor = scheduler.New(immediateProber{}, history.NewMemoryStore(2), []domain.Target{target("a")}, 1, errorTrigger{err: triggerFailure})
	if err := monitor.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := monitor.Wait(); !errors.Is(err, triggerFailure) {
		t.Fatalf("Wait() trigger error = %v", err)
	}
}

func TestRepeatedStartStopJoinsOwnedGoroutines(t *testing.T) {
	baseline := runtime.NumGoroutine()
	for range 25 {
		trigger := scheduler.NewManualTrigger()
		monitor := scheduler.New(immediateProber{}, history.NewMemoryStore(1), []domain.Target{target("a")}, 1, trigger)
		ctx, cancel := context.WithCancel(context.Background())
		if err := monitor.Start(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
		if err := monitor.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	runtime.GC()
	if current := runtime.NumGoroutine(); current > baseline+2 {
		t.Fatalf("goroutines grew from %d to %d", baseline, current)
	}
}

type cycleResult struct {
	results []domain.Observation
	err     error
}

type gatedProber struct {
	entered chan struct{}
	release chan struct{}
	active  atomic.Int32
	maximum atomic.Int32
}

func newGatedProber() *gatedProber {
	return &gatedProber{entered: make(chan struct{}, 20), release: make(chan struct{}, 20)}
}

func (prober *gatedProber) Probe(ctx context.Context, target domain.Target) domain.Observation {
	active := prober.active.Add(1)
	for {
		maximum := prober.maximum.Load()
		if active <= maximum || prober.maximum.CompareAndSwap(maximum, active) {
			break
		}
	}
	prober.entered <- struct{}{}
	select {
	case <-prober.release:
	case <-ctx.Done():
	}
	prober.active.Add(-1)
	return domain.Observation{Target: target.Name, CheckedAt: time.Unix(0, 0), Status: domain.StatusHealthy}
}

type orderedProber struct {
	entered chan string
	release map[string]chan struct{}
}

func (prober *orderedProber) Probe(ctx context.Context, target domain.Target) domain.Observation {
	prober.entered <- target.Name
	select {
	case <-prober.release[target.Name]:
	case <-ctx.Done():
	}
	return domain.Observation{Target: target.Name, CheckedAt: time.Unix(0, 0), Status: domain.StatusHealthy}
}

type immediateProber struct{}

func (immediateProber) Probe(_ context.Context, target domain.Target) domain.Observation {
	return domain.Observation{Target: target.Name, CheckedAt: time.Unix(0, 0), Status: domain.StatusHealthy}
}

type signallingStore struct {
	*history.MemoryStore
	recorded chan struct{}
}

func (store *signallingStore) Record(observation domain.Observation) error {
	if err := store.MemoryStore.Record(observation); err != nil {
		return err
	}
	select {
	case store.recorded <- struct{}{}:
	default:
	}
	return nil
}

type failingStore struct{}

func (failingStore) Record(domain.Observation) error {
	return errors.New("fixture store failure")
}

func (failingStore) Current() []domain.Observation {
	return nil
}

func (failingStore) History(string, int) ([]domain.Observation, error) {
	return nil, nil
}

type historyErrorStore struct{}

func (historyErrorStore) Record(domain.Observation) error {
	return nil
}

func (historyErrorStore) Current() []domain.Observation {
	return nil
}

func (historyErrorStore) History(string, int) ([]domain.Observation, error) {
	return nil, errors.New("fixture history failure")
}

type errorTrigger struct {
	err error
}

func (trigger errorTrigger) Wait(context.Context) error {
	return trigger.err
}

type wallClock struct {
	mu  sync.Mutex
	now time.Time
}

func (clock *wallClock) Now() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.now
}

func target(name string) domain.Target {
	return targetWithInterval(name, 100)
}

func targetWithInterval(name string, interval int) domain.Target {
	return domain.Target{Name: name, IntervalMS: interval, TimeoutMS: interval}
}

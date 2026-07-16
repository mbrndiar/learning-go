package history_test

import (
	"sync"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/m1"
)

func TestMemoryStoreTransitionsAndEviction(t *testing.T) {
	store := history.NewMemoryStore(2)
	checkedAt := time.Date(2026, 7, 16, 8, 0, 0, 999_000_000, time.FixedZone("offset", 3600))
	record(t, store, domain.Observation{Target: "a", CheckedAt: checkedAt, DurationMS: -1, Status: domain.StatusHealthy})
	record(t, store, domain.Observation{Target: "b", CheckedAt: checkedAt, Status: domain.StatusDegraded})
	record(t, store, domain.Observation{Target: "a", CheckedAt: checkedAt, Status: domain.StatusUnhealthy})

	current := store.Current()
	if len(current) != 2 || current[0].Target != "a" || current[1].Target != "b" {
		t.Fatalf("Current() = %+v", current)
	}
	latest := current[0]
	if latest.Sequence != 3 || latest.PreviousStatus != domain.StatusHealthy || !latest.Transition {
		t.Fatalf("latest a = %+v", latest)
	}
	if latest.DurationMS != 0 || latest.CheckedAt.Location() != time.UTC || latest.CheckedAt.Nanosecond() != 999_000_000 {
		t.Fatalf("normalized latest = %+v", latest)
	}
	aHistory, err := store.History("a", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(aHistory) != 1 || aHistory[0].Sequence != 3 {
		t.Fatalf("a history = %+v", aHistory)
	}
	bHistory, err := store.History("b", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(bHistory) != 1 || bHistory[0].Sequence != 2 {
		t.Fatalf("b history = %+v", bHistory)
	}
}

func TestMemoryStoreNoTransitionAndCopies(t *testing.T) {
	store := history.NewMemoryStore(3)
	httpStatus := 200
	code := domain.ErrorTransport
	record(t, store, domain.Observation{
		Target: "a", CheckedAt: time.Now(), Status: domain.StatusHealthy,
		HTTPStatus: &httpStatus, ErrorCode: &code,
	})
	record(t, store, domain.Observation{Target: "a", CheckedAt: time.Now(), Status: domain.StatusHealthy})
	observations, err := store.History("a", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 2 || observations[1].Transition || observations[1].PreviousStatus != domain.StatusHealthy {
		t.Fatalf("history = %+v", observations)
	}
	*observations[0].HTTPStatus = 500
	current := store.Current()
	if current[0].HTTPStatus != nil {
		t.Fatalf("latest status pointer = %v", current[0].HTTPStatus)
	}
	first, err := store.History("a", 2)
	if err != nil {
		t.Fatal(err)
	}
	if *first[0].HTTPStatus != 200 {
		t.Fatalf("stored HTTP status mutated: %d", *first[0].HTTPStatus)
	}
}

func TestMemoryStoreValidation(t *testing.T) {
	store := history.NewMemoryStore(0)
	if store.Limit() != 1 {
		t.Fatalf("Limit() = %d", store.Limit())
	}
	for _, observation := range []domain.Observation{
		{Status: domain.StatusHealthy},
		{Target: "a", Status: domain.StatusUnknown},
	} {
		err := store.Record(observation)
		m1.RequireErrorKind(t, err, domain.ErrHistory)
	}
	for _, limit := range []int{0, 2} {
		_, err := store.History("a", limit)
		m1.RequireErrorKind(t, err, domain.ErrInvalidLimit)
	}
	empty, err := store.History("missing", 1)
	if err != nil || len(empty) != 0 || empty == nil {
		t.Fatalf("missing history = %#v, %v", empty, err)
	}
}

func TestMemoryStoreConcurrentAccess(t *testing.T) {
	store := history.NewMemoryStore(100)
	var workers sync.WaitGroup
	for worker := range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range 100 {
				name := string(rune('a' + worker%4))
				_ = store.Record(domain.Observation{
					Target: name, CheckedAt: time.Unix(int64(index), 0), Status: domain.StatusHealthy,
				})
				_ = store.Current()
				_, _ = store.History(name, 10)
			}
		}()
	}
	workers.Wait()
	if len(store.Current()) != 4 {
		t.Fatalf("current targets = %d", len(store.Current()))
	}
}

func record(t *testing.T, store *history.MemoryStore, observation domain.Observation) {
	t.Helper()
	if err := store.Record(observation); err != nil {
		t.Fatal(err)
	}
}

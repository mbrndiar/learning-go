package pipeline

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestGenerateSendsAllValuesThenCloses(t *testing.T) {
	ch := Generate(context.Background(), 1, 2, 3)
	var got []int
	for v := range ch {
		got = append(got, v)
	}
	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestGenerateStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := Generate(ctx, 1, 2, 3, 4, 5)

	if first := <-ch; first != 1 {
		t.Fatalf("first value = %d, want 1", first)
	}
	cancel()

	// The channel must still eventually close even though not every value
	// was consumed; draining the rest must not hang forever.
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Generate did not close its channel after ctx was canceled")
	}
}

func TestMergeCombinesAllValues(t *testing.T) {
	ctx := context.Background()
	a := Generate(ctx, 1, 2)
	b := Generate(ctx, 3, 4)
	c := Generate(ctx, 5)
	merged := Merge(ctx, a, b, c)

	got := map[int]bool{}
	for v := range merged {
		got[v] = true
	}
	for _, want := range []int{1, 2, 3, 4, 5} {
		if !got[want] {
			t.Errorf("missing value %d in merged output %v", want, got)
		}
	}
	if len(got) != 5 {
		t.Errorf("merged output has %d unique values, want 5", len(got))
	}
}

func TestMergeStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	neverCloses := make(chan int) // deliberately never sent to or closed
	merged := Merge(ctx, neverCloses)

	cancel()

	done := make(chan struct{})
	go func() {
		for range merged {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Merge did not close its output channel after ctx was canceled")
	}
}

func TestRunBoundedRespectsMaxConcurrency(t *testing.T) {
	const maxConcurrency = 3
	var current, maxObserved int32

	jobs := make([]Job, 20)
	for i := range jobs {
		jobs[i] = Job{ID: i, N: i}
	}

	fn := func(ctx context.Context, j Job) (int, error) {
		n := atomic.AddInt32(&current, 1)
		for {
			old := atomic.LoadInt32(&maxObserved)
			if n <= old || atomic.CompareAndSwapInt32(&maxObserved, old, n) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		atomic.AddInt32(&current, -1)
		return j.N * 2, nil
	}

	results, err := RunBounded(context.Background(), jobs, maxConcurrency, fn)
	if err != nil {
		t.Fatalf("RunBounded returned error: %v", err)
	}
	if len(results) != len(jobs) {
		t.Fatalf("got %d results, want %d", len(results), len(jobs))
	}
	for i, r := range results {
		if r.ID != jobs[i].ID || r.Err != nil || r.Value != jobs[i].N*2 {
			t.Errorf("results[%d] = %+v, want ID=%d Value=%d", i, r, jobs[i].ID, jobs[i].N*2)
		}
	}
	if got := atomic.LoadInt32(&maxObserved); got > maxConcurrency {
		t.Errorf("observed max concurrency %d, want <= %d", got, maxConcurrency)
	}
}

var errBoom = errors.New("boom")

func TestRunBoundedPropagatesFirstError(t *testing.T) {
	jobs := make([]Job, 10)
	for i := range jobs {
		jobs[i] = Job{ID: i, N: i}
	}
	fn := func(ctx context.Context, j Job) (int, error) {
		if j.ID == 3 {
			return 0, errBoom
		}
		select {
		case <-time.After(50 * time.Millisecond):
			return j.N, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	start := time.Now()
	_, err := RunBounded(context.Background(), jobs, 4, fn)
	elapsed := time.Since(start)

	if !errors.Is(err, errBoom) {
		t.Fatalf("RunBounded error = %v, want errBoom", err)
	}
	if elapsed > time.Second {
		t.Errorf("RunBounded took %v after an error; the remaining jobs should be canceled quickly", elapsed)
	}
}

func TestRunBoundedRespectsContextTimeout(t *testing.T) {
	jobs := make([]Job, 5)
	for i := range jobs {
		jobs[i] = Job{ID: i, N: i}
	}
	fn := func(ctx context.Context, j Job) (int, error) {
		select {
		case <-time.After(time.Second):
			return j.N, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := RunBounded(ctx, jobs, 2, fn)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunBounded error = %v, want context.DeadlineExceeded", err)
	}
}

func TestRunBoundedPreservesOrder(t *testing.T) {
	jobs := make([]Job, 8)
	for i := range jobs {
		jobs[i] = Job{ID: i, N: i}
	}
	fn := func(ctx context.Context, j Job) (int, error) {
		// Sleep longer for earlier jobs so completion order differs from
		// input order; results must still line up with jobs by index.
		time.Sleep(time.Duration(len(jobs)-j.N) * time.Millisecond)
		return j.N * j.N, nil
	}
	results, err := RunBounded(context.Background(), jobs, 4, fn)
	if err != nil {
		t.Fatalf("RunBounded error: %v", err)
	}
	for i, r := range results {
		if r.ID != i || r.Value != i*i {
			t.Errorf("results[%d] = %+v, want ID=%d Value=%d", i, r, i, i*i)
		}
	}
}

// Package solution is the reference implementation for
// exercises/10_concurrency.
package solution

import (
	"context"
	"sync"
)

// Job is one unit of work: N is the input value, ID identifies its position
// so results can be matched back to inputs.
type Job struct {
	ID int
	N  int
}

// Result is the outcome of running a Job.
type Result struct {
	ID    int
	Value int
	Err   error
}

// Generate sends each of nums on the returned channel, in order, then
// closes it. It owns the returned channel and stops early, closing it
// promptly, if ctx is canceled before every value has been sent.
func Generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Merge fan-ins every channel in chans into one returned channel. It owns
// the returned channel and closes it, exactly once, once every input
// channel has been drained or ctx is canceled, whichever happens first. It
// never closes any of the input channels.
func Merge(ctx context.Context, chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(chans))

	for _, c := range chans {
		go func(c <-chan int) {
			defer wg.Done()
			for {
				select {
				case v, ok := <-c:
					if !ok {
						return
					}
					select {
					case out <- v:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// RunBounded runs fn for every job in jobs, allowing at most maxConcurrency
// calls to fn to be in flight at once, returning one Result per job in the
// same order as jobs.
//
// If any call to fn returns a non-nil error, RunBounded cancels a context
// derived from ctx so in-flight and not-yet-started calls stop promptly,
// waits for everything to finish, and returns that first error. If ctx
// itself is canceled or its deadline is exceeded, RunBounded returns
// ctx.Err() the same way.
func RunBounded(ctx context.Context, jobs []Job, maxConcurrency int, fn func(context.Context, Job) (int, error)) ([]Result, error) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]Result, len(jobs))
	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	recordErr := func(err error) {
		errOnce.Do(func() {
			firstErr = err
			cancel()
		})
	}

launchLoop:
	for i, job := range jobs {
		select {
		case <-runCtx.Done():
			recordErr(runCtx.Err())
			break launchLoop
		case sem <- struct{}{}:
			wg.Add(1)
			go func(i int, job Job) {
				defer wg.Done()
				defer func() { <-sem }()

				value, err := fn(runCtx, job)
				if err != nil {
					results[i] = Result{ID: job.ID, Err: err}
					recordErr(err)
					return
				}
				results[i] = Result{ID: job.ID, Value: value}
			}(i, job)
		}
	}

	wg.Wait()

	if firstErr != nil {
		return results, firstErr
	}
	return results, nil
}

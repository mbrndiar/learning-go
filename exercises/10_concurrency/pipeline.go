// Package pipeline exercises bounded concurrency, cancellation, error
// propagation across goroutines, and channel-ownership discipline: whichever
// goroutine creates a channel is the one responsible for closing it, and
// every goroutine that can block on a channel send/receive must also select
// on a context's Done channel so it does not leak when the consumer or
// caller gives up early.
package pipeline

import "context"

// Job is one unit of work: N is the input value, ID identifies its position
// so results can be matched back to inputs.
type Job struct {
	ID int
	N  int
}

// Result is the outcome of running a Job: exactly one of Value or Err is
// meaningful, matching Job.ID.
type Result struct {
	ID    int
	Value int
	Err   error
}

// Generate sends each of nums on the returned channel, in order, then closes
// it. Generate owns the returned channel: it is the only goroutine that ever
// sends on it or closes it. If ctx is canceled before every value has been
// sent, Generate stops sending and closes the channel promptly instead of
// blocking forever on a send nobody will receive.
//
// TODO(task 1): implement Generate. Run it in its own goroutine internally;
// select between sending the next value and ctx.Done() on every send.
func Generate(ctx context.Context, nums ...int) <-chan int {
	panic("not implemented")
}

// Merge fan-ins every channel in chans into one returned channel, preserving
// no particular order. Merge owns the returned channel and is responsible
// for closing it, exactly once, after every input channel has been drained
// (or ctx is canceled, whichever happens first). Merge does not close any of
// the input channels -- it did not create them, so it does not own them.
//
// TODO(task 2): implement Merge using one goroutine per input channel plus a
// sync.WaitGroup, and a final goroutine that closes the output channel once
// the WaitGroup completes. Respect ctx cancellation on both the receive from
// each input and the send to the output.
func Merge(ctx context.Context, chans ...<-chan int) <-chan int {
	panic("not implemented")
}

// RunBounded runs fn for every job in jobs, allowing at most maxConcurrency
// calls to fn to be in flight at once, and returns one Result per job in the
// same order as jobs (Results[i] corresponds to jobs[i]).
//
// If any call to fn returns a non-nil error, RunBounded stops launching new
// jobs, cancels the context passed to any still-running fn calls, waits for
// them to finish, and returns that first error (in job-index order; if two
// jobs fail "simultaneously" either of their errors may be returned, but it
// must be one of them, not a nil error). If ctx is canceled or its deadline
// is exceeded before all jobs complete, RunBounded returns ctx.Err().
//
// TODO(task 3): implement RunBounded using a semaphore channel of capacity
// maxConcurrency, a context.WithCancel derived from ctx so the first error
// (or ctx's own cancellation) stops sibling work, and a sync.WaitGroup.
// Each goroutine must write only to its own jobs[i]/results[i] index so no
// mutex is needed for the results slice; running with `go test -race` must
// report no data races.
func RunBounded(ctx context.Context, jobs []Job, maxConcurrency int, fn func(context.Context, Job) (int, error)) ([]Result, error) {
	panic("not implemented")
}

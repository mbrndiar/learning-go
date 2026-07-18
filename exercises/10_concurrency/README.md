# 🚦 Exercises: Concurrency

This exercise applies [Module 10](../../lessons/10_concurrency/README.md).
`pipeline` builds a small worker pipeline out of goroutines and channels,
covering the four ideas that make concurrent Go code trustworthy: bounded
concurrency, cancellation, error propagation, and channel ownership — then
proves it with `go test -race`.

## 🧩 Tasks

1. `Generate` — send each value on a channel it creates, then close that
   channel. **Channel ownership**: the goroutine that creates a channel is
   the one that closes it, and it must also stop early (without leaking) if
   the caller's `context.Context` is canceled before every value is sent.
2. `Merge` — fan-in several input channels into one output channel that
   `Merge` itself creates, closing it exactly once, after all inputs are
   drained or `ctx` is canceled. Merge never closes an input channel — it
   does not own those.
3. `RunBounded` — run a function for every job with at most `maxConcurrency`
   calls in flight at once (a semaphore channel is the idiomatic tool).
   Results must land back in input order, with each goroutine writing only
   to its own result slot so no mutex is needed.
4. Within `RunBounded`, implement **exception-equivalent error propagation**:
   Go has no exceptions, so the first error returned by any job's function
   must cancel a derived `context.Context` shared by the rest, stopping
   sibling work the way an exception would unwind concurrent tasks in other
   languages, and that first error must be the one `RunBounded` returns.
5. Also make `RunBounded` respect the caller's own `ctx`: a canceled or
   timed-out parent context must stop new work and surface `ctx.Err()`.

## ▶️ Commands

```bash
go test ./exercises/10_concurrency/...
go test -run '^$' ./exercises/10_concurrency
go test ./exercises/10_concurrency/solution
go test -race ./exercises/10_concurrency/solution/...
gofmt -l exercises/10_concurrency
```

## 📝 Notes

- A semaphore is just a buffered channel: `sem := make(chan struct{}, n)`;
  `sem <- struct{}{}` acquires a slot, `<-sem` releases it.
- `context.WithCancel` returns a `cancel` you can call once an error occurs,
  even though nothing "timed out" — cancellation is not only for timeouts.
- `sync.Once` is a clean way to capture only the *first* error among many
  goroutines that might fail concurrently.
- Writing to `results[i]` from goroutine `i` and never touching any other
  index is race-free without a mutex — sharing one mutable variable (like a
  single running total, or `append`ing to one slice from many goroutines)
  is what actually needs synchronization.
- `go test -race` instruments memory accesses at runtime; a clean pass does
  not prove correctness, but a race report always indicates a real bug.
- Compare with `solution/` only after a genuine attempt.

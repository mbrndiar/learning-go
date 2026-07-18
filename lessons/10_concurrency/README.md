# 🚦 Module 10: Concurrency

Go's most distinctive feature is how cheaply it lets you run things at the
same time, and how directly its tools let you reason about that safely. This
module builds up, one small runnable package at a time, from starting a
single goroutine to running a bounded, cancellable worker pool without
leaking a single goroutine or racing on shared memory.

## ✅ Prerequisites

Comfort with [Module 9](../09_tooling_cli_observability/README.md), especially
the race detector (`03_race_detector`), which this module uses throughout.

## 🎯 Objectives

By the end of this module you will be able to:

- start goroutines and explain why `main` does not wait for them by default;
- choose between unbuffered and buffered channels, and explain what each
  guarantees about blocking and ordering;
- write channel-typed function signatures (`chan<-`, `<-chan`) and follow the
  rule that only a channel's owning sender may close it;
- use `select` to wait on multiple channel operations, make a receive
  non-blocking with `default`, and bound a wait with a timeout;
- protect shared state with `sync.Mutex`, wait for a known number of
  goroutines with `sync.WaitGroup`, and use `sync/atomic` for simple
  lock-free counters;
- build a bounded worker pool that limits concurrency regardless of how many
  jobs there are;
- cancel in-flight work with `context.Context`, and prove with a
  synchronization point (not a guess or a sleep) that a goroutine actually
  stopped;
- recognize the goroutine-leak and data-race anti-patterns, and use
  `go test -race` to catch races before they reach production.

## 📖 Concurrency, explained

A **goroutine** is a function running independently of its caller, started
with the `go` keyword. Goroutines are cheap (a few kilobytes of stack that
grows as needed), so Go programs routinely run thousands of them. But cheap
does not mean automatic: `go f()` returns immediately, and `main` can exit
long before `f` finishes. Nothing "joins" a goroutine implicitly. You always
need an explicit completion signal: receive a result or dedicated `done`
value, close a channel owned by the goroutine, or wait with
`sync.WaitGroup.Wait()`. A closed `context.Done()` requests cancellation; it
does **not** prove the goroutine observed that request or returned.

A **channel** is a typed conduit for passing values between goroutines. It is
also the language's built-in synchronization primitive: an **unbuffered**
channel (`make(chan T)`) has no storage at all, so a send blocks until
another goroutine is ready to receive, and a receive blocks until another
goroutine sends. This handoff is what programs mean by "synchronous"
communication. A **buffered** channel (`make(chan T, n)`) can hold up to `n`
values without a receiver present; a send only blocks once the buffer is
full, and a receive only blocks once it is empty. Buffering decouples the
pace of the sender from the pace of the receiver, up to that limit.

Channel **direction** types describe intent at compile time:
`func produce(out chan<- int)` may only send on `out`, and
`func consume(in <-chan int) `may only receive from `in`. The compiler
rejects the wrong operation, catching a class of mistakes before you ever
run the program. **Closing** a channel (`close(ch)`) tells every receiver "no
more values are coming"; a `range` loop over a channel exits automatically
once it is closed and drained, and a receive using the comma-ok form
(`v, ok := <-ch`) reports `ok == false` once that happens. Ownership is a
strict rule: the sending side owns closure. With one producer it usually
closes the channel itself; with several producers, a coordinator closes it
after all senders finish. A receiver must not close a channel merely to stop
receiving. Closing twice or sending after close panics immediately.

`select` lets a goroutine wait on several channel operations at once and
proceeds with whichever becomes ready first; if more than one is ready
simultaneously, Go picks pseudo-randomly among them, so do not depend on
which one wins in that case. Adding a `default` case makes the whole `select`
non-blocking: it runs immediately if no other case is ready. Combining
`select` with `time.After`, `time.NewTimer`, or (preferably) a
`context.WithTimeout` deadline is the standard way to bound how long a
goroutine will wait for something.

`sync.Mutex` protects shared, mutable state: any goroutine must call `Lock`
before touching the protected data and `Unlock` (usually via `defer`)
afterward, so only one goroutine at a time can be inside that critical
section. `sync.WaitGroup` tracks a count of outstanding goroutines: call
`Add` before starting them, have each call `Done` when it finishes (again,
usually via `defer`), and call `Wait` to block until the count reaches zero.
`sync/atomic` (types like `atomic.Int64`) offers lock-free operations for a
single value, useful for simple counters and flags, but it only makes one
operation indivisible — reach for a `Mutex` as soon as you need to update
several related fields together or run a check-then-act sequence.

These primitives also define **happens-before** relationships: a channel send
happens before the matching receive completes, closing a channel happens
before a receive observes it closed, and a goroutine's `Done` happens before
`Wait` returns. Those guarantees make earlier writes visible after the
synchronization point. Time passing or `time.Sleep` creates no such guarantee.

A **bounded worker pool** starts a fixed number of worker goroutines that all
read from one shared jobs channel, capping how much work runs concurrently
no matter how many jobs are queued. This keeps memory and CPU use predictable
under load, which matters for both batch pipelines and request-handling
servers.

`context.Context` is the standard way to carry a cancellation signal (and
optional deadline) down a call chain. Any goroutine doing work that might
need to stop early should include `<-ctx.Done()` as a `select` case,
returning `ctx.Err()` when it fires. Always call the `cancel` function
returned by `context.With*` (typically via `defer`) to release the
context's resources, even if you never explicitly cancel it yourself.

A **goroutine leak** happens when a goroutine blocks forever — usually
sending on or receiving from a channel that will never be ready again — and
never returns, so it is never garbage collected. A **data race** happens
when two goroutines access the same memory location without synchronization
and at least one of them writes; the result is not reliably determined, and
the program's behavior can change between runs or hosts. `go test -race`
instruments your code to detect races at runtime; it does not detect
leaks, so leak prevention is a design discipline (always give every
goroutine a clear owner and a guaranteed way to stop) rather than a tool you
run.

## 🧭 Lessons

1. [`01_goroutines_basics/`](01_goroutines_basics/) — starting goroutines and
   why `main` needs a synchronization point to know when they finish.
2. [`02_unbuffered_channels/`](02_unbuffered_channels/) — the
   send-blocks-until-received rendezvous of an unbuffered channel.
3. [`03_buffered_channels/`](03_buffered_channels/) — capacity, `len`/`cap`,
   and how buffering decouples sender and receiver pace.
4. [`04_channel_direction_and_closing/`](04_channel_direction_and_closing/) —
   `chan<-`/`<-chan` signatures and the single-owner closing rule.
5. [`05_select_and_timeouts/`](05_select_and_timeouts/) — waiting on multiple
   channels, non-blocking receives, and timeout patterns.
6. [`06_waitgroup_and_mutex/`](06_waitgroup_and_mutex/) — waiting for many
   goroutines and protecting shared state.
7. [`07_atomic_counters/`](07_atomic_counters/) — lock-free counters with
   `sync/atomic`, and when to prefer a `Mutex` instead.
8. [`08_worker_pool/`](08_worker_pool/) — a bounded pool of workers sharing
   one jobs channel.
9. [`09_context_cancellation/`](09_context_cancellation/) — propagating
   cancellation and deadlines, and proving a goroutine actually stopped.
10. [`10_goroutine_leaks_and_races/`](10_goroutine_leaks_and_races/) —
    recognizing leaks and races, and the fixes that prevent them.

## ▶️ Running the lessons

Each lesson is its own runnable package:

```bash
go run ./lessons/10_concurrency/01_goroutines_basics
```

Run every lesson's tests, with the race detector on, from the repository
root:

```bash
go test -race ./lessons/10_concurrency/...
```

Try changing something small and re-running with `-race`: in
`04_channel_direction_and_closing`, remove the `chan<-`/`<-chan` direction
constraint from a function signature and see what compiles that shouldn't;
in `08_worker_pool`, change `workerCount` up or down and observe how the total
run time and interleaving of output change.

When you are ready to write a concurrent pipeline yourself, do the matching
exercise in
[`exercises/10_concurrency/`](../../exercises/10_concurrency/README.md).

## ⚠️ Common mistakes

- **Forgetting to synchronize before reading a result.** Starting a
  goroutine and immediately reading a value it was supposed to produce,
  without a channel receive or `WaitGroup.Wait()` in between, is a race:
  the goroutine may not have run yet.
- **Closing a channel from the receiving side.** Only the sender knows when
  the last value has been sent. A receiver that closes "to be tidy" can
  cause a `send on closed channel` panic elsewhere in the program.
- **Closing a channel twice, or sending after closing.** Both panic
  immediately; track ownership carefully, especially when several functions
  share a reference to the same channel.
- **Relying on which `select` case wins when several are ready.** The choice
  is pseudo-random by design; if you need one case prioritized, structure the
  code (or use nested selects) instead of assuming an order.
- **Using a `Mutex` correctly for reads but not writes (or vice versa).**
  Every access to shared state — read or write — must go through the lock;
  a single unguarded read is still a data race.
- **Believing tests pass because `go test` is quiet about races.** Always
  run concurrency tests with `-race`; a plain `go test` run can pass while
  still containing a genuine data race.
- **Forgetting to call the `cancel` function from `context.With*`.** Even if
  you never cancel manually, deferring `cancel()` releases resources
  associated with the context and its timers.
- **Proving a fix with `time.Sleep` instead of a real signal.** A sleep
  might "usually" be long enough, but it is not deterministic; wait on a
  result/done channel or `WaitGroup` that confirms completion. `ctx.Done()`
  alone confirms only that cancellation was requested.

## ❓ Review questions

1. Why does `go f()` not guarantee that `f` has run by the time the next
   line of `main` executes, and what are three ways to guarantee it has?
2. What is the practical difference between `make(chan int)` and
   `make(chan int, 5)` in terms of when a send blocks?
3. Why must only the sender close a channel, and what happens if a receiver
   closes it, or if the sender sends after closing it?
4. When would you add a `default` case to a `select`, and what does it
   change about the statement's blocking behavior?
5. What is the difference between what `sync.Mutex` protects and what
   `sync/atomic` protects? When would you choose one over the other?
6. In the worker pool lesson, why does closing the `jobs` channel not also
   close the `results` channel, and why is `results` closed from its own
   goroutine after `wg.Wait()`?
7. What does a goroutine need to do to respect `context` cancellation, and
   why is `ctx.Done()` usually one case inside a `select`?
8. How would you demonstrate, in a test, that a function does not leak a
   goroutine — without relying on a fixed `time.Sleep`?
9. What kind of bug does `go test -race` detect, and what kind of bug (also
   covered in this module) does it not detect?

## 🏁 Checkpoint

Combine everything in this module: implement a bounded worker pool that
accepts a `context.Context`, stops issuing new work and drains in-flight
workers cleanly when the context is canceled, reports results through a
channel closed exactly once, and passes `go test -race` with no leaked
goroutines.

## 🔗 Related reading

- <https://go.dev/ref/mem>
- <https://pkg.go.dev/sync>
- <https://pkg.go.dev/sync/atomic>
- <https://go.dev/blog/pipelines>
- <https://go.dev/blog/context>

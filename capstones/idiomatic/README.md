# 🐹 Idiomatic Go capstone — service health monitor

[`SPEC.md`](SPEC.md) defines the monitor behavior and five milestones. This
README describes the Go harness used to implement that contract.

## Go layout

```text
starter/monitor/       learner implementation target
solution/monitor/      reference implementation target
  domain/              configuration and JSON domain types
  probe/               Prober boundary and HTTP prober
  history/             HistoryStore boundary and bounded memory store
  scheduler/           trigger, lifecycle, and goroutine ownership boundary
  api/                 ordinary net/http Handler construction
  cli/                 testable check/serve process boundary
  cmd/monitor/         signal/os.Args/os.Exit wiring only
tests/contract/        reusable target-independent harness assertions
tests/fixtures/        reserved for deterministic milestone fixtures
tests/m1 ... tests/m5  reserved for progressive contract suites
```

Dependencies point inward: `domain` imports no monitor package; `probe` and
`history` depend only on `domain`; `scheduler` composes those interfaces; `api`
depends on `domain` and `history`; and `cli` is the outer application boundary.
The command imports `cli`. This arrangement keeps future concurrency and HTTP
work testable without import cycles.

## Commands

Compile the starter:

```bash
go test -run '^$' ./capstones/idiomatic/starter/...
```

Run only its harness smoke test:

```bash
go test ./capstones/idiomatic/starter/monitor
```

Run the solution target and shared support:

```bash
go test \
  ./capstones/idiomatic/solution/... \
  ./capstones/idiomatic/tests/...
```

Run either command target:

```bash
go run ./capstones/idiomatic/starter/monitor/cmd/monitor check --config monitor.json
go run ./capstones/idiomatic/solution/monitor/cmd/monitor check --config monitor.json
```

At the harness stage both commands intentionally print
`monitor: not implemented` to stderr and exit `1`. The API placeholder returns
HTTP `501`; no probes, scheduler goroutines, or servers are started.

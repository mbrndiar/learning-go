# 🐹 Idiomatic Go capstone — service health monitor

[`SPEC.md`](SPEC.md) defines the monitor behavior and five milestones. The
`solution` tree is a complete standard-library implementation; the matching
`starter` tree is the guided learner target.

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
tests/fixtures/        deterministic JSON and loopback HTTP fixtures
tests/m1 ... tests/m5  shared progressive contract assertions
```

Dependencies point inward: `domain` imports no monitor package; `probe` and
`history` depend only on `domain`; `scheduler` composes those interfaces; `api`
depends on `domain` and `history`; and `cli` is the outer application boundary.
The command imports `cli`. This arrangement keeps concurrency and HTTP behavior
testable without import cycles.

## Commands

Compile the starter:

```bash
go test -run '^$' ./capstones/idiomatic/starter/...
```

Run its harness smoke test:

```bash
go test ./capstones/idiomatic/starter/monitor
```

Run the solution target and shared support:

```bash
go test \
  ./capstones/idiomatic/solution/... \
  ./capstones/idiomatic/tests/...
```

Run race and coverage gates:

```bash
go test -race \
  ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...
go test -coverprofile=idiomatic-coverage.out \
  ./capstones/idiomatic/solution/monitor/...
bash scripts/check-coverage.sh idiomatic-coverage.out 85
```

Run either command target from the repository root:

```bash
go run ./capstones/idiomatic/starter/monitor/cmd/monitor check --config monitor.json
go run ./capstones/idiomatic/solution/monitor/cmd/monitor check --config monitor.json
go run ./capstones/idiomatic/solution/monitor/cmd/monitor \
  serve --config monitor.json --listen 127.0.0.1:8080
```

The starter command intentionally prints `monitor: not implemented` and its API
returns HTTP `501` until a learner completes the milestones. The solution
strictly validates configuration, probes with bounded concurrency and no
redirects or retries, retains bounded in-memory history, and shuts down all
owned scheduler/server work through contexts. It creates no snapshot or
database file.

The loopback API exposes:

- `GET /healthz`
- `GET /v1/targets`
- `GET /v1/history/{name}?limit=N`

Use `--json-errors` before `check` or `serve` for machine-readable command
errors. Successful `check` data is one JSON document on stdout; structured
server diagnostics use `slog` on stderr.

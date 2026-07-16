# Idiomatic capstone specification: service health monitor

## Status and interpretation

This is the learner contract for the required Go idiomatic capstone, equal in
weight to the comparative SQLite key/value capstone. Observable configuration,
probe classification, commands, HTTP responses, lifecycle behavior, and
acceptance criteria are normative. Package decomposition and concrete
goroutine/channel design are not, provided ownership and shutdown are explicit.

Both capstone solutions are complete and included in the course quality gates.
The superseded connected Task projects remain available in their
[last pre-removal snapshot](../../README.md#historical-task-project-migration)
as migration/reference examples, not as unfinished replacements.

## Bounded problem

Build a local monitor that:

- loads a validated set of HTTP targets;
- probes targets once or on a periodic schedule;
- classifies each observation as `healthy`, `degraded`, or `unhealthy`;
- records transitions from the previous state, initially `unknown`;
- retains a bounded in-memory history;
- emits a deterministic one-shot JSON report; and
- serves current state/history from a small `net/http` API.

Required fixtures are loopback `httptest` services. The solution is not a
production monitoring platform and does not discover or contact public systems.
State is derived from probes; it is not task CRUD or a generic SQLite KV API.

## Learning goals and course mapping

| Course material | Capstone outcome |
| --- | --- |
| [Modules 1–4](../../lessons/README.md) | Use Go values, control flow, functions, slices/maps, UTF-8 strings, and stable sorting for configuration and reports. |
| [Module 5: structs, methods, interfaces](../../lessons/05_structs_methods_interfaces/README.md) | Model targets/results/transitions and define small consumer-owned capabilities. |
| [Module 6: errors, files, JSON](../../lessons/06_errors_files_json/README.md) | Wrap explicit errors, validate JSON configuration, own response/file cleanup with `defer`, and preserve causes. |
| [Module 7: packages and generics](../../lessons/07_packages_and_generics/README.md) | Organize packages and use generic helpers only where they improve a real boundary. |
| [Module 8: testing](../../lessons/08_testing/README.md) | Write table tests, helpers/cleanup, `httptest`, examples, fuzz seeds, and coverage checks. |
| [Module 9: tooling, CLI, observability](../../lessons/09_tooling_cli_observability/README.md) | Use standard flags, `slog`, race detection, vet, staticcheck, and profiling-aware design. |
| [Module 10: concurrency](../../lessons/10_concurrency/README.md) | Own goroutines, bounded work, channels, `select`, timers, contexts, cancellation, joins, and race safety. |
| [Module 11: SQL and SQLite](../../lessons/11_sql_and_sqlite/README.md) | Define a narrow history boundary and make the capstone's bounded in-memory persistence choice explicit. |
| [Module 12: REST APIs and clients](../../lessons/12_rest_apis_and_clients/README.md) | Use client timeouts, HTTP handlers/middleware, JSON contracts, and graceful server shutdown. |

## Configuration contract

The UTF-8 JSON configuration shape is:

```json
{
  "schema_version": 1,
  "max_concurrency": 4,
  "history_limit": 20,
  "targets": [
    {
      "name": "catalog",
      "url": "http://127.0.0.1:9001/health",
      "interval_ms": 1000,
      "timeout_ms": 250,
      "expected_status_min": 200,
      "expected_status_max": 399,
      "max_body_bytes": 4096
    }
  ]
}
```

Rules:

- the top-level object and every target reject unknown or missing fields;
- only `schema_version: 1` is supported;
- `max_concurrency`: `1..32`; `history_limit`: `1..1_000`;
- `targets`: `1..100` entries;
- `name` matches `[A-Za-z0-9][A-Za-z0-9._-]{0,63}` and is unique;
- `url` is absolute `http` or `https`, has a host, contains no user information
  or fragment, and is preserved as supplied in reports;
- `interval_ms`: `100..86_400_000`;
- `timeout_ms`: `1..interval_ms`;
- expected status bounds are integers from `100..599` with min <= max;
- `max_body_bytes`: `0..1_048_576`; `0` means the body is not read;
- JSON numbers must be integers in the stated range; duplicates and trailing
  JSON values are invalid.

Required tests use only `http://127.0.0.1` or handler-level fakes. Supporting
real HTTPS uses the standard client trust store but custom certificates,
credentials, proxies, and redirect policy are out of scope. The HTTP prober
must not follow redirects; a 3xx is classified by the configured status range.

## Probe and transition semantics

One probe produces:

```json
{
  "sequence": 1,
  "target": "catalog",
  "checked_at": "2026-07-16T08:00:00.000Z",
  "duration_ms": 12,
  "status": "healthy",
  "previous_status": "unknown",
  "transition": true,
  "http_status": 204,
  "error_code": null,
  "message": "status 204 was within 200..399"
}
```

Classification is fixed:

- `healthy`: an HTTP response arrives before timeout, the body does not exceed
  `max_body_bytes`, and status is within the expected inclusive range;
- `degraded`: a complete bounded response arrives but status is outside the
  expected range;
- `unhealthy`: timeout, context cancellation during the probe, connection/DNS/
  TLS/read error, or body larger than the configured bound.

Required `error_code` values are `timeout`, `cancelled`, `transport_error`,
`body_read_error`, and `body_too_large`; otherwise it is `null`.
`http_status` is retained when a response was received, including a body error.
Messages are stable fixture strings but automation classifies by fields.

`checked_at` comes from an injected clock and is normalized to UTC millisecond
form. `duration_ms` is non-negative and comes from an injected elapsed-time
source in deterministic tests. No retry or hysteresis occurs.

Each monitor process assigns positive, contiguous observation sequences.
History is bounded globally to the most recent `history_limit` observations.
Current target state is the most recent observation for that target.
`transition` is true when `status != previous_status`, including
`unknown -> healthy/degraded/unhealthy`.

For observations completed concurrently, commit/report order is target order
from the configuration, not completion time. Each scheduling cycle finishes or
is cancelled before another cycle for the same monitor begins; cycles do not
overlap.

## Public Go boundary

Starter and solution must expose equivalent exported domain JSON types and these
behavioral capabilities, though concrete package paths may use the prescribed
layout below:

```go
type Prober interface {
    Probe(context.Context, Target) Observation
}

type HistoryStore interface {
    Record(Observation) error
    Current() []Observation
    History(target string, limit int) ([]Observation, error)
}
```

The scheduler must have a public start/run boundary accepting a context and a
deterministic trigger/ticker capability, and a completion boundary that lets
callers wait for every owned goroutine. The HTTP layer must expose an ordinary
`http.Handler`. Exact struct fields represented in JSON are normative; helper
types and algorithms are not.

## Observable CLI

Run from the repository root; `<impl>` is `starter` or `solution`:

```bash
go run ./capstones/idiomatic/<impl>/monitor/cmd/monitor \
  check --config PATH [--target NAME]

go run ./capstones/idiomatic/<impl>/monitor/cmd/monitor \
  serve --config PATH [--listen 127.0.0.1:8080]
```

`check` runs one cycle, in config order, and writes exactly one JSON document:

```json
{
  "checked_at": "2026-07-16T08:00:00.000Z",
  "summary": {"healthy": 1, "degraded": 1, "unhealthy": 0},
  "results": []
}
```

`results` contains observations in target order. `checked_at` is the earliest
observation time, or the injected cycle start when a fake produces no result.
Unhealthy/degraded targets are successful monitoring data and do not make
`check` fail. `--target` is repeatable, preserves configuration order, and an
unknown name is a usage error.

Both subcommands accept the global `--json-errors` option before the subcommand.

`serve` starts the scheduler and API under one root context. SIGINT/SIGTERM
handling belongs only to the thin main boundary: stop accepting requests,
cancel the scheduler and in-flight probes, wait for owned goroutines, then call
`http.Server.Shutdown`. A second signal may terminate immediately; that behavior
is not tested.

Structured diagnostics use `slog` on stderr. Successful command data never
shares stderr logging records.

## HTTP API

All responses are `application/json; charset=utf-8`. Required routes:

| Request | Response |
| --- | --- |
| `GET /healthz` | `200 {"status":"ok"}` while running; `503 {"status":"stopping"}` during shutdown |
| `GET /v1/targets` | `200 {"targets":[current-state...]}` in config order |
| `GET /v1/history/{name}?limit=N` | `200 {"target":"name","observations":[...]}` ascending by sequence |

Before its first observation, a target state is:

```json
{
  "target": "catalog",
  "status": "unknown",
  "observation": null
}
```

Afterward `observation` is the latest observation and `status` repeats its
status. `limit` defaults to the configured history limit and is `1..history_limit`.

Error shape:

```json
{"error":{"code":"target_not_found","message":"target \"missing\" was not configured"}}
```

Required mapping:

- `400`: malformed query (`invalid_limit`);
- `404`: unknown route or target (`not_found`, `target_not_found`);
- `405`: wrong method, with `Allow`;
- `500`: internal history/encoding error;
- `503`: service stopping.

Handlers set status before writing, escape dynamic values through JSON encoding,
and do not expose stack traces or local paths.

## Failure behavior and exits

Sentinel categories must support `errors.Is` where callers need classification;
wrapping must preserve causes.

| Exit | Meaning |
| --- | --- |
| `0` | check/serve completed normally; target health may be non-healthy |
| `2` | flag/usage error |
| `3` | invalid/unsupported configuration |
| `4` | configuration file I/O |
| `5` | monitor/server startup or internal failure |
| `130` | cancelled one-shot check before a complete report |

Required stable codes include `invalid_config`, `unsupported_schema`,
`duplicate_target`, `target_not_found`, `config_io`, `server_start`,
`history_error`, and `cancelled`. CLI errors are one JSON object on stderr when
`--json-errors` is supplied; otherwise they are concise text.

## Five guided milestones

### Milestone 1 — domain and probe contract

Implement validated configuration/domain types, pure classification,
transitions, explicit errors, a fake prober, and bounded in-memory history.

Acceptance:

- table tests cover every classification and configuration boundary;
- history eviction/current state are deterministic;
- errors wrap causes and retain sentinel classification;
- no goroutine or network is needed;
- `go test ./capstones/idiomatic/tests/m1/...` passes.

### Milestone 2 — one-shot application

Implement the standard-library HTTP prober, response bounds, context handling,
`check`, summaries, and stable JSON.

Acceptance:

- `httptest.Server` covers expected/unexpected status, timeout, oversized body,
  and read failure;
- bodies close on every path and redirects are not followed;
- results remain config ordered despite completion order;
- health states are data, while configuration/runtime failures use exact exits;
- milestone 2 contract tests pass.

### Milestone 3 — scheduler ownership

Implement periodic triggering, bounded worker concurrency, per-probe timeouts,
cycle serialization, cancellation, and joining.

Acceptance:

- a manual trigger starts cycles without wall-clock sleeps;
- active probes never exceed `max_concurrency`;
- cancellation unblocks sends/receives and every owned goroutine exits;
- `go test -race` reports no race in scheduler contracts;
- repeated start/stop tests show no growing goroutine count.

### Milestone 4 — history and API

Implement current state, bounded history, routes, JSON errors, middleware
logging, and server integration.

Acceptance:

- handlers are tested directly as `http.Handler`;
- route/method/query behavior and ordering match the contract;
- concurrent reads during recording are race-safe;
- stopping state returns `503` without panics;
- milestone 4 API/history contracts pass.

### Milestone 5 — operations and integration

Complete subprocess CLI tests, fake services with scripted responses, graceful
shutdown, fuzzed config decoding, examples, coverage, and repository tools.

Acceptance:

- child `check` output is semantically stable;
- loopback tests leave no server/goroutine leaks;
- config fuzzing never panics and valid seeds retain semantics;
- full normal and race suites pass on supported Go versions;
- coverage/staticcheck/govulncheck/link gates pass.

## Starter, solution, and test architecture

```text
capstones/idiomatic/
├── SPEC.md
├── starter/monitor/
│   ├── domain/
│   ├── probe/
│   ├── scheduler/
│   ├── history/
│   ├── api/
│   └── cmd/monitor/
├── solution/monitor/
│   └── ...matching public packages...
└── tests/
    ├── contract/
    ├── fixtures/
    └── m1/ ... m5/
```

Shared contract functions accept factories/interfaces and are called by thin
starter/solution `_test.go` wrappers. The starter contains complete exported
types, docs, constructors/signatures, and explicit TODO failures. CI compiles it
with `go test -run '^$'`; learner milestone tests intentionally fail until
implemented. The solution runs every contract.

## Deterministic fixtures and seams

Required fixtures:

- valid/minimal, duplicate-target, unknown-field, and unsupported-version config;
- scripted handler steps for healthy, degraded, timeout, oversized, and
  connection/read failures;
- expected one-shot report and API response JSON.

Required seams are `Prober`, cycle trigger/ticker, wall/elapsed clock,
`HistoryStore`, logger output, and optional server/listener factory. Use
`httptest`, `t.TempDir`, manually released channels, contexts, and cleanup
functions. No DNS, public network, real clock scheduling, random ports outside
loopback listeners, or elapsed-time performance assertions are permitted.

## Dependencies and supported runtime

- Module language version: Go `1.25.0`.
- CI/runtime matrix: Go `1.25.x` and `1.26.x`.
- Capstone runtime/test dependencies: Go standard library only.
- Existing `modernc.org/sqlite v1.53.0` remains pinned for the comparative
  capstone but is rejected for this capstone; history is bounded in memory to
  avoid a second database project.
- Existing tool pins used by CI remain `staticcheck@v0.7.0` and
  `govulncheck@v1.6.0`.
- Rejected: routers, schedulers, retry libraries, assertion suites, goroutine
  leak packages, Prometheus clients, and database packages. No new pin is
  proposed.

The required solution is portable across the Go-supported Linux, macOS, and
Windows environments. Signal tests may be Unix-specific smoke tests, but core
shutdown acceptance uses contexts and does not require a POSIX signal.

## Exclusions

No ICMP, DNS diagnostics, TLS certificate audit, authentication, cloud/service
discovery, alerts, retries, hysteresis, Prometheus, remote agents, persistent
history, user accounts, production daemon installation, distributed leadership,
or performance SLO is required.

## Quality and coverage commands

Focused:

```bash
go test ./capstones/idiomatic/tests/m1/...
go test ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...
go test -race ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...
```

Final repository validation must include the capstone in:

```bash
gofmt -l .
go vet ./...
go test ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...
go test -race ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...
go test -coverprofile=coverage.out \
  ./capstones/idiomatic/solution/monitor/...
bash scripts/check-coverage.sh coverage.out 85
go run ./tools/checklinks
```

Raw `go test ./...` is not the clean-checkout gate because it runs unfinished
exercise starter tests. CI also compiles both capstone starters, runs their
harnesses, and runs the pinned staticcheck and govulncheck commands described in
[`docs/QUALITY.md`](../../docs/QUALITY.md).

## Migration and reuse guidance

Reuse/refactor:

- context-aware HTTP clients, bounded bodies/timeouts, `httptest`, method-aware
  handlers, structured JSON errors, `slog`, and graceful shutdown from the
  historical `project/taskapi/` and `project/taskclient/` paths;
- consumer-owned interfaces, explicit error wrapping, deterministic contract
  helpers, and temporary resources from the historical
  `project/taskmanager/` path;
- worker ownership, channel closing, cancellation, and race-test patterns from
  the concurrency lessons/exercises.

Do not migrate Task validation, CRUD, storage, REST resource routes, or the Task
SQLite schema. The immutable comparison source is commit
[`b3211f9`](https://github.com/mbrndiar/learning-go/tree/b3211f99fc2ce5da54b88c59da3f12aacbed30ff/project)
at path `project/`; the live repository contains only the current capstones.

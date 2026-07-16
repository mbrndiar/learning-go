# 🔗 Connected Task Projects (legacy reference)

These three completed packages formed the course's earlier progressive
capstone:

- [`taskapi/`](taskapi/) owns remote task data in SQLite and exposes JSON over
  HTTP.
- [`taskclient/`](taskclient/) provides a reusable typed client and standalone
  CLI.
- [`taskmanager/`](taskmanager/) owns the domain model and selects local JSON or
  REST storage.

```text
Task Manager CLI -> Manager -> Storage
                             |-> FileStorage -> tasks.json
                             `-> RESTStorage -> Client -> API -> SQLiteStore
```

The projects demonstrate structs, interfaces, dependency injection, explicit
errors, atomic file writes, JSON, HTTP, contexts, SQLite, testing, and
concurrency without hiding the fundamentals behind frameworks.

The current learner capstones live under [`../capstones/`](../capstones/README.md).
Keep this code available for comparison and migration practice, but add new
capstone behavior to the comparative or idiomatic starter instead of extending
`project/`.

## 🧭 Old-to-new concept map

| Earlier Task concept | Current capstone destination | What carries forward |
| --- | --- | --- |
| [`taskmanager/task.go`](taskmanager/task.go) domain values and validation | [`capstones/idiomatic/.../domain`](../capstones/idiomatic/README.md) and [`capstones/comparative/.../domain`](../capstones/comparative/README.md) | Explicit values, validation at boundaries, sentinel/typed errors, and stable JSON—not the `Task` CRUD model. |
| [`taskmanager/storage.go`](taskmanager/storage.go) consumer-owned `Storage` interface | Comparative `storage.Store`/`storage.Opener`; idiomatic `probe.Prober` and `history.Store` | Small interfaces defined by consumers and fakeable contract seams. |
| [`taskmanager/contract_test.go`](taskmanager/contract_test.go) shared backend contract | Both [`tests/contract`](../capstones/README.md) trees and milestone packages | Reusable assertions with thin implementation wrappers; milestone suites now make progression explicit. |
| [`taskmanager/filestorage.go`](taskmanager/filestorage.go) schema migration and synchronized persistence | [Comparative SQLite milestones](../capstones/comparative/README.md) | Versioned migration, deterministic persistence, and contention testing; replace file/Task semantics with the frozen KV contract and real multi-process cases. |
| [`taskapi/store.go`](taskapi/store.go) SQLite store | Comparative solution `storage` package | Parameterized SQL, transactions, cleanup, and typed failures; do not copy the Task schema or CRUD routes. |
| [`taskapi/api.go`](taskapi/api.go) strict JSON HTTP server and graceful shutdown | Idiomatic `api` and `cli` packages | Method-aware `net/http`, bounded bodies, structured errors, finite timeouts, contexts, and graceful ownership. |
| [`taskclient/client.go`](taskclient/client.go) typed context-aware HTTP client | Idiomatic `probe.HTTPProber` | Bounded requests, response cleanup, status classification, and inspectable errors; monitoring observations replace Task resources. |
| The three thin [`cmd`](taskmanager/cmd/task-manager) packages | Comparative `app`/`cmd/kvstore` and idiomatic `cli`/`cmd/monitor` | Keep parsing and process exit wiring thin while testing importable application logic. |
| Task `httptest`, race, and coverage suites | Capstone contract, fixture, real-process, race, fuzz, and coverage gates | Deterministic fixtures and cleanup remain; the new suites add explicit milestones, fuzz targets, and subprocess conformance. |

The mapping is conceptual, not a package move. Preserve the old implementation
while using the new starter/solution boundaries and their specifications.

## 🚀 Run the retained projects

```bash
# Local JSON backend
go run ./project/taskmanager/cmd/task-manager add "Local task"

# Keep this API running in one terminal
go run ./project/taskapi/cmd/task-api

# Then use the HTTP client from another terminal
go run ./project/taskclient/cmd/task-client add "Remote task"
```

Each project directory contains its own architecture, usage, and testing
documentation. The historical staged extension list in
[`taskmanager/README.md`](taskmanager/README.md) remains useful for optional
migration practice across all three packages.

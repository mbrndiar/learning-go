# 🗂️ taskmanager

`taskmanager` owns the task **domain model** and coordinates persistence behind
a small, consumer-owned `Storage` interface. It is the top of the capstone:
a CLI drives a `Manager`, and the `Manager` delegates to whichever backend the
user selects — a local JSON file or the remote REST API.

## 🧭 Architecture

```text
task-manager CLI
      │
      ▼
   Manager ── validates titles/ids, wraps errors
      │
      ▼
   Storage (interface)
      ├── FileStorage  → atomic JSON document (tasks.json)
      └── RESTStorage  → taskclient.Client → task API → SQLite
```

- [`task.go`](task.go) — the validated `Task` value plus `NormalizeTitle` and
  `Task.Validate`. Titles are trimmed, UTF-8 checked, and length limited.
- [`storage.go`](storage.go) — the `Storage` interface and the shared
  `ErrTaskNotFound` sentinel. The interface lives here, where it is *consumed*,
  so backends stay decoupled from the domain package.
- [`manager.go`](manager.go) — `Manager`, the single entry point that applies
  domain rules before delegating and wraps every backend error with `%w`.
- [`filestorage.go`](filestorage.go) — `FileStorage`, an atomic JSON backend
  with a versioned schema, a monotonic `next_id`, legacy top-level-array
  migration, and mutex-guarded concurrent access.
- [`reststorage.go`](reststorage.go) — `RESTStorage`, an adapter that converts
  between the domain `Task` and the wire `Task` and translates the client's
  not-found sentinel into `ErrTaskNotFound`.
- [`cmd/task-manager`](cmd/task-manager) — the CLI, built on the standard
  `flag` package with a `run` function that is unit-tested directly.

Every `Storage` method takes a `context.Context`, so cancellation and timeouts
flow from the CLI all the way to disk or the network.

## 🔒 The storage contract

Both backends satisfy the same behavioral guarantees, enforced by a single
shared test suite ([`contract_test.go`](contract_test.go)):

- identifiers are **positive and strictly increasing**;
- removed identifiers are **never reused** (monotonic `next_id` / SQLite
  `AUTOINCREMENT`);
- missing tasks return an error satisfying `errors.Is(err, ErrTaskNotFound)`;
- blank titles are rejected before anything is stored.

## 🚀 Commands

Local JSON backend (default):

```bash
go run ./project/taskmanager/cmd/task-manager add "Write the report"
go run ./project/taskmanager/cmd/task-manager list
go run ./project/taskmanager/cmd/task-manager complete 1
go run ./project/taskmanager/cmd/task-manager remove 1
```

Remote REST backend (start the API from
[`../taskapi`](../taskapi) first):

```bash
go run ./project/taskmanager/cmd/task-manager -backend rest -url http://localhost:8080 add "Remote task"
go run ./project/taskmanager/cmd/task-manager -backend rest list
```

Flags: `-backend file|rest`, `-file <path>`, `-url <baseURL>`, `-timeout <dur>`.

## 🧪 Tests

```bash
go test ./project/taskmanager/...
go test -race ./project/taskmanager/...
go test -cover ./project/taskmanager
```

The tests cover the shared contract for both `FileStorage` and `RESTStorage`
(the latter through a real `httptest` server backed by SQLite), legacy-file
migration, schema validation, atomic saves, monotonic identifiers, concurrent
writes, context cancellation, and the CLI's file and REST paths.

## ✅ Learning checklist

- [ ] Explain why `Storage` is defined in the consumer package, not the backend.
- [ ] Trace a title from CLI input through validation to disk and to the API.
- [ ] Describe how atomic saves avoid a half-written file.
- [ ] Show why `next_id` guarantees identifiers are never reused.
- [ ] Explain how `errors.Is`/`errors.As` replace string comparisons here.
- [ ] Run the contract suite against a new backend of your own.

## 🧩 Staged extension exercises

Work these in order. Each stage must **keep the shared storage contract green**
(`go test ./project/taskmanager/...`) and avoid comparing error strings.

1. **Show a single task.** Add a `show <id>` command to the CLI that prints one
   task using the existing table renderer. No storage changes required.
2. **Search by title.** Add a `find <substring>` command that lists matching
   tasks. Do the filtering in the CLI first; then push it into `Manager` and
   compare the trade-offs.
3. **Bump the file schema to v2 with priorities.** Add a `Priority` field to
   `Task`, extend `Validate`, raise `currentSchemaVersion` to `2`, and migrate
   v1 documents on load by defaulting the new field. Keep reading legacy arrays.
4. **Add due dates.** Introduce a `time.Time` due date, parse it from the CLI in
   RFC 3339, and render it. Decide how the JSON and SQLite schemas represent it.
5. **Filter and sort `list`.** Add `-done`, `-pending`, and `-sort id|title`
   flags. Keep the default output stable so existing tests still pass.
6. **Add an in-memory backend.** Implement `MemoryStorage` and run
   `runStorageContract` against it. Confirm it upholds monotonic identifiers.
7. **Prevent lost updates.** Add an optimistic-concurrency version/ETag to
   `Task`, thread it through `Complete`, and reject stale writes with a new
   sentinel error. Update the API and client to carry the version.
8. **Observe the pipeline.** Add a `*slog.Logger` to `Manager` and log each
   operation with the request context. Propagate a request ID through the
   context from CLI to `FileStorage` and to the API handlers.

## 🔗 Related packages

- [`../taskclient`](../taskclient) — the typed client `RESTStorage` wraps.
- [`../taskapi`](../taskapi) — the SQLite-backed HTTP API behind the client.

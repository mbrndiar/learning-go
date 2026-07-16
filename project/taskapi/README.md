# 🌐 taskapi

`taskapi` is the remote end of the retained Task pipeline. It persists tasks in
**SQLite** through `database/sql` and exposes them as **JSON over HTTP** with a
small, framework-free `net/http` server.

This is a completed legacy reference. Current capstone work lives under
[`../../capstones/`](../../capstones/README.md); use the
[old-to-new concept map](../README.md#-old-to-new-concept-map) when reusing its
HTTP and SQLite patterns.

## 🧭 Architecture

```text
HTTP client (taskclient, curl, browser)
      │  JSON over HTTP
      ▼
   API (net/http) ── method-aware routes, validation, structured errors
      │
      ▼
   Store (interface)
      └── SQLiteStore → database/sql → modernc.org/sqlite (pure Go)
```

- [`store.go`](store.go) — `SQLiteStore`, opened with `OpenSQLiteStore`. It
  creates the schema, runs parameterized CRUD queries, and returns a typed
  `ErrNotFound`. Identifiers use `INTEGER PRIMARY KEY AUTOINCREMENT`, so they
  are monotonic and never reused after deletion.
- [`api.go`](api.go) — the `API` type and the consumer-owned `Store` interface.
  `Handler` builds method-aware routes (`GET`/`POST`/`DELETE`), enforces a
  request-body-size limit, rejects unknown JSON fields, and writes structured
  `{"error": "..."}` responses. `NewServer` wraps it with finite timeouts.
- [`cmd/task-api`](cmd/task-api) — the entry point. It wires the store to the
  handler, serves, and shuts down gracefully on `SIGINT`/`SIGTERM`.

The driver is the **pure-Go** `modernc.org/sqlite`, so no `cgo` or system
SQLite library is required.

## 🧾 HTTP routes

| Method & path              | Purpose                | Success status |
| -------------------------- | ---------------------- | -------------- |
| `GET /tasks`               | list all tasks         | `200`          |
| `POST /tasks`              | create a task          | `201`          |
| `GET /tasks/{id}`          | fetch one task         | `200`          |
| `POST /tasks/{id}/complete`| mark a task done       | `200`          |
| `DELETE /tasks/{id}`       | delete a task          | `204`          |

Errors use conventional codes: `400` for invalid input, `404` for a missing
task, `405` for an unsupported method, and `500` for internal failures (whose
details are logged, never leaked to the client).

## 🚀 Commands

```bash
go run ./project/taskapi/cmd/task-api
go run ./project/taskapi/cmd/task-api -addr :9090 -db ./tasks.db
```

Flags: `-addr <listen address>`, `-db <sqlite path or :memory:>`.

Exercise it with `curl`:

```bash
curl -s localhost:8080/tasks
curl -s -X POST localhost:8080/tasks -d '{"title":"Ship it"}'
curl -s -X POST localhost:8080/tasks/1/complete
curl -s -X DELETE localhost:8080/tasks/1 -i
```

## 🧪 Tests

```bash
go test ./project/taskapi/...
go test -race ./project/taskapi/...
go test -cover ./project/taskapi
```

The tests use a temporary/`:memory:` SQLite store and `httptest.Server` to
cover CRUD, the typed not-found error, monotonic identifiers, file persistence
across reopen, request validation, the body-size limit, method routing,
structured error bodies, and graceful shutdown of the command.

## ✅ Learning checklist

- [ ] Explain why the `Store` interface is defined next to the `API`.
- [ ] Describe how parameterized queries prevent SQL injection.
- [ ] Show how `AUTOINCREMENT` keeps identifiers monotonic after deletes.
- [ ] Explain how `errors.Is(err, ErrNotFound)` maps a store error to a `404`.
- [ ] Point out every finite timeout protecting the server.
- [ ] Describe what `signal.NotifyContext` plus `server.Shutdown` achieve.

## 🔗 Related packages

- [`../taskclient`](../taskclient) — the typed client for this API.
- [`../taskmanager`](../taskmanager) — the domain layer that consumes it.

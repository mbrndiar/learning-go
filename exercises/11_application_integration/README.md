# 🔗 Exercises: Application Integration

`taskapi` is a small task-tracking HTTP service that ties together JSON
validation, `httptest`-tested handlers and clients, context timeouts,
middleware, and a `database/sql` repository. `fakedb.go` is given,
complete infrastructure: a minimal in-memory `database/sql` driver (no
SQLite, no cgo, no external dependency) so `SQLTaskStore` can be exercised
through the real `database/sql` API. Do not edit it; everything else is
yours to implement.

## 🧩 Tasks

1. `Task.Validate` — a trimmed, non-empty `Title` of at most 200 characters,
   and, if `DueDate` is set, it must not be before the `now` passed in.
2. `SQLTaskStore.Create` / `.Get` / `.List` — the `database/sql` boundary:
   `ExecContext`/`QueryRowContext`/`QueryContext` against the fake driver,
   translating `sql.ErrNoRows` into `ErrNotFound`, and converting the
   nullable `due_date` column to/from `*time.Time`.
3. `NewServer` and its three handlers — `POST /tasks`, `GET /tasks`, and
   `GET /tasks/{id}` on a `*http.ServeMux` using Go's method-and-path
   patterns. Invalid JSON or a failed `Validate` must produce `400` with a
   JSON `{"error": "..."}` body, never a panic; a missing id must produce
   `404`.
4. `WithTimeout` — middleware that derives a `context.WithTimeout` from the
   request's context and swaps it in before calling `next`, demonstrating
   that a **context timeout** only helps if the code it wraps actually
   watches `ctx.Done()`.
5. `WithLogging` — middleware that records the response status (via a
   wrapping `http.ResponseWriter`) and logs one structured line per request.
6. `Client.CreateTask` / `.GetTask` — an HTTP client tested against
   `httptest.NewServer`, propagating `ctx` into every request and turning a
   `404` into an error matching `errors.Is(err, ErrNotFound)`.

## ▶️ Commands

```bash
go test ./exercises/11_application_integration/...
go test -run '^$' ./exercises/11_application_integration
go test ./exercises/11_application_integration/solution
go test -race ./exercises/11_application_integration/solution/...
gofmt -l exercises/11_application_integration
```

## 📝 Notes

- `fakedb.go` recognizes exactly three fixed query strings (see the
  `query*` constants); it is not a SQL parser, so `SQLTaskStore` must use
  those exact strings.
- `OpenFakeDB` returns a brand-new isolated in-memory database on every
  call via `sql.OpenDB` + a `driver.Connector` — no shared global state, so
  parallel and repeated (`go test -count=N`) test runs stay isolated.
- `driver.Value` (and therefore `driver.NamedValue.Value`) only carries
  `int64`, `float64`, `bool`, `[]byte`, `string`, `time.Time`, or `nil` —
  that is why the optional due date round-trips as an RFC 3339 string
  wrapped in `sql.NullString`, not as a raw `*time.Time`.
- `errors.Is(err, sql.ErrNoRows)` after `Scan` is the idiomatic way to
  detect "no such row"; do not compare errors with `==`.
- Go's `net/http` `ServeMux` supports `"METHOD /path/{param}"` patterns and
  `r.PathValue("param")` since Go 1.22 — no router dependency is needed.
- `http.NewRequestWithContext` is what makes a context deadline actually
  cancel an in-flight HTTP request on the client side.
- Compare with `solution/` only after a genuine attempt.

# 🔗 Module 11: Application Integration

This module connects the language and concurrency foundations from earlier
modules to the shape of a real service: routing HTTP requests, defining a
clean JSON boundary, composing cross-cutting behavior with middleware,
calling other services with a well-behaved client, thinking about
persistence in terms of an interface, and shutting a server down without
dropping requests. Every lesson uses only the standard library; the
capstone project plugs a real SQLite driver into the same storage
interface introduced here.

## 🎯 Objectives

By the end of this module you will be able to:

- register method-aware routes (`"GET /items/{id}"`) with `net/http` and
  read path segments with `(*http.Request).PathValue`;
- decode JSON request bodies strictly and encode JSON responses with
  correct status codes, headers, and a consistent error envelope;
- write middleware as plain `func(http.Handler) http.Handler` values and
  compose several of them into one chain;
- call another HTTP service with an `http.Client` that has both a default
  `Timeout` and a per-request `context` deadline, and always release
  response bodies;
- describe the `database/sql` programming model (queries, scanning rows,
  `sql.ErrNoRows`, contexts) and design a narrow storage interface around
  it, even without an external driver;
- explain why parameterized SQL prevents injection, and implement an
  atomic, all-or-nothing operation using the `Begin`/`Commit`/`Rollback`
  transaction pattern;
- shut an `http.Server` down gracefully, letting in-flight requests finish
  instead of dropping them.

## 📖 Application integration, explained

Since Go 1.22, `http.ServeMux` patterns can include an HTTP method prefix,
so `"GET /items/{id}"` and `"POST /items"` route to two different handlers
for the same path without any handler needing to branch on `r.Method`
itself. The `{id}` segment is captured automatically; a handler reads it
back with `r.PathValue("id")` instead of parsing `r.URL.Path` by hand. If a
path matches a registered pattern but not for the request's method, the mux
replies `405 Method Not Allowed` on its own.

Every HTTP handler has a **request/response boundary**: bytes coming in
must be decoded and validated before you trust them, and values going out
must be encoded consistently. `json.Decoder.DisallowUnknownFields` rejects
a body containing a field your server does not know about — usually a sign
the client is out of date. To reject trailing data, attempt a second decode
and require `io.EOF`; `Decoder.More` is for elements inside the current array
or object, not for validating a top-level request body. Set headers (like
`Content-Type`) and the status code with `WriteHeader` *before* writing any
response body; net/http
ignores header changes made afterward. A small, consistent error envelope
(for example `{"error": "..."}`) lets every client parse failures the same
way it parses success responses.

**Middleware** is nothing more than a function of type
`func(http.Handler) http.Handler`: it takes a handler, returns a new
handler that wraps it, and can run code before calling the wrapped
handler, after, or both. Logging, panic recovery, and request-scoped values
(passed through `context.WithValue` with an unexported key type, to avoid
collisions with other packages) are all naturally expressed this way.
Composing several middleware is just nested function calls: no framework
or special interface is required.

An `http.Client` making outbound requests should always have both a
default `Timeout` (so a client with no explicit context still cannot hang
forever) and, per call, a `context.Context` built with
`http.NewRequestWithContext` and typically `context.WithTimeout`. Whatever
happens, always `defer resp.Body.Close()` as soon as you have a non-nil
response — forgetting this is one of the most common resource and
connection-pool leaks in Go HTTP clients.

`database/sql` is a driver-agnostic API: you `Open` a `*sql.DB` connection
pool, then run `QueryContext`/`ExecContext`/`QueryRowContext` with a SQL
string containing placeholders and separate argument values, and read rows
back with `rows.Next()` / `rows.Scan(...)`, always checking `rows.Err()`
after the loop and always closing `rows`. A missing row is reported as the
sentinel error `sql.ErrNoRows`, checked with `errors.Is`, the same pattern
this module's `ErrNotFound` fake follows. Because this course avoids adding
an external dependency, this module models that same API shape — a narrow
`TaskStore` interface plus an in-memory fake that implements it — so the
lessons stay standard-library-only while still teaching the real
`database/sql` vocabulary. The capstone project satisfies the identical
interface with a real `*sql.DB` and a pinned pure-Go SQLite driver.

**Parameterized SQL** means passing values as separate arguments
(`db.ExecContext(ctx, "UPDATE accounts SET balance = ? WHERE name = ?", amount, name)`)
instead of concatenating them into the query string. The driver sends
parameters as data, never as part of the SQL text, which is what prevents
SQL injection for values. Placeholders cannot substitute table names, column
names, or SQL keywords; choose those from a fixed allowlist instead of user
input. A **transaction**, started with `db.BeginTx` and ended with
either `tx.Commit()` or `tx.Rollback()`, groups several statements into one
all-or-nothing unit: if any step fails, rolling back guarantees none of the
steps took effect, which matters whenever two or more writes must succeed
or fail together (like moving money between two accounts).

**Graceful shutdown** means an `http.Server` stops accepting new
connections and then waits — up to a bounded timeout — for requests already
in flight to finish, instead of severing them mid-response. The standard
pattern is `signal.NotifyContext` to turn an OS interrupt into a canceled
`context.Context`, and `server.Shutdown(ctx)` (with its own timeout) to
perform the drain once that context is done.

## 🧭 Lessons

1. [`01_http_routing_pathvalue/`](01_http_routing_pathvalue/) — method-aware
   routes and reading path segments with `PathValue`.
2. [`02_json_request_response/`](02_json_request_response/) — strict
   decoding, explicit encoding, and a consistent JSON error envelope.
3. [`03_middleware_functions/`](03_middleware_functions/) — middleware as
   composable `func(http.Handler) http.Handler` values.
4. [`04_http_client_context_timeout/`](04_http_client_context_timeout/) —
   calling another service with client timeouts and context deadlines.
5. [`05_database_sql_concepts/`](05_database_sql_concepts/) — the
   `database/sql` vocabulary, modeled with an interface and a teaching fake.
6. [`06_parameterized_sql_transactions/`](06_parameterized_sql_transactions/)
   — SQL parameters against injection, and the transaction pattern.
7. [`07_graceful_shutdown/`](07_graceful_shutdown/) — draining in-flight
   requests before an `http.Server` stops.

## ▶️ Running the lessons

Each lesson is its own runnable package:

```bash
go run ./lessons/11_application_integration/01_http_routing_pathvalue
```

Run every lesson's tests, with the race detector on, from the repository
root:

```bash
go test -race ./lessons/11_application_integration/...
```

## ⚠️ Common mistakes

- **Branching on `r.Method` inside one handler instead of registering
  separate method-aware patterns.** `"GET /items/{id}"` and
  `"POST /items"` keep each handler focused and let the mux return `405`
  automatically for unsupported methods.
- **Trusting a decoded JSON body without validating field values.**
  Successful decoding only means the JSON was well-formed, not that the
  values make sense (an empty title, a negative amount, and so on).
- **Setting response headers after calling `WriteHeader` or writing to the
  body.** Once the status code (or any body byte) is written, headers are
  locked in; set `Content-Type` and similar headers first.
- **Forgetting `resp.Body.Close()` on every code path**, including error
  responses. Leaving response bodies unclosed leaks connections and
  prevents the underlying transport from reusing them.
- **Giving an `http.Client` no `Timeout` at all.** A per-request context
  deadline is good practice, but a shared client with an unlimited default
  timeout will still hang forever on a request created without one.
- **Building SQL by concatenating strings instead of using parameters.**
  Even for "trusted" input, parameters are the only safe default; string
  concatenation of query text and values is the classic SQL injection bug.
- **Forgetting `defer tx.Rollback()` right after `BeginTx`.** Calling
  `Rollback` after a successful `Commit` is a documented no-op, so
  deferring it unconditionally is the safe default that guarantees cleanup
  on every early-return error path.
- **Calling `os.Exit` or letting `main` return immediately on shutdown**
  instead of calling `server.Shutdown` with a bounded context, which drops
  any request that was still being handled.

## ❓ Review questions

1. How does `"GET /items/{id}"` differ from registering just `"/items/{id}"`
   for a single handler that switches on `r.Method`, and what does
   `PathValue` give you that manual path parsing does not?
2. Why does `json.Decoder.DisallowUnknownFields` matter, and what kind of
   client mistake does it catch that ordinary decoding would silently
   accept?
3. Why must a handler set headers and the status code before writing any
   response body?
4. Write the type signature for a middleware function. Why does composing
   two middleware just mean calling one function with the other's result?
5. Why does an `http.Client` need both a `Timeout` field and a per-request
   `context.Context` deadline; what does each one protect against that the
   other does not?
6. What does `sql.ErrNoRows` (and this module's `ErrNotFound`) let a caller
   do with `errors.Is` that comparing an error string would not?
7. Why is `db.ExecContext(ctx, "... WHERE name = ?", name)` safe against SQL
   injection while string-concatenating `name` into the query is not?
8. In the transfer example, what guarantees that a failed transfer leaves
   both balances completely unchanged, and how does that map to
   `tx.Rollback()` against a real database?
9. What is the difference in behavior between an `http.Server` that is
   killed outright and one that is stopped with `server.Shutdown(ctx)`
   while a request is in flight?

## 🏁 Checkpoint

Begin the [idiomatic health-monitor capstone](../../capstones/idiomatic/README.md):
wire method-aware routes, a strict JSON boundary, an outbound HTTP probe with a
bounded context, an interface-backed history layer, and graceful shutdown into
one runnable service. Use the retained Task projects'
[old-to-new concept map](../../project/README.md#-old-to-new-concept-map) to
identify reusable patterns without copying the Task CRUD model.

# 📡 taskclient

`taskclient` is a **reusable, typed HTTP client** for the task API, plus a small
standalone CLI. It hides transport concerns — base-URL resolution, finite
timeouts, JSON encoding/decoding, response validation, and error translation —
so callers work with Go values and sentinel errors instead of raw responses.

This is a completed legacy reference. Current capstone work lives under
[`../../capstones/`](../../capstones/README.md); use the
[old-to-new concept map](../README.md#-old-to-new-concept-map) when reusing its
HTTP-client patterns.

## 🧭 Architecture

```text
your code / task-client CLI / RESTStorage
      │
      ▼
   Client ── base URL, *http.Client with a finite timeout
      │  context-aware List/Get/Add/Complete/Remove
      ▼
   task API (JSON over HTTP)
```

- [`client.go`](client.go) — the `Client`, its functional options
  (`WithHTTPClient`, `WithTimeout`), and every request method. It validates
  decoded tasks, resolves paths against a configurable base URL, and centralises
  transport in one `do` method.
- [`cmd/task-client`](cmd/task-client) — a `flag`-based CLI with a testable
  `run` function.

## 🧯 Error model

Errors are sentinels you match with `errors.Is`/`errors.As`, never strings:

- `ErrNotFound` — the task does not exist. An `*APIError` with a `404` status
  satisfies `errors.Is(err, ErrNotFound)` via its `Is` method.
- `APIError` — a structured non-2xx response that **preserves the HTTP status
  code** and the server's message.
- `ErrTimeout` — the request exceeded its deadline (translated from context
  deadlines and `net.Error` timeouts).
- `ErrUnavailable` — the API could not be reached (connection refused, DNS,
  reset, etc.).
- `ErrInvalidResponse` — the API returned malformed or invalid JSON.

Transport failures are wrapped with `%w` so the original cause (for example
`context.DeadlineExceeded`) remains inspectable.

## 🚀 Commands

Start the API from [`../taskapi`](../taskapi), then:

```bash
go run ./project/taskclient/cmd/task-client add "Remote task"
go run ./project/taskclient/cmd/task-client list
go run ./project/taskclient/cmd/task-client get 1
go run ./project/taskclient/cmd/task-client complete 1
go run ./project/taskclient/cmd/task-client remove 1
```

Flags: `-url <baseURL>` (default `http://localhost:8080`), `-timeout <dur>`.

## 🧑‍💻 Library usage

```go
client, err := taskclient.New("http://localhost:8080", taskclient.WithTimeout(5*time.Second))
if err != nil {
    return err
}

task, err := client.Add(ctx, "Write docs")
if errors.Is(err, taskclient.ErrTimeout) {
    // retry or surface a friendly message
}
```

## 🧪 Tests

```bash
go test ./project/taskclient/...
go test -race ./project/taskclient/...
go test -cover ./project/taskclient
```

The tests drive an `httptest.Server` to cover happy-path CRUD, `404`-to-
`ErrNotFound` translation, `APIError` status preservation, response validation,
base-path resolution, and timeout, network-unavailable, and cancellation
behavior — plus the CLI against a real API server.

## ✅ Learning checklist

- [ ] Explain why the client owns timeouts instead of its callers.
- [ ] Show how `APIError.Is` bridges a status code to a sentinel error.
- [ ] Describe how transport errors become `ErrTimeout` vs `ErrUnavailable`.
- [ ] Explain why decoded tasks are validated before being returned.
- [ ] Trace how a base-URL path prefix is preserved by `resolve`.

## 🔗 Related packages

- [`../taskapi`](../taskapi) — the API this client talks to.
- [`../taskmanager`](../taskmanager) — wraps this client as `RESTStorage`.

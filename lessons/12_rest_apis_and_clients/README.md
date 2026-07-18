# 🌐 Module 12: REST APIs and HTTP Clients

**Prerequisites:** Modules 1–11, especially JSON encoding/decoding and error
wrapping (Module 6), table-driven tests with `httptest` (Module 8),
`context.Context` cancellation and timeouts (Module 10), and hiding a
persistence boundary behind a narrow interface (Module 11). An HTTP handler
is really a JSON boundary wrapped in error-to-status mapping, an HTTP client
is a context-aware call across that same boundary from the other side, and
both are only trustworthy once every request, response, and connection is
proven to be cleaned up — the same discipline Module 11 applied to
`*sql.DB` and `*sql.Rows`.

This module builds focused HTTP boundaries with `net/http`, JSON, contexts,
middleware, clients, `httptest`, and graceful shutdown. It uses pinned Chi,
Gin, and Resty releases for direct comparison, but every lesson besides that
comparison stays standard-library only.

## 🎯 Objectives

By the end of this module you will be able to:

- trace one inbound request through decoding and shape validation, a domain
  operation, and error-to-status-and-JSON mapping, and explain why each step
  happens in that order;
- write route registration and middleware as adapters around
  `http.Handler`/`http.HandlerFunc`, composing cross-cutting behavior (like
  request IDs or centralized error mapping) without duplicating it in every
  handler;
- build an outbound HTTP client call that carries a `context.Context`,
  enforces a timeout, and validates both the response status and its decoded
  body before trusting it;
- explain client-side request/response body ownership, including why every
  successful `Do` result's response body must be closed regardless of whether
  decoding succeeds;
- implement the same route contract with `net/http`, Chi, and Gin, and
  explain what each framework's native API changes and what it deliberately
  leaves the same;
- implement graceful shutdown that stops accepting new requests and drains
  in-flight ones inside a bounded context, instead of dropping connections
  or hanging indefinitely;
- own finite HTTP resources explicitly: reuse clients, close response bodies,
  close test servers/listeners, and shut down servers deterministically;
- explain why this module's scope stops at one service boundary, deferring
  OpenAPI generation and cross-service interoperability to the Tasks
  project.

## 📖 HTTP boundaries, explained

An HTTP handler in this module always follows the same shape: **inbound
bytes → decode and validate shape → domain operation → error, status, and
JSON response.** `decodeJSON` in lesson 1 reads the request body through a
size-limited reader, calls `DisallowUnknownFields` so a typo or unexpected
field is rejected instead of silently ignored, and confirms the body holds
exactly one JSON value. Only after decoding succeeds does a handler validate
the *values* it received (a required field, a positive ID) — decoding proves
the request has the right shape; validation proves its content is usable.
Only after both succeed does the handler call into domain logic; whatever
that logic returns (a value or an error) is mapped to exactly one HTTP status
and one JSON body, so a caller never has to guess a response's meaning from
its shape alone.

Go expresses **routes and middleware as one `http.Handler` contract.**
`http.HandleFunc("METHOD /path", fn)` registers a method-aware route; a
middleware is a function that takes an `http.Handler` and returns one,
letting it run code before and after calling the next handler in the chain.
Both an entire router and a single route satisfy `http.Handler`, so
middleware, routers, and individual handlers compose without special cases:
`withRequestID(adapt(item))` in lesson 2 is three separate `http.Handler`
values wrapping one another. Centralizing error-to-status mapping in one
adapter (rather than repeating a `switch` in every handler) means adding a
new sentinel error to the mapping updates every route through it at once.

An outbound call is a mirror image of the same boundary, from the client's
side: build a **request carrying a `context.Context`** with
`http.NewRequestWithContext` so a caller-supplied deadline or cancellation
reaches the network call, send it through an `*http.Client` with its own
`Timeout` as a backstop, then validate the **response** before trusting it —
check the status code first, and only decode the body as JSON once the
status confirms the server actually returned the expected content. A
non-success status and a malformed body are different failure modes and
should produce different, specific errors (as `HTTPStatusClient.Check` in
the exercises does with `ErrUpstream`), not one generic "request failed."

On the client side, **body ownership** is explicit. The transport closes the
outbound request body; when `Client.Do` returns a response with no error, the
caller must close `resp.Body` even for a non-success status. Reading the body to
EOF before closing also gives the transport the best chance to reuse the
persistent connection. On the server side, `net/http` closes the inbound
request body after `ServeHTTP` returns, so a handler reads and bounds it but
does not need its own `defer r.Body.Close()`.

The **framework comparison** in lesson 4 implements one identical route —
the same path, same path parameter, same response — with `net/http`, Chi,
and Gin side by side. The HTTP contract observed by a client never changes;
what changes is each framework's native surface for expressing it:
`mux.HandleFunc("GET /items/{id}", ...)` and `r.PathValue("id")` for
`net/http`, `router.Get("/items/{id}", ...)` and `chi.URLParam(r, "id")` for
Chi, `router.GET("/items/:id", ...)` and `c.Param("id")` for Gin. Comparing
them directly is the point: a router is an implementation detail behind an
`http.Handler` (or, for Gin, an `http.Handler`-compatible engine), not a
change to what the API promises its callers.

**Graceful shutdown** means an `http.Server` stops *admitting new requests*
and then *drains in-flight ones*, both inside a bounded window, instead of
either dropping connections immediately or hanging forever. `server.Shutdown(ctx)`
closes listeners immediately (no new connections are accepted) and waits for
active handlers to finish, but only until its own context's deadline; a
handler that never returns can still force `Shutdown` to give up and report
an error once that deadline passes. Lesson 6 demonstrates the full sequence
with channels: cancel the outer context, confirm a request already
in-flight is allowed to finish and write its response, and confirm
`Serve` returns `http.ErrServerClosed` (not a real error) once shutdown
completes.

Finally, HTTP tests own **finite resources with explicit cleanup**, exactly
like the database tests in Module 11. An `httptest.NewServer` opens a real
listener on `127.0.0.1` and must be closed with `t.Cleanup(server.Close)`.
Clients should be reused rather than created per request; their response bodies
must be closed, and an owner that no longer needs a custom client's idle pool
may call `CloseIdleConnections`. Forgetting server or body cleanup can retain
sockets for the life of the test binary rather than just one test.

This module deliberately stops at one service's boundary: request shape,
domain error mapping, one outbound client call, and graceful shutdown of one
server. Describing that boundary formally (an OpenAPI/Swagger specification)
and coordinating it across multiple independently deployed services are
real, valuable skills — but they are deferred to the
[Tasks project](../../projects/tasks/), where a whole system, not one
lesson-sized package, makes that investment worthwhile.

## 🧭 Lessons

1. [`01_http_routing_and_json/`](01_http_routing_and_json/) — method-aware
   routes, path values, strict request JSON, and response envelopes.
2. [`02_middleware_and_error_mapping/`](02_middleware_and_error_mapping/) —
   composable middleware and centralized domain-error mapping.
3. [`03_http_client_context_timeout/`](03_http_client_context_timeout/) —
   context cancellation, client timeouts, body cleanup, and status handling.
4. [`04_router_framework_comparison/`](04_router_framework_comparison/) —
   runnable implementations of the same route with net/http, Chi, and Gin.
5. [`05_resty_client/`](05_resty_client/) — Resty client configuration,
   context propagation, JSON bodies, and typed responses.
6. [`06_graceful_shutdown/`](06_graceful_shutdown/) — signal cancellation and
   bounded draining of in-flight requests.

## ▶️ Running the lessons

Each lesson is its own runnable package:

```bash
go run ./lessons/12_rest_apis_and_clients/01_http_routing_and_json
go run ./lessons/12_rest_apis_and_clients/02_middleware_and_error_mapping
go run ./lessons/12_rest_apis_and_clients/03_http_client_context_timeout
go run ./lessons/12_rest_apis_and_clients/04_router_framework_comparison
go run ./lessons/12_rest_apis_and_clients/05_resty_client
go run ./lessons/12_rest_apis_and_clients/06_graceful_shutdown
```

Run every lesson's tests, then again with the race detector, from the
repository root:

```bash
go test ./lessons/12_rest_apis_and_clients/...
go test -race ./lessons/12_rest_apis_and_clients/...
```

## 🧪 Try it yourself

- In lesson 1, remove `decoder.DisallowUnknownFields()` from `decodeJSON`
  and send `{"name":"ok","extra":true}`: notice the request that should be
  rejected as malformed now succeeds instead.
- In lesson 3, change the request's context to `context.Background()`
  (dropping the timeout) against the test's `/status/slow` handler and
  observe the call hang instead of returning `context.DeadlineExceeded`.
- In lesson 6, shorten the shutdown timeout below how long the in-flight handler
  takes to finish and observe `server.Shutdown` return an error instead of
  draining cleanly. Separately remove `defer cancel()` and note that the simple
  run may still appear to work even though the timer resources are retained
  until the deadline — cleanup obligations are not always visible in output.

## ⚠️ Common mistakes

- **Validating field values before confirming the request's shape.**
  Decode and reject unknown fields or extra JSON values first; only trust
  decoded values enough to validate their content once decoding has already
  succeeded.
- **Writing a response header after calling `WriteHeader` (directly or via
  `Encode`).** Headers must be set before the status is written; setting
  `Content-Type` after the body has started writing has no effect.
- **Forgetting to close a response body, especially on an error path.**
  `defer resp.Body.Close()` belongs immediately after a successful `Do`/`Get`
  call, before checking the status code — not only after a successful
  decode.
- **Treating a non-2xx status and a malformed body as the same failure.**
  Check `resp.StatusCode` before decoding; a client that decodes first can
  report a confusing JSON error for what was actually an upstream 503.
- **Assuming a router change alters the HTTP contract.** Swapping `net/http`
  for Chi or Gin should not change a route's path, status codes, or response
  body — only how the handler is registered and reads path parameters.
- **Calling `os.Exit` or returning from `main` instead of `server.Shutdown`.**
  That drops in-flight connections immediately; graceful shutdown requires
  stopping new admissions and waiting for existing handlers, bounded by a
  separate shutdown context.
- **Leaving an `httptest.NewServer` unclosed in a test.** Always pair
  `httptest.NewServer(...)` with `t.Cleanup(server.Close)` so the listener
  is released even if the test fails partway through.

## ❓ Review questions

1. Why must a handler decode and validate a request's *shape* before
   validating the *values* it contains?
2. How does expressing middleware as `func(http.Handler) http.Handler` let
   `withRequestID(adapt(item))` compose three independent behaviors without
   any one of them knowing about the others?
3. Why does an outbound client need both a `context.Context` deadline and
   the `*http.Client`'s own `Timeout`, instead of relying on just one?
4. Why must a response body be closed even when its status code indicates
   an error, and even if the caller never reads it?
5. Why should a client check `resp.StatusCode` before attempting to decode
   the body as JSON?
6. In the framework comparison, what stays identical across the `net/http`,
   Chi, and Gin implementations, and what is each framework's own native
   surface for expressing the same route?
7. What does `server.Shutdown(ctx)` guarantee about new connections versus
   in-flight ones, and what happens if its context's deadline passes before
   an in-flight handler finishes?
8. Why does `Serve` returning `http.ErrServerClosed` after a graceful
   shutdown count as success rather than failure?
9. What do an `httptest.NewServer` and a `*sql.DB` from Module 11 have in
   common as resources under test, and how does each get cleaned up
   deterministically?
10. Why does this module stop at one service's request/response contract
    instead of also covering an OpenAPI specification or multi-service
    interoperability?

## 🏁 Checkpoint

Continue with [`exercises/12_rest_apis_and_clients/`](../../exercises/12_rest_apis_and_clients/README.md)
to build a strict-JSON route set with centralized error mapping and a
context-aware HTTP client of your own, then apply this whole module in the
required [Tasks project](../../projects/tasks/).

## 🔗 Related reading

- <https://pkg.go.dev/net/http>
- <https://pkg.go.dev/net/http/httptest>
- <https://pkg.go.dev/context>
- <https://go.dev/doc/database/cancel-operations>
- <https://pkg.go.dev/github.com/go-chi/chi/v5>
- <https://gin-gonic.com/en/docs/>
- <https://resty.dev/docs/>

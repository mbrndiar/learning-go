# 🌐 Exercises: REST APIs and HTTP Clients

This exercise applies [Module 12](../../lessons/12_rest_apis_and_clients/README.md).
Implement the TODOs in `api.go`. This is a boundary exercise, not another
full service or persistence project — `Store` and `StatusClient` are
injected dependencies the tests fake, so the work is entirely the HTTP
contract around them.

## 🔍 What this practices

- Inbound request bytes → strict-shape JSON decoding → value validation →
  domain call → error/status/JSON response, in that order.
- Route registration and one centralized place that maps sentinel errors to
  HTTP statuses.
- Parsing a path ID and an optional boolean query filter.
- Invoking an injected dependency from a handler instead of constructing one
  inline, so the handler stays testable with a fake.
- An outbound, context-aware HTTP client that distinguishes a non-success
  status from a malformed or incomplete body, and closes every response body
  it opens.
- `httptest.NewServer` with deterministic `t.Cleanup` in tests.

## 🧩 Tasks

- register method-aware routes and enforce one strict JSON request shape;
- validate values and emit consistent JSON error envelopes;
- parse a positive path ID and an optional boolean filter;
- map invalid, missing, upstream, and internal errors to appropriate
  statuses;
- invoke one injected `StatusClient` operation from a handler;
- implement a context-aware HTTP client that handles malformed and
  non-success responses;
- use `httptest.NewServer` with deterministic `t.Cleanup` in tests.

## ▶️ Commands

```bash
# Starter tests intentionally fail until implemented.
go test ./exercises/12_rest_apis_and_clients
go test ./exercises/12_rest_apis_and_clients/solution
go test -race ./exercises/12_rest_apis_and_clients/solution
```

## ⚠️ Common mistakes

- Decoding the request loosely (accepting unknown fields or extra JSON
  values) instead of rejecting a malformed shape with `400` before
  validating any value.
- Checking `resp.StatusCode` after decoding the body instead of before,
  which can turn an upstream `503` into a confusing JSON decode error.
- Forgetting `defer resp.Body.Close()` in `HTTPStatusClient.Check`, even on
  the non-success or malformed-body paths.

## 📮 Feedback

Compare with the matching files under `solution/` only after a genuine
attempt; a failing `go test` output tells you exactly which request shape,
status code, or error mapping still needs work.

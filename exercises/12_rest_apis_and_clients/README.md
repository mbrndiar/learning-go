# 🌐 Exercises: REST APIs and HTTP Clients

Implement the TODOs in `api.go`. This is a boundary exercise, not another full
service or persistence project.

## Tasks

- register method-aware routes and enforce one strict JSON request shape;
- validate values and emit consistent JSON error envelopes;
- parse a positive path ID and an optional boolean filter;
- map invalid, missing, upstream, and internal errors to appropriate statuses;
- invoke one injected `StatusClient` operation from a handler;
- implement a context-aware HTTP client that handles malformed and non-success responses;
- use `httptest.NewServer` with deterministic `t.Cleanup` in tests.

```bash
# Starter tests intentionally fail until implemented.
go test ./exercises/12_rest_apis_and_clients
go test ./exercises/12_rest_apis_and_clients/solution
go test -race ./exercises/12_rest_apis_and_clients/solution
```

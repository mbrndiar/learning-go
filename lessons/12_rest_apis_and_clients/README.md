# 🌐 Module 12: REST APIs and HTTP Clients

This module builds focused HTTP boundaries with `net/http`, JSON, contexts,
middleware, clients, `httptest`, and graceful shutdown.

## Lessons

1. [`01_http_routing_and_json/`](01_http_routing_and_json/) — method-aware routes, path values, strict request JSON, and response envelopes.
2. [`02_middleware_and_error_mapping/`](02_middleware_and_error_mapping/) — composable middleware and centralized domain-error mapping.
3. [`03_http_client_context_timeout/`](03_http_client_context_timeout/) — context cancellation, client timeouts, body cleanup, and status handling.
4. [`04_router_framework_comparison/`](04_router_framework_comparison/) — runnable implementations of the same route with net/http, Chi, and Gin.
5. [`05_resty_client/`](05_resty_client/) — Resty client configuration, context propagation, JSON bodies, and typed responses.
6. [`06_graceful_shutdown/`](06_graceful_shutdown/) — signal cancellation and bounded draining of in-flight requests.

## Focused dependencies

The comparison packages use pinned Chi, Gin, and Resty releases so every
example is real and runnable. The rest of the module remains standard-library
focused. OpenAPI tooling such as kin-openapi is intentionally deferred to the
later project phase.

```bash
go test ./lessons/12_rest_apis_and_clients/...
go test -race ./lessons/12_rest_apis_and_clients/...
```

Set response headers before status/body writes, validate decoded values, close
response bodies, bound outbound requests, and use `httptest` cleanup in tests.

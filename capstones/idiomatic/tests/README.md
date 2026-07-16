# Idiomatic test architecture

`contract/` contains the initial implementation-neutral harness. The milestone
packages contain focused shared assertions used by thin solution package tests:

- `m1`: strict domain validation, sentinels, transitions, and bounded history;
- `m2`: deterministic HTTP classification and response ownership;
- `m3`: bounded scheduler concurrency, cancellation, triggers, and joins;
- `m4`: handler routing, ordering, limits, errors, and stopping state;
- `m5`: CLI/process integration and JSON error envelopes.

`fixtures/` provides valid/invalid JSON documents, expected report/API JSON,
and a race-safe scripted loopback handler. Tests use `httptest`, temporary
directories, injected clocks, channels, and contexts; they require no public
network access.

The scheduler checks build on Module 10; the HTTP classification, handler,
client cleanup, JSON, and shutdown checks extend Module 12.

Run one milestone package or the complete solution:

```bash
go test ./capstones/idiomatic/tests/m1/...
go test ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...
```

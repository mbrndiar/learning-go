# Task REST API and clients

Build one Task application behind three HTTP server adapters and use it through
two HTTP client transports. The goal is one domain and one observable contract,
not five unrelated applications.

This required project belongs after Module 12 and before the final
[`capstones`](../../capstones/README.md). Finish the prerequisite course modules,
especially SQL/SQLite, HTTP/JSON, CLI, testing, contexts, and resource cleanup,
before starting.

Phase 1 provides matching, compileable `starter/` and `solution/` package trees,
the portable contract, and milestone-test package skeletons. Chi, Gin, and Resty
imports are intentionally deferred until their implementation milestones.

## Start with the contract

- [`docs/SPEC.md`](docs/SPEC.md) defines domain, persistence, HTTP, client, and
  failure behavior.
- [`docs/openapi.yaml`](docs/openapi.yaml) is the OpenAPI 3.1 HTTP contract.
- [`docs/PLAN.md`](docs/PLAN.md) is the reusable, language-neutral adaptation
  plan.
- [`docs/PROMPT.md`](docs/PROMPT.md) is the reusable agent prompt.

The specification and OpenAPI document are the behavioral source of truth.

## Go architecture

Both implementation roots have the same package layout and exported surface:

```text
projects/tasks/{starter,solution}/
├── task/                       domain and application boundaries
├── storage/{sqlite,markdown}/  persistence adapters
├── api/{nethttp,chi,gin}/      inbound HTTP adapters
├── client/{nethttp,resty}/     outbound HTTP transports
├── cli/                        shared command policy
├── server/                     server composition and lifecycle
└── cmd/{tasks-api,tasks}/      thin executable entry points
```

Dependencies point inward. `task` must not import an HTTP framework or client
library. Storage adapters depend on the domain boundary. API adapters translate
HTTP into service operations. Client transports implement the same remote
contract, and `cli` owns command parsing, output, and exit-code policy. Every
client must interoperate with every server; directory names are comparisons, not
pairings.

SQLite and the versioned Markdown checklist are independent stores. Switching
backends does not copy or synchronize data. Both must satisfy the same repository
contract, including monotonic IDs, restart persistence, and missing/corrupt data
behavior.

## Five milestones

1. **Domain and contracts** — Task values, validation, domain errors, repository
   and client boundaries, and the application service.
2. **Persistence** — SQLite and deterministic, versioned Markdown repositories
   passing one shared contract.
3. **Standard library** — a `net/http` server and `net/http` client with routing,
   JSON, status, timeout, and cleanup behavior visible.
4. **Chi and Resty** — thin Chi routes and an idiomatic Resty transport sharing
   only the core contract and command policy.
5. **Gin and interoperability** — thin Gin routes, black-box parity, OpenAPI
   comparison, and the complete two-client-by-three-server matrix.

Attempt each starter milestone before reading the corresponding solution.

## Intended commands

Run commands from the repository root. During Phase 1 the executable entry
points intentionally report `task.ErrNotImplemented`; later milestones replace
those placeholders.

Compile every project package without running future behavior tests:

```bash
go test -run '^$' ./projects/tasks/...
```

Run the exported-boundary check:

```bash
go test ./projects/tasks -run TestStarterAndSolutionExportedBoundariesMatch
```

Start a solution server with any adapter and backend:

```bash
go run ./projects/tasks/solution/cmd/tasks-api \
  --server nethttp --backend sqlite --data tasks.db --addr 127.0.0.1:8000

go run ./projects/tasks/solution/cmd/tasks-api \
  --server chi --backend markdown --data tasks.md --addr 127.0.0.1:8000

go run ./projects/tasks/solution/cmd/tasks-api \
  --server gin --backend sqlite --data tasks.db --addr 127.0.0.1:8000
```

Choose either client regardless of the running server:

```bash
go run ./projects/tasks/solution/cmd/tasks \
  --client nethttp --base-url http://127.0.0.1:8000 add "Learn REST"

go run ./projects/tasks/solution/cmd/tasks \
  --client resty --base-url http://127.0.0.1:8000 list --completed false
```

The CLI contract also includes `show`, `update`, `complete`, and `remove`; see
[`docs/SPEC.md`](docs/SPEC.md#command-line-client-contract).

## What the comparison should reveal

| Stack | Makes explicit | Adds or provides |
| --- | --- | --- |
| `net/http` server | Handler composition, method/path routing, JSON, headers, status selection, and lifecycle | Go's standard HTTP primitives |
| Chi | Middleware-oriented routing and route parameters | A small router designed around `net/http` |
| Gin | Framework request/response context and binding boundaries | Concise routing and framework-native middleware/testing patterns |
| `net/http` client | Request construction, contexts, transport errors, status checks, body ownership, and decoding | Go's standard HTTP client |
| Resty | Client/request configuration and response handling | A higher-level synchronous HTTP API |

Do not hide these differences behind a home-grown universal framework.

## Educational and non-production boundaries

Servers bind to loopback for local learning and tests use ephemeral ports,
finite timeouts, temporary storage, and no public network. This project does not
provide production deployment guidance, authentication, authorization, TLS
termination, containers, migrations, ORM use, distributed transactions,
cross-process Markdown locking, retries, generated SDKs, or operational
hardening. The SQLite adapter owns one process-scoped `*sql.DB`; its built-in
pool is deliberately configured for this local workload, while advanced pool
tuning, capacity planning, and production scaling remain out of scope. The
complete non-goals are in
[`docs/SPEC.md`](docs/SPEC.md#explicit-non-goals).

# Task REST API and clients

Build one Task application behind three HTTP server adapters and use it through
two HTTP client transports. The goal is one domain and one observable contract,
not five unrelated applications.

This required project belongs after Module 12 and before the final
[`capstones`](../../capstones/README.md). Finish the prerequisite course modules,
especially SQL/SQLite, HTTP/JSON, CLI, testing, contexts, and resource cleanup,
before starting.

The `solution/` tree is complete: all three servers, both clients, both
repositories, the CLI, OpenAPI checks, and the full interoperability matrix are
implemented. The matching `starter/` tree stays compileable and deliberately
incomplete so each milestone can be attempted before reading the solution.

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

## Starter and solution workflow

Run commands from the repository root. Start by compiling every learner package
without running milestone behavior:

```bash
go test -timeout=2m -run '^$' ./projects/tasks/starter/...
```

Then run the starter's no-side-effect harness. It verifies that unfinished
operations remain explicit and that commands do not create storage:

```bash
go test -timeout=2m -count=1 ./projects/tasks/starter/...
```

After completing a milestone, compare with the reference packages. The complete
solution command includes domain, repository, server, client, CLI, real-loopback,
OpenAPI, interoperability, and exported-boundary checks:

```bash
go test -timeout=3m -count=1 \
  ./projects/tasks/solution/... \
  ./projects/tasks/tests/... \
  ./projects/tasks
```

Start a solution server with any adapter and backend:

```bash
go run ./projects/tasks/solution/cmd/tasks-api \
  --server nethttp --backend sqlite --data tasks.db \
  --host 127.0.0.1 --port 8000

go run ./projects/tasks/solution/cmd/tasks-api \
  --server chi --backend markdown --data tasks.md \
  --host 127.0.0.1 --port 8000

go run ./projects/tasks/solution/cmd/tasks-api \
  --server gin --backend sqlite --data tasks.db \
  --host 127.0.0.1 --port 8000
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

## Project quality gates

CI runs the starter and complete solution on Go 1.25.x and 1.26.x. The current
Go job additionally race-tests the whole project and requires at least 85%
statement coverage across every non-command solution package. Thin
`cmd/tasks-api` and `cmd/tasks` composition entry points are excluded from the
coverage denominator; their selection and failure behavior still has direct
tests.

```bash
go test -race -timeout=5m -count=1 ./projects/tasks/...

task_packages=$(go list ./projects/tasks/solution/... | grep -v '/cmd/')
task_coverpkg=$(printf '%s\n' "$task_packages" | paste -sd, -)
go test -timeout=3m -count=1 \
  -coverpkg="$task_coverpkg" \
  -coverprofile=tasks-coverage.out \
  $task_packages \
  ./projects/tasks/tests/... \
  ./projects/tasks
bash scripts/check-coverage.sh tasks-coverage.out 85
rm tasks-coverage.out
```

See [`../../docs/QUALITY.md`](../../docs/QUALITY.md) for the complete course
gates, pinned analyzers, vulnerability scan, and link check.

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

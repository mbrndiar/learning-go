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

### Where to start

Start learning in [`starter/`](starter/). It is the project entry point for
milestone work: begin with the domain in `starter/task`, then follow the five
milestones into persistence, servers, clients, and interoperability. Use
[`solution/`](solution/) only as the reference implementation after attempting
the corresponding starter milestone.

The runtime entry points are separate from that learning path:

- `cmd/tasks-api` selects a persistence backend and HTTP adapter, then delegates
  composition and lifecycle to `server`.
- `cmd/tasks` selects an HTTP client transport, then delegates command parsing,
  output, and exit-code policy to `cli`.

`starter` and `solution` do not import or share production implementations. They
mirror the same exported boundaries, and the root
[`boundary_test.go`](boundary_test.go) verifies that those surfaces remain
aligned. Starter placeholders may have fewer imports until their milestone is
implemented, but their target dependency direction is the same as the solution.

### Package dependencies and shared boundaries

Arrows below mean "imports or delegates to" in the complete solution and in a
starter tree after the corresponding milestones are implemented:

```text
cmd/tasks
  ├──> cli ──> client ──> task
  │      └────────────────> task
  ├──> client/nethttp ──> client + task + client/internal/httpcontract
  └──> client/resty   ──> client + task + client/internal/httpcontract

cmd/tasks-api ──> server
                  ├──> api/{nethttp,chi,gin} ──> api ──> task
                  ├──> storage/{sqlite,markdown} ──────> task
                  └────────────────────────────────────> task
```

| Folder | Direct project dependencies | Used by | Shared responsibility |
| --- | --- | --- | --- |
| `task` | None | Storage, API, clients, CLI, and server composition | Domain values, errors, repository boundary, and service |
| `api` | `task` | All three API adapters | Transport-neutral DTOs, decoding, error mapping, and JSON responses |
| `api/{nethttp,chi,gin}` | `api` | `server` | Framework-specific routing and recovery for one HTTP contract |
| `storage/{sqlite,markdown}` | `task` | `server` | Alternative implementations of one repository contract |
| `client` | `task` | CLI, both client transports, and `cmd/tasks` | Client configuration, transport interface, and error types |
| `client/internal/httpcontract` | `client`, `task` | Both solution client transports | Strict URL, response, and protocol validation |
| `client/{nethttp,resty}` | `client`, `task`, shared HTTP contract in the solution | Solution `cmd/tasks` and interoperability tests; target wiring for the starter | Alternative implementations of one remote client contract |
| `cli` | `client`, `task` | `cmd/tasks` | Command parsing, JSON output, transport-independent behavior, and exit codes |
| `server` | API adapters, storage adapters, `task` | `cmd/tasks-api` | Backend/framework selection, resource ownership, and HTTP lifecycle |

The [`tests/`](tests/) packages are reusable contracts rather than a third
implementation. Milestone 1 assertions are imported by both starter and
solution tests. Later milestone contracts are reused across the solution's
alternative repositories, servers, and clients, while the root boundary test
compares starter and solution as a whole.

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

# Reusable implementation plan: Task REST API project

## How to use this plan

This is a language-independent build plan for recreating the Task REST API
applied project in another learning repository. It is not the execution plan
for one coding session or pull request.

Keep the observable behavior in [`SPEC.md`](SPEC.md) and
[`openapi.yaml`](openapi.yaml) stable. Adapt package names, source layout,
commands, and implementation techniques to the target ecosystem rather than
transliterating another language's code.

## Decisions to record first

Before editing, identify:

- the supported language/runtime versions;
- the lowest-level server API learners should understand;
- two representative server frameworks;
- the standard or lowest-level HTTP client and one or two client libraries,
  selecting two or three meaningful clients in total;
- the ordinary SQLite binding;
- formatter, linter, static analyzer, test runner, and coverage tool;
- conventions already used by the learning repository; and
- where the project belongs in the course prerequisite sequence.

Choose technologies that are maintained, locally reproducible, and meaningfully
different. Do not add wrappers with nearly identical teaching value merely to
reach a fixed count.

For example, the Go adaptation selects three servers (`net/http`, Chi, and Gin)
and two clients (`net/http` and Resty). That is an ecosystem-specific selection,
not a requirement that reusable adaptations use Go or exactly two clients.

## Required roles

Preserve these roles even when the target directory names differ:

```text
project/
├── README and contract documents
├── starter/
│   ├── shared core and repositories
│   ├── three server adapters
│   └── two or three client transports
├── solution/
│   └── the same public structure
└── shared tests and fixtures
```

The shared core owns the Task model, validation, application service, repository
abstraction, SQLite repository, and versioned Markdown repository. Server
adapters own inbound HTTP concerns. Client adapters own outbound HTTP concerns.
Dependencies point toward the core.

Tests select `starter` or `solution` without copying source into the harness.
The untouched starter should build or import successfully and fail incomplete
operations deliberately, not through broken wiring.

## Stage 1: contract and course placement

1. Read the target repository's guides, build configuration, existing projects,
   and current uncommitted changes.
2. Place the project after learners know basic SQL/SQLite and HTTP/JSON, but
   before advanced concurrency or final capstones.
3. Copy or adapt the human-readable specification and OpenAPI contract without
   changing the Task, persistence, status, error, or CLI behavior.
4. Write a learner README with prerequisites, architecture, milestones, exact
   commands, framework comparisons, and production-safety boundaries.
5. Add project-local dependency guidance using the ecosystem's normal format.

**Gate:** parse the OpenAPI document, check documentation links, and review the
written contract for internal consistency before implementation starts.

## Stage 2: domain and application boundaries

1. Create an immutable or read-only Task value.
2. Implement pure title, ID, update, and filter validation.
3. Define narrow validation, not-found, and storage errors.
4. Define the repository capability using the ecosystem's idiomatic abstraction:
   interface, trait, protocol, structural type, or injected functions.
5. Implement a small service that coordinates validation and repository calls.
6. Define the client transport boundary independently from any one HTTP library.
7. Add fakes and tests for normal, boundary, and failure behavior.

Avoid inheritance hierarchies, service locators, global mutable stores, and a
generic repository framework. The domain should not know which server or client
library is in use.

**Gate:** domain and service tests pass for both source roots; the core has no
framework or HTTP client imports.

## Stage 3: persistence adapters

Implement SQLite first:

1. initialize the small schema idempotently;
2. use parameterized SQL and explicit row mapping;
3. allocate monotonic IDs with the platform's SQLite mechanism;
4. use one transaction per mutation; and
5. close connections on success and failure.

Then implement the Markdown checklist:

1. parse and validate the version and `next-id` metadata;
2. parse every checklist row as untrusted input;
3. preserve ID ordering and monotonic allocation;
4. serialize deterministic UTF-8 output;
5. protect load-modify-save within one process; and
6. publish a complete sibling file with atomic replacement.

Run one parameterized repository contract against both backends. Include
restart persistence, deletion without ID reuse, malformed data, missing IDs,
and failed mutation tests.

**Gate:** both adapters are behaviorally interchangeable through the repository
boundary, without forced storage-specific base classes.

## Stage 4: lowest-level HTTP server and client

Build the low-level server before a framework:

1. inject the shared service into the handler or server factory;
2. make path matching, method handling, bytes, UTF-8, content length, JSON,
   headers, and status selection visible;
3. centralize mapping from domain/storage errors to the shared envelope;
4. log unexpected failures internally and sanitize `500` responses; and
5. provide explicit startup and shutdown ownership.

Build the lowest-level client:

1. share command parsing, output, exit codes, and response validation;
2. encode paths and queries rather than concatenating untrusted values;
3. set a finite timeout;
4. close responses;
5. distinguish API, malformed-response, and connection failures; and
6. avoid automatic retries.

Test the adapter in process where useful, then run a real loopback flow on an
ephemeral port.

**Gate:** the server passes the HTTP contract and the client passes the shared
client contract with no public network access.

## Stage 5: first framework and client library

Use the first framework's native strengths:

- construct dependencies through an application factory or equivalent;
- keep route handlers thin;
- use centralized error handlers;
- use the framework's supported local test client; and
- retain the shared domain and error behavior.

Use the first client library's normal session/client lifetime and response APIs.
Do not make it mimic the low-level client internally. It should share only the
command policy and transport contract that are genuinely common.

**Gate:** both servers pass the same black-box contract, both clients pass the
same client contract, and each client can call each server.

## Stage 6: second framework and any remaining client

Use boundary request/response models and dependency injection where the
framework supports them. Keep those models at the adapter boundary rather than
turning framework types into domain values.

Map framework validation failures to the shared `422` envelope. Compare
generated OpenAPI semantically with the checked-in language-neutral contract;
generated framework output does not replace it.

If the ecosystem selection includes a third client, implement it idiomatically
here, with explicit client lifetime and timeouts. Synchronous behavior is
sufficient even if the library also supports asynchronous requests.

**Gate:** all three servers and every selected client pass their shared
contracts.

## Stage 7: interoperability and learner progression

Use a small, representative matrix rather than every backend/process
combination:

| Client | Low-level server | Framework A | Framework B |
| --- | --- | --- | --- |
| Low-level client | Required | Required | Required |
| Client A | Required | Required | Required |
| Client B, when selected | Required | Required | Required |

Use both repositories elsewhere in the suite so persistence parity remains
visible. Keep all network tests on loopback, use finite timeouts, allocate
ephemeral ports, and guarantee cleanup.

Organize learner tests into five milestones:

1. domain and contracts;
2. persistence;
3. low-level server and client;
4. first framework and client; and
5. second framework, any remaining client, and interoperability.

Starter checks should provide actionable incomplete messages. Solution checks
cover normal, boundary, corruption, protocol, and cleanup failures.

## Stage 8: repository integration

1. Add project dependencies to clean-install and CI flows without weakening
   existing course checks.
2. Extend formatting, linting, static analysis, and type-checking scopes.
3. Measure project coverage separately so mature projects cannot hide gaps.
4. Link prerequisite lessons to the project and the project to later modules and
   capstones.
5. Run documented commands from a clean environment.
6. Scan for generated databases, Markdown state files, logs, credentials, and
   other local artifacts before committing.

**Final gate:** both source roots have valid wiring; all solution tests and
quality tools pass; OpenAPI and links validate; every client interoperates with
every server; and unrelated user changes remain untouched.

## Adaptation guardrails

- Preserve behavior, not language-specific names.
- Prefer official documentation for version-sensitive framework behavior.
- Use the target ecosystem's normal error, resource, package, and test patterns.
- Keep framework lifecycle code visible enough to teach.
- Do not introduce production infrastructure that is outside the learning goal.
- Do not add authentication, pagination, asynchronous execution, ORMs,
  migrations, checksum manifests, or generated SDKs.
- If a framework has unavoidable default behavior that conflicts with the
  contract, normalize it at the adapter boundary and test the result.

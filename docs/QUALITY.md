# ✅ Course Verification and CI

Run commands from the repository root. The canonical automation is
[`.github/workflows/course.yml`](../.github/workflows/course.yml), which tests
Go 1.25.x and 1.26.x. The `.x` selectors intentionally resolve the latest
security patch in each supported release line.

## Why `go test ./...` fails in a learner checkout

The twelve packages directly under `exercises/` are learner starters. Their
tests describe the work and intentionally fail while the TODO implementations
remain. Therefore a raw:

```bash
go test ./...
```

is useful after completing every exercise, but it is not the clean-checkout
health command. The capstone starters are different: their small harnesses
recognize the explicit `ErrNotImplemented`/HTTP 501 placeholders, while CI also
type-checks every starter package with `go test -run '^$'`. The Tasks starter
uses the same compile-plus-harness pattern and additionally proves placeholder
commands have no storage side effects.

CI keeps learner intent separate from repository health:

```bash
go test ./lessons/...
while IFS= read -r package; do
  [[ -z "$package" ]] && continue
  go run "$package"
done < <(
  go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./lessons/... |
    sed '/^$/d'
)

go list ./exercises/... |
  grep -v '/solution$' |
  xargs -r go test -run '^$'
go list ./exercises/... |
  grep '/solution$' |
  xargs -r go test

go test -timeout=2m -run '^$' ./projects/tasks/starter/...
go test -timeout=2m -count=1 ./projects/tasks/starter/...
go test -timeout=3m -count=1 \
  ./projects/tasks/solution/... \
  ./projects/tasks/tests/... \
  ./projects/tasks

go test -run '^$' \
  ./capstones/comparative/starter/... \
  ./capstones/idiomatic/starter/...
go test \
  ./capstones/comparative/starter/kvstore \
  ./capstones/comparative/solution/... \
  ./capstones/comparative/tests/... \
  ./capstones/idiomatic/starter/monitor \
  ./capstones/idiomatic/solution/... \
  ./capstones/idiomatic/tests/... \
  ./capstones
```

## Module, formatting, vet, and links

```bash
go mod download
go mod tidy
git diff --exit-code -- go.mod go.sum

test -z "$(gofmt -l .)"
go vet ./...
go run ./tools/checklinks
```

The link checker validates repository-local Markdown targets. External URLs and
heading fragments are intentionally outside its offline scope.

## Race detector

The Go 1.26 CI job race-checks the code that owns concurrency, storage, HTTP, or
process boundaries:

```bash
go test -race \
  ./capstones/comparative/tests/... \
  ./capstones/idiomatic/solution/monitor/... \
  ./capstones/idiomatic/tests/...

mkdir -p capstones/comparative/.conformance/race
go build -race \
  -o capstones/comparative/.conformance/race/kvstore \
  ./capstones/comparative/solution/kvstore/cmd/kvstore
COMPARATIVE_KV_PROGRAM="$PWD/capstones/comparative/.conformance/race/kvstore" \
  go test -race -count=2 ./capstones/comparative/solution/kvstore
rm -rf capstones/comparative/.conformance/race

go test -race ./lessons/10_concurrency/...
go test -race ./lessons/11_sql_and_sqlite/...
go test -race ./lessons/12_rest_apis_and_clients/...
go test -race ./exercises/10_concurrency/solution
go test -race ./exercises/11_sql_and_sqlite/solution
go test -race ./exercises/12_rest_apis_and_clients/solution
go test -race -timeout=5m -count=1 ./projects/tasks/...
```

It does not use `go test -race ./...`, because that would run the intentionally
failing exercise starter tests and would obscure the defined race surface. The
Tasks project itself is fully race-tested, including its starter harness,
solution packages, shared contracts, real-loopback lifecycle, and complete
client/server matrix.

## Coverage

CI enforces 85% for every non-command Tasks solution package together, 85% for
the idiomatic monitor, and 75% for the comparative command/process
implementation. Tasks coverage instruments the complete substantive solution
surface and exercises it through unit, contract, OpenAPI, lifecycle, and
interoperability tests; only the two thin `cmd` entry-point packages are
excluded:

```bash
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

go test -coverprofile=idiomatic-coverage.out \
  ./capstones/idiomatic/solution/monitor/...
bash scripts/check-coverage.sh idiomatic-coverage.out 85
```

The comparative capstone is measured through its real executable boundary:

```bash
mkdir -p \
  capstones/comparative/.coverage/bin \
  capstones/comparative/.coverage/data
go build -cover \
  -coverpkg=./capstones/comparative/solution/... \
  -o capstones/comparative/.coverage/bin/kvstore \
  ./capstones/comparative/solution/kvstore/cmd/kvstore
COMPARATIVE_KV_PROGRAM="$PWD/capstones/comparative/.coverage/bin/kvstore" \
GOCOVERDIR="$PWD/capstones/comparative/.coverage/data" \
  go test ./capstones/comparative/solution/kvstore
coverage_inputs=$(
  find capstones/comparative/.coverage/data \
    -mindepth 1 -maxdepth 1 -type d -printf '%p,' |
    sed 's/,$//'
)
go tool covdata textfmt \
  -i="$coverage_inputs" \
  -o=comparative-coverage.out
bash scripts/check-coverage.sh comparative-coverage.out 75
rm -rf capstones/comparative/.coverage
```

The comparative recipe uses GNU `find` as provided by the Ubuntu CI runner.

## Fuzz, staticcheck, and vulnerability gates

CI performs bounded fuzz smoke runs so fuzz targets are continuously executable:

```bash
go test -run '^$' \
  -fuzz '^FuzzDecodeCreate$' \
  -fuzztime=2s \
  ./projects/tasks/solution/server/api
go test -run '^$' \
  -fuzz '^FuzzDecodeUpdate$' \
  -fuzztime=2s \
  ./projects/tasks/solution/server/api
go test -run '^$' \
  -fuzz '^FuzzParseDocument$' \
  -fuzztime=2s \
  ./projects/tasks/solution/server/storage/markdown
go test -run '^$' \
  -fuzz '^FuzzParseKeyValue$' \
  -fuzztime=2s \
  ./lessons/08_testing/08_fuzzing
go test -run '^$' \
  -fuzz '^FuzzLoadConfig$' \
  -fuzztime=2s \
  ./capstones/idiomatic/solution/monitor/domain
```

Install the same pinned analyzers as CI, then analyze completed teaching and
reference packages rather than intentional starter implementations:

```bash
go install honnef.co/go/tools/cmd/staticcheck@v0.7.0
{
  go list ./lessons/...
  go list ./exercises/... | grep '/solution$'
  go list ./projects/tasks/solution/...
  go list ./projects/tasks/tests/...
  go list ./projects/tasks
  go list ./capstones/comparative/solution/...
  go list ./capstones/comparative/tests/...
  go list ./capstones/idiomatic/solution/...
  go list ./capstones/idiomatic/tests/...
  go list ./capstones/testsupport
  go list ./tools/...
} | xargs "$(go env GOPATH)/bin/staticcheck"

go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
"$(go env GOPATH)/bin/govulncheck" ./...
```

`go vet ./...` and `govulncheck ./...` inspect the whole buildable module.
Staticcheck includes the completed Tasks solution and reusable contracts, and
deliberately omits learner placeholders so they do not become the repository's
style baseline.
`govulncheck` also scans the Go standard library, so run it with a current patch
toolchain; an unpatched base release can correctly fail even when dependencies
are clean.

## Curriculum and contract evidence matrix

This maintained matrix is the target-repository conformance record. `covered`
means the course provides explanation, observable code, active practice, and a
later application where the topic is part of the promised outcome. `not
applicable` requires a target-specific reason. A future `partial` or `missing`
row blocks a course-complete claim until the gap is resolved or the promise is
explicitly narrowed.

| Area | Status | Inspectable target evidence | Learner impact | Action or scope rationale |
| --- | --- | --- | --- | --- |
| Course profile, prerequisites, versions, environment, scope | covered | [`README.md`](../README.md), [`SETUP.md`](SETUP.md), `go.mod`, CI matrix | Learners know the entry level, outcomes, supported Go lines, verified Ubuntu workflow, and non-goals. | Full macOS/Windows workflow parity is not claimed because coverage uses Bash and GNU `find`. |
| Running programs, syntax, and control flow | covered | Modules [1](../lessons/01_basics/README.md)–[2](../lessons/02_control_flow/README.md), matching exercises | Establishes the executable mental model before later abstractions. | Includes explicit Boolean conditions; Go has no truthiness. |
| Scalars, zero/nil values, conversions, and numeric boundaries | covered | Module [1](../lessons/01_basics/README.md), exercises 1 and 5, testing float caveat | Prevents integer-conversion and floating-point comparison mistakes. | Covers truncation, overflow, binary approximation, tolerance, NaN/infinity, and exact-decimal boundaries. |
| Text, Unicode, bytes, and encoding | covered | Module [1](../lessons/01_basics/README.md), exercise 1, project title validation | Learners distinguish bytes, runes, code points, and visible characters. | Grapheme segmentation remains an explicit advanced boundary. |
| Collections, iteration, mutation, copying, equality, and sets | covered | Module [4](../lessons/04_collections/README.md), exercise 4 | Prevents aliasing, nil-map, ordering, and accidental-mutation bugs. | Sets use `map[T]struct{}`; deterministic ordering is explicit. |
| Functions, parameters, closures, pointers, and returns | covered | Module [3](../lessons/03_functions_and_pointers/README.md), exercise 3 | Builds Go's value-semantics model before methods and concurrency. | Shared backing data is distinguished from pass-by-reference language. |
| Structs, methods, interfaces, and native data modeling | covered | Module [5](../lessons/05_structs_methods_interfaces/README.md), exercise 5 | Supports consumer-owned interfaces and composition used by projects. | Inheritance hierarchies are deliberately not taught as a Go default. |
| Errors, ownership, cleanup, and lifecycle | covered | Module [6](../lessons/06_errors_files_json/README.md), exercises 6, 10, and 12 | Makes failure and cleanup paths reviewable before applied work. | Expected failures use errors; deliberate panic examples are labeled. |
| Files, paths, streams, JSON, and validation | covered | Module [6](../lessons/06_errors_files_json/README.md), exercise 6, Tasks storage/API contracts | Teaches untrusted-boundary validation and deterministic cleanup. | Advanced serialization formats are outside the beginner scope. |
| Dates, times, durations, zones, timestamps, and clocks | covered | Module 6 [time lesson](../lessons/06_errors_files_json/05_time_durations_and_clocks/), exercise 6 time helpers, idiomatic capstone | Prepares learners for timeouts, SQL timestamps, reports, and deterministic clock seams. | Focuses on RFC 3339, UTC, elapsed duration, locations, `Equal`, and monotonic-clock boundaries. |
| Packages, modules, dependencies, build, and public docs | covered | Module [7](../lessons/07_packages_and_generics/README.md), `go.mod`, `go.sum`, setup/tooling docs | Connects package APIs to reproducible dependency management. | `go.sum` is correctly described as checksums, not a complete lockfile. |
| Idiomatic repository structure and artifact roles | covered | Root navigation, [`lessons/`](../lessons/README.md), [`exercises/`](../exercises/README.md), [`projects/`](../projects/README.md), [`capstones/`](../capstones/README.md) | Learners can locate instruction, starter, solution, tests, and applied destinations. | Roles use Go package/module conventions rather than a cross-language fixed tree. |
| CLI, environment, process interoperability, and exits | covered | Module [9](../lessons/09_tooling_cli_observability/README.md), Tasks CLI, both capstones | Teaches testable process boundaries and stable exit behavior. | Shell portability limits are stated in the support profile. |
| Typing, generics, and static analysis | covered | Modules [5](../lessons/05_structs_methods_interfaces/README.md), [7](../lessons/07_packages_and_generics/README.md), and [9](../lessons/09_tooling_cli_observability/README.md) | Learners use Go-native abstractions rather than translated class patterns. | Generics are introduced after concrete code and premature abstraction is discouraged. |
| Testing, debugging, formatting, linting, coverage, and CI | covered | Modules [8](../lessons/08_testing/README.md)–[9](../lessons/09_tooling_cli_observability/README.md), this guide, workflow | Provides an observable local-to-CI developer loop and failure path. | Starter behavior tests are intentionally separate from clean-checkout health. |
| Concurrency, cancellation, and lifecycle | covered | Module [10](../lessons/10_concurrency/README.md), exercise 10, idiomatic capstone | Establishes ownership, synchronization, race safety, cancellation, and joining before independent application. | Timing sleeps are not presented as synchronization. |
| Persistence, network/API boundaries, and untrusted input | covered | Modules [11](../lessons/11_sql_and_sqlite/README.md)–[12](../lessons/12_rest_apis_and_clients/README.md), exercises 11–12, Tasks project | Connects parameterized SQL, strict JSON, HTTP status/error contracts, and resource cleanup. | Public deployment, auth, TLS termination, and distributed systems remain explicit non-goals. |
| Algorithmic and performance fundamentals | covered | Recursion/search/sorting lessons and exercises, Module [8](../lessons/08_testing/README.md) benchmarks, Module [9](../lessons/09_tooling_cli_observability/README.md) profiling | Gives beginner-appropriate trade-off and measurement skills. | Formal algorithm-analysis coursework and advanced performance engineering are out of scope. |
| Theory-to-example-to-practice-to-project continuity | covered | Module links, matching exercise folders, required [Tasks project](../projects/tasks/README.md), both [capstones](../capstones/README.md) | Every required applied milestone points back to taught prerequisites and a feedback path. | Projects integrate existing outcomes rather than acting as prerequisite discovery. |
| Starter and solution integrity | covered | Exercise split, Tasks boundary/architecture tests, capstone harnesses and API parity tests | Starters remain usable without leaking solutions; references are complete destinations. | Intentional placeholders are explicit and compile at defined milestones. |
| Idiomaticity, currency, dependencies, security, portability | covered | Supported Go matrix, pinned dependencies/analyzers, strict boundary examples, stated Ubuntu verification boundary | Prevents preview/legacy or insecure shortcuts from becoming defaults. | Go 1.25/1.26 are supported; portability claims do not exceed automated evidence. |
| Projects and capstones | covered | Tasks [`SPEC.md`](../projects/tasks/docs/SPEC.md), comparative [`SPEC.md`](../capstones/comparative/spec/SPEC.md), idiomatic [`SPEC.md`](../capstones/idiomatic/SPEC.md) | Milestones expose architecture, normal behavior, boundaries, failures, and incremental feedback. | Production-appropriate structure is not presented as production-complete operations. |
| Automated validation and local/CI parity | covered | This guide, [workflow](../.github/workflows/course.yml), coverage script, link checker | Completion claims remain reproducible and inspectable. | GNU `find` is an explicit Ubuntu-only workflow detail. |
| Refinement and delivery evidence | covered | This matrix, independent refinement history in Git, CI results, focused commits and remote branch history | Future reviews can challenge claims against durable target evidence. | Each refinement still requires change-class checks and exactly one final risk-selected diff review. |

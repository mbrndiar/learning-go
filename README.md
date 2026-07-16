# 🐹 learning-go

A complete, hands-on introduction to Go for independent learners. The course
combines written explanations, small runnable programs, exercises with tests
and reference solutions, review questions, two staged capstone projects, and a
syntax reference. No previous programming experience is assumed.

## 🎯 What you will learn

By the end of the course, you will be able to:

- read, write, build, test, and debug Go programs;
- work confidently with values, control flow, functions, pointers, slices,
  maps, structs, methods, interfaces, and generics;
- model failures with explicit error values and manage resources with `defer`;
- read and write files and exchange structured data with JSON;
- organize code into packages and manage dependencies with Go modules;
- write table-driven tests, benchmarks, fuzz targets, and HTTP tests;
- build command-line programs and structured logs;
- design race-safe concurrent programs with goroutines, channels, and contexts;
- build HTTP/JSON services and persist data through `database/sql`; and
- design, test, and extend service and command-line applications.

## ✅ Requirements

- The latest patch release of Go 1.25 or newer
- Git for cloning the repository
- Most lessons use the standard library. Module 11 and the comparative capstone
  use the pinned pure-Go SQLite driver; Module 12 adds pinned Chi, Gin, and
  Resty examples.

See [`docs/SETUP.md`](docs/SETUP.md) for installation, editor setup, toolchain
selection, and troubleshooting.

## ▶️ How to run a lesson

Each runnable lesson is its own Go package. From the repository root:

```bash
go run ./lessons/01_basics/01_hello_world
```

For an unchanged learner checkout, a raw `go test ./...` is intentionally not
the repository health check: it runs the unfinished exercise starter tests and
fails until those exercises are implemented. CI instead compiles starter
packages without running their behavior tests, then tests lessons, reference
solutions, and capstone harnesses/solutions. See
[`docs/QUALITY.md`](docs/QUALITY.md) for the exact test, race, coverage, fuzz,
static-analysis, vulnerability, and link-check commands.

For each module:

1. Read its `README.md`, including the common mistakes.
2. Predict a lesson program's output before running it.
3. Run the program, then change one value or operation at a time.
4. Answer the review questions without looking back.
5. Complete the matching exercises before reading the solution.
6. Rebuild a small example from memory.

## 📐 Conventions used in this course

- Terminal commands are marked `bash`; Go code is marked `go`.
- `...` in a sample means omitted code unless the text says otherwise.
- Exported names begin with an uppercase letter; local names normally use
  `camelCase`.
- `gofmt` defines formatting. Do not align code manually.
- Errors shown intentionally are part of a lesson, not broken examples.
- Prefer explicit errors over `panic` for expected failures.

## 🧠 Practice exercises

Every module has a matching folder under
[`exercises/`](exercises/README.md). Starter packages contain TODOs and tests;
reference implementations live in separate `solution` packages.

## 🧩 Required applied project

After Module 12, complete the required
[`Task REST API and clients`](projects/tasks/README.md) project before beginning
the final capstones. It applies the course's domain, persistence, HTTP, client,
CLI, and testing material to one portable contract while comparing three Go
server stacks and two Go client stacks.

## 🏆 Capstone projects

The course now provides two staged capstones under
[`capstones/`](capstones/README.md):

- a shared versioned SQLite key/value store for comparing implementations
  across learning repositories; and
- an idiomatic Go service health monitor built around interfaces, contexts,
  goroutines, and `net/http`.

Both have matching starter/solution package boundaries and reusable contract
test support. Each includes a complete reference solution and guided
five-milestone starter; unfinished starter behavior remains explicit.

### Historical Task-project migration

The superseded connected Task Manager, client, and API were removed after the
current capstones became canonical. For comparison or migration study, their
last pre-removal snapshot is commit
[`b3211f9`](https://github.com/mbrndiar/learning-go/tree/b3211f99fc2ce5da54b88c59da3f12aacbed30ff/project)
at path `project/`. Treat that immutable snapshot as historical guidance. New
Task REST API work belongs under [`projects/tasks/`](projects/tasks/README.md);
new capstone work belongs under `capstones/`.

## 🗒️ Cheat sheet

[`CHEATSHEET.md`](CHEATSHEET.md) is a compact syntax, standard-library, testing,
concurrency, HTTP, and tooling reference.

## 🗺️ Course outline

1. **[Basics](lessons/01_basics/)**
   - [`01_hello_world/main.go`](lessons/01_basics/01_hello_world/main.go) — packages, imports, `main`, and `fmt`
   - [`02_variables_and_types/main.go`](lessons/01_basics/02_variables_and_types/main.go) — variables, constants, zero values, and conversions
   - [`03_operators/main.go`](lessons/01_basics/03_operators/main.go) — arithmetic, comparison, logical, and bitwise operators
   - [`04_strings_bytes_runes/main.go`](lessons/01_basics/04_strings_bytes_runes/main.go) — UTF-8 strings, bytes, runes, and Unicode
2. **[Control Flow](lessons/02_control_flow/)**
   - [`01_if_and_init/main.go`](lessons/02_control_flow/01_if_and_init/main.go) — `if` with scoped initialization
   - [`02_switch/main.go`](lessons/02_control_flow/02_switch/main.go) — expression and type switches
   - [`03_for_and_range/main.go`](lessons/02_control_flow/03_for_and_range/main.go) — Go's `for` forms and `range`
   - [`04_break_continue_labels/main.go`](lessons/02_control_flow/04_break_continue_labels/main.go) — loop control and labels
3. **[Functions and Pointers](lessons/03_functions_and_pointers/)**
   - [`01_functions_basics/main.go`](lessons/03_functions_and_pointers/01_functions_basics/main.go) — parameters and multiple returns
   - [`02_variadic_closures/main.go`](lessons/03_functions_and_pointers/02_variadic_closures/main.go) — variadic functions and closures
   - [`03_recursion/main.go`](lessons/03_functions_and_pointers/03_recursion/main.go) — recursive problem solving
   - [`04_pointers/main.go`](lessons/03_functions_and_pointers/04_pointers/main.go) — value semantics and pointers
4. **[Collections](lessons/04_collections/)**
   - [`01_arrays_and_slices/main.go`](lessons/04_collections/01_arrays_and_slices/main.go) — arrays, slices, length, and capacity
   - [`02_slice_sharing_and_copy/main.go`](lessons/04_collections/02_slice_sharing_and_copy/main.go) — backing arrays, cloning, and copying
   - [`03_maps/main.go`](lessons/04_collections/03_maps/main.go) — maps, membership, deletion, and ordering
   - [`04_sorting/main.go`](lessons/04_collections/04_sorting/main.go) — sorting and collection helpers
5. **[Structs, Methods and Interfaces](lessons/05_structs_methods_interfaces/)**
   - [`01_structs_and_literals/main.go`](lessons/05_structs_methods_interfaces/01_structs_and_literals/main.go) — structs and zero values
   - [`02_methods_and_receivers/main.go`](lessons/05_structs_methods_interfaces/02_methods_and_receivers/main.go) — value and pointer receivers
   - [`03_composition_and_interfaces/main.go`](lessons/05_structs_methods_interfaces/03_composition_and_interfaces/main.go) — embedding, composition, and interfaces
   - [`04_iota_and_nil_interfaces/main.go`](lessons/05_structs_methods_interfaces/04_iota_and_nil_interfaces/main.go) — enum-like constants and nil interfaces
6. **[Errors, Files and JSON](lessons/06_errors_files_json/)**
   - [`01_error_values_and_sentinels/main.go`](lessons/06_errors_files_json/01_error_values_and_sentinels/main.go) — explicit and sentinel errors
   - [`02_wrapping_and_inspecting_errors/main.go`](lessons/06_errors_files_json/02_wrapping_and_inspecting_errors/main.go) — `%w`, `errors.Is`, and `errors.As`
   - [`03_files_and_defer/main.go`](lessons/06_errors_files_json/03_files_and_defer/main.go) — files, buffered I/O, paths, and `defer`
   - [`04_json_encoding_and_validation/main.go`](lessons/06_errors_files_json/04_json_encoding_and_validation/main.go) — JSON tags and boundary validation
7. **[Packages, Modules and Generics](lessons/07_packages_and_generics/)**
   - [`01_package_organization/main.go`](lessons/07_packages_and_generics/01_package_organization/main.go) — packages, exports, and `internal`
   - [`02_generic_helpers/main.go`](lessons/07_packages_and_generics/02_generic_helpers/main.go) — type parameters and constraints
   - [`03_iterators_range_over_func/main.go`](lessons/07_packages_and_generics/03_iterators_range_over_func/main.go) — iterator functions and range-over-function
   - [`04_avoiding_premature_abstraction/main.go`](lessons/07_packages_and_generics/04_avoiding_premature_abstraction/main.go) — choosing concrete code or generics
8. **[Testing](lessons/08_testing/)**
   - [`01_basic_tests/stringutils_test.go`](lessons/08_testing/01_basic_tests/stringutils_test.go) — test discovery and assertions
   - [`02_table_driven/calc_test.go`](lessons/08_testing/02_table_driven/calc_test.go) — table-driven tests and subtests
   - [`03_subtests_helpers/config_test.go`](lessons/08_testing/03_subtests_helpers/config_test.go) — helpers, cleanup, and temporary directories
   - [`04_httptest/server_test.go`](lessons/08_testing/04_httptest/server_test.go) — handler and client testing
   - [`05_examples/format_test.go`](lessons/08_testing/05_examples/format_test.go) — executable examples
   - [`06_benchmarks/fib_test.go`](lessons/08_testing/06_benchmarks/fib_test.go) — benchmarks
   - [`07_coverage/grade_test.go`](lessons/08_testing/07_coverage/grade_test.go) — coverage interpretation
   - [`08_fuzzing/parse_test.go`](lessons/08_testing/08_fuzzing/parse_test.go) — fuzzing and regression seeds
9. **[Tooling, CLI and Observability](lessons/09_tooling_cli_observability/)**
   - [`01_cli_flags/main.go`](lessons/09_tooling_cli_observability/01_cli_flags/main.go) — testable standard-library flags
   - [`02_structured_logging/main.go`](lessons/09_tooling_cli_observability/02_structured_logging/main.go) — structured logging with `slog`
   - [`03_race_detector/main.go`](lessons/09_tooling_cli_observability/03_race_detector/main.go) — finding data races
   - [`04_pprof_delve/main.go`](lessons/09_tooling_cli_observability/04_pprof_delve/main.go) — profiling and debugger orientation
10. **[Concurrency](lessons/10_concurrency/)**
    - [`01_goroutines_basics/main.go`](lessons/10_concurrency/01_goroutines_basics/main.go) — goroutines and ownership
    - [`02_unbuffered_channels/main.go`](lessons/10_concurrency/02_unbuffered_channels/main.go) — synchronous channel communication
    - [`03_buffered_channels/main.go`](lessons/10_concurrency/03_buffered_channels/main.go) — bounded buffering
    - [`04_channel_direction_and_closing/main.go`](lessons/10_concurrency/04_channel_direction_and_closing/main.go) — channel direction and closing
    - [`05_select_and_timeouts/main.go`](lessons/10_concurrency/05_select_and_timeouts/main.go) — `select` and timeouts
    - [`06_waitgroup_and_mutex/main.go`](lessons/10_concurrency/06_waitgroup_and_mutex/main.go) — coordination and locks
    - [`07_atomic_counters/main.go`](lessons/10_concurrency/07_atomic_counters/main.go) — atomic operations
    - [`08_worker_pool/main.go`](lessons/10_concurrency/08_worker_pool/main.go) — bounded worker pools
    - [`09_context_cancellation/main.go`](lessons/10_concurrency/09_context_cancellation/main.go) — cancellation with contexts
    - [`10_goroutine_leaks_and_races/main.go`](lessons/10_concurrency/10_goroutine_leaks_and_races/main.go) — leak and race prevention
11. **[Relational Databases and SQL with SQLite](lessons/11_sql_and_sqlite/)**
    - [`01_relational_model_and_database_sql/main.go`](lessons/11_sql_and_sqlite/01_relational_model_and_database_sql/main.go) — relations, keys, constraints, and `database/sql`
    - [`02_parameterized_crud_and_row_mapping/main.go`](lessons/11_sql_and_sqlite/02_parameterized_crud_and_row_mapping/main.go) — parameterized CRUD and row mapping
    - [`03_joins_aggregates_and_indexes/main.go`](lessons/11_sql_and_sqlite/03_joins_aggregates_and_indexes/main.go) — joins, aggregates, and indexes
    - [`04_transactions_and_sqlite/main.go`](lessons/11_sql_and_sqlite/04_transactions_and_sqlite/main.go) — atomic transactions and SQLite specifics
    - [`05_repository_pattern/main.go`](lessons/11_sql_and_sqlite/05_repository_pattern/main.go) — narrow persistence interfaces
12. **[REST APIs and HTTP Clients](lessons/12_rest_apis_and_clients/)**
    - [`01_http_routing_and_json/main.go`](lessons/12_rest_apis_and_clients/01_http_routing_and_json/main.go) — method-aware routes and strict JSON
    - [`02_middleware_and_error_mapping/main.go`](lessons/12_rest_apis_and_clients/02_middleware_and_error_mapping/main.go) — middleware and error responses
    - [`03_http_client_context_timeout/main.go`](lessons/12_rest_apis_and_clients/03_http_client_context_timeout/main.go) — clients, contexts, and timeouts
    - [`04_router_framework_comparison/main.go`](lessons/12_rest_apis_and_clients/04_router_framework_comparison/main.go) — runnable net/http, Chi, and Gin comparison
    - [`05_resty_client/main.go`](lessons/12_rest_apis_and_clients/05_resty_client/main.go) — configured Resty requests and typed responses
    - [`06_graceful_shutdown/main.go`](lessons/12_rest_apis_and_clients/06_graceful_shutdown/main.go) — server shutdown and cleanup

Work through the modules in order. Module 10 establishes concurrency and
cancellation, Module 11 adds relational persistence, and Module 12 applies those
boundaries to HTTP services and clients. Then complete the required
[`Tasks project`](projects/tasks/README.md) before the capstones.

## 🆘 Getting help from the material

Read compiler errors from top to bottom: Go usually reports the source location
and the violated rule precisely. For runtime failures, reduce the program to
the smallest reproducible case, inspect returned errors, and add a regression
test after fixing the issue.

Solutions demonstrate one clear approach, not the only correct one. Compare
behavior, error handling, ownership, readability, race safety, and tests.

## 🧭 Course boundaries

This course aims to make a beginner independently productive with core Go and
its standard library. Specialized distributed systems, cloud-native
orchestration, advanced performance engineering, and framework-specific web
development require further study after these foundations.
# 🐹 learning-go

A complete, hands-on introduction to Go for independent learners. The course
combines written explanations, small runnable programs, exercises with tests
and reference solutions, review questions, a connected capstone project, and a
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
- design, test, and extend a connected command-line application.

## ✅ Requirements

- The latest patch release of Go 1.25 or newer
- Git for cloning the repository
- The lessons themselves use the standard library. The capstone additionally
  uses a pinned pure-Go SQLite driver.

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
solutions, capstone harnesses/solutions, and the retained Task projects. See
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

The connected [Task projects](project/README.md) remain available as completed
legacy integration examples:

```text
Task Manager CLI -> Manager -> Storage
                             |-> JSON file
                             `-> REST client -> REST API -> SQLite
```

Use their [old-to-new concept map](project/README.md#-old-to-new-concept-map)
when moving from the Task code to the current capstones. The old code remains
available for comparison, but new capstone work belongs under `capstones/`.

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
11. **[Application Integration](lessons/11_application_integration/)**
    - [`01_http_routing_pathvalue/main.go`](lessons/11_application_integration/01_http_routing_pathvalue/main.go) — method-aware routing and path values
    - [`02_json_request_response/main.go`](lessons/11_application_integration/02_json_request_response/main.go) — strict JSON boundaries
    - [`03_middleware_functions/main.go`](lessons/11_application_integration/03_middleware_functions/main.go) — composable middleware
    - [`04_http_client_context_timeout/main.go`](lessons/11_application_integration/04_http_client_context_timeout/main.go) — clients, contexts, and timeouts
    - [`05_database_sql_concepts/main.go`](lessons/11_application_integration/05_database_sql_concepts/main.go) — `database/sql` abstractions
    - [`06_parameterized_sql_transactions/main.go`](lessons/11_application_integration/06_parameterized_sql_transactions/main.go) — parameters and transactions
    - [`07_graceful_shutdown/main.go`](lessons/11_application_integration/07_graceful_shutdown/main.go) — server shutdown and cleanup

Work through the modules in order. Module 11 bridges directly into the
capstone; module 10 provides the context cancellation and goroutine ownership
used by production-style HTTP programs.

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
# 🧠 Exercises

Each lesson module has a matching practice folder. Starter packages contain
TODOs and tests; `solution/` contains one reference implementation. Each
exercise README links back to the module explanation so the task contract does
not need to repeat the theory.

## ▶️ Workflow

```bash
go test ./exercises/01_basics
```

1. Read the task contract and tests.
2. Implement one behavior at a time.
3. Run the smallest relevant test.
4. Add a boundary or failure case.
5. Compare with `solution/` only after making a genuine attempt.

Starter tests are expected to fail until implementation. CI compiles starter
packages without running their tests and runs every solution package.

## 🗂️ Modules

1. [`01_basics/`](01_basics/) — conversions, operators, and Unicode-safe string helpers.
2. [`02_control_flow/`](02_control_flow/) — classification, loops, linear search, and binary search.
3. [`03_functions_and_pointers/`](03_functions_and_pointers/) — multiple returns, variadics, closures, and pointer mutation.
4. [`04_collections/`](04_collections/) — slices, maps, copying, sorting, and non-mutating transformations.
5. [`05_structs_methods_interfaces/`](05_structs_methods_interfaces/) — shape types, methods, interfaces, `fmt.Stringer`, and `iota`.
6. [`06_errors_files_json/`](06_errors_files_json/) — wrapped errors, files/JSON, timestamps, durations, and directory operations.
7. [`07_packages_and_generics/`](07_packages_and_generics/) — constraints, generic helpers, stacks, queues, and package API boundaries.
8. [`08_testing/`](08_testing/) — table tests, benchmarks, fuzz targets, and static test-completeness checks.
9. [`09_tooling_cli_observability/`](09_tooling_cli_observability/) — argument parsing, thin commands, structured logging, and diagnostics.
10. [`10_concurrency/`](10_concurrency/) — channels, fan-in, bounded work, cancellation, and first-error propagation.
11. [`11_sql_and_sqlite/`](11_sql_and_sqlite/) — schemas, parameterized CRUD, joins, transactions, and repositories.
12. [`12_rest_apis_and_clients/`](12_rest_apis_and_clients/) — strict HTTP/JSON handlers, error mapping, and context-aware clients.

After all twelve exercise solutions, continue with the required
[`Task REST API and clients`](../projects/tasks/README.md) project before the
capstones.

## 🧩 Problem-solving process

Define inputs, outputs, ownership, zero-value behavior, and errors before
coding. Prefer small functions and explicit errors. Use table-driven tests when
several examples share the same behavior.

## 🔍 Reference solutions

A solution is a comparison point, not a required implementation. Evaluate
clarity, allocations, mutation, error wrapping, race safety, and testability.

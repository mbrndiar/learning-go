# 🎓 Lessons

Twelve modules teach Go through small packages that compile and run
independently. Each module README explains the prerequisite concepts, mental
models, important boundaries, experiments, and review questions; the lesson
packages provide the concrete code to predict, run, and modify.

## 🗂️ Modules

1. [`01_basics/`](01_basics/) — values, types, operators, Unicode, bytes, and runes.
2. [`02_control_flow/`](02_control_flow/) — Boolean conditions, switches, loops, range, and labels.
3. [`03_functions_and_pointers/`](03_functions_and_pointers/) — functions, closures, recursion, value semantics, and pointers.
4. [`04_collections/`](04_collections/) — arrays, slices, maps, copying, mutation, and deterministic sorting.
5. [`05_structs_methods_interfaces/`](05_structs_methods_interfaces/) — structs, receivers, composition, interfaces, `iota`, and typed nils.
6. [`06_errors_files_json/`](06_errors_files_json/) — errors, resource cleanup, files, directories, JSON, durations, and timestamps.
7. [`07_packages_and_generics/`](07_packages_and_generics/) — packages, modules, generics, iterators, and abstraction trade-offs.
8. [`08_testing/`](08_testing/) — unit/table tests, helpers, `httptest`, examples, benchmarks, coverage, and fuzzing.
9. [`09_tooling_cli_observability/`](09_tooling_cli_observability/) — Go tooling, testable CLIs, structured logging, race detection, profiling, and debugging.
10. [`10_concurrency/`](10_concurrency/) — goroutines, channels, synchronization, worker pools, contexts, and lifecycle.
11. [`11_sql_and_sqlite/`](11_sql_and_sqlite/) — relational modeling,
    parameterized SQL, transactions, and repositories with SQLite.
12. [`12_rest_apis_and_clients/`](12_rest_apis_and_clients/) — strict HTTP
    boundaries, middleware, clients, framework comparison, and shutdown.

## ▶️ How to use a module

1. Read the module README.
2. Open lesson packages in order.
3. Predict output before running `go run`.
4. Change one detail and explain the result.
5. Complete the matching exercise package.
6. Answer the review questions from memory.

## 🔁 Recommended study loop

Preview, predict, run, modify, test, explain, and rebuild from memory. Compiler
success is not the final goal: you should be able to explain ownership,
zero-value behavior, error paths, and why a design is race safe.

## 🚩 Checkpoints

- After modules 1-2: build a menu-driven number game.
- After modules 3-4: build a text analyzer using functions, slices, and maps.
- After modules 5-6: model and persist a collection of structs.
- After modules 7-9: package, test, benchmark, and instrument a CLI.
- After module 10: implement a bounded worker pool with cancellation.
- After module 11: design a relational schema and implement a tested
  `database/sql` repository with one atomic multi-step operation.
- After module 12: complete the required
  [`Task REST API and clients`](../projects/tasks/README.md) project, progressing
  from the shared domain and repositories to three servers, two clients,
  OpenAPI validation, and full interoperability.

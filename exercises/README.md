# 🧠 Exercises

Each lesson module has a matching practice folder. Starter packages contain
TODOs and tests; `solution/` contains one reference implementation.

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

1. [`01_basics/`](01_basics/)
2. [`02_control_flow/`](02_control_flow/)
3. [`03_functions_and_pointers/`](03_functions_and_pointers/)
4. [`04_collections/`](04_collections/)
5. [`05_structs_methods_interfaces/`](05_structs_methods_interfaces/)
6. [`06_errors_files_json/`](06_errors_files_json/)
7. [`07_packages_and_generics/`](07_packages_and_generics/)
8. [`08_testing/`](08_testing/)
9. [`09_tooling_cli_observability/`](09_tooling_cli_observability/)
10. [`10_concurrency/`](10_concurrency/)
11. [`11_sql_and_sqlite/`](11_sql_and_sqlite/)

## 🧩 Problem-solving process

Define inputs, outputs, ownership, zero-value behavior, and errors before
coding. Prefer small functions and explicit errors. Use table-driven tests when
several examples share the same behavior.

## 🔍 Reference solutions

A solution is a comparison point, not a required implementation. Evaluate
clarity, allocations, mutation, error wrapping, race safety, and testability.

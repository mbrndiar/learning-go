# 🧪 Module 08 — Testing

Go treats testing as a first-class part of the toolchain, not an add-on
library. This module builds one small, focused package per topic so each
idea is easy to isolate, run, and modify.

## 🎯 Learning goals

By the end of this module you will be able to:

- write and run tests with `go test` and `*testing.T`;
- structure many related cases as **table-driven tests** with `t.Run`
  subtests;
- write test **helper functions** with `t.Helper()`, manage temporary files
  with `t.TempDir()`, and register teardown with `t.Cleanup`;
- test HTTP handlers and full request/response round trips with
  `net/http/httptest`;
- write **Example** functions that double as verified documentation;
- write and compare **benchmarks** with `testing.B`;
- read `go test -cover` output to find untested branches; and
- write a **fuzz target** with seed corpus entries and regression seeds.

## 🗂️ Lesson map

1. [`01_basic_tests/`](01_basic_tests/) — `*testing.T` basics: `t.Error` vs
   `t.Fatal`, `t.Log`, `t.Skip`.
2. [`02_table_driven/`](02_table_driven/) — table-driven tests with
   `t.Run` subtests.
3. [`03_subtests_helpers/`](03_subtests_helpers/) — helper functions,
   `t.Helper()`, `t.TempDir()`, `t.Cleanup`.
4. [`04_httptest/`](04_httptest/) — testing HTTP handlers with
   `httptest.NewRecorder` and `httptest.NewServer`.
5. [`05_examples/`](05_examples/) — `Example` functions verified by
   `// Output:` comments.
6. [`06_benchmarks/`](06_benchmarks/) — benchmarks with `testing.B` and
   sub-benchmarks.
7. [`07_coverage/`](07_coverage/) — reading `go test -cover` /
   `go tool cover` output.
8. [`08_fuzzing/`](08_fuzzing/) — fuzz targets, seed corpus, and regression
   seeds under `testdata/fuzz/`.

## ▶️ How to run these lessons

Every lesson package here is a **non-main package** exercised through
`go test`, not `go run` — there is nothing to execute except the tests
themselves. From the repository root:

```bash
go test ./lessons/08_testing/01_basic_tests
go test -v ./lessons/08_testing/02_table_driven          # verbose: show every (sub)test name
go test -run TestAdd/negative ./lessons/08_testing/02_table_driven
go test ./lessons/08_testing/...                          # every package in this module
```

## 🔬 Topic notes

### `*testing.T` basics (`01_basic_tests`)

- A test function must be named `TestXxx`, take a single `*testing.T`
  parameter, and live in a file ending in `_test.go`.
- `t.Error`/`t.Errorf` record a failure and let the function continue.
- `t.Fatal`/`t.Fatalf` record a failure and stop the function immediately
  (via `runtime.Goexit`) — use them when continuing would be meaningless or
  would panic (for example, after a failed setup step).
- `t.Log`/`t.Logf` print diagnostics that are shown only on failure or with
  `-v`.
- `t.Skip`/`t.Skipf` mark a test as intentionally not run, with a reason.

### Table-driven tests (`02_table_driven`)

Define test cases as data (a slice or map of structs), then loop over them
running one `t.Run` subtest per case:

```go
tests := []struct {
    name string
    a, b int
    want int
}{
    {name: "positive", a: 2, b: 3, want: 5},
}

for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
        if got := Add(test.a, test.b); got != test.want {
            t.Errorf("Add(%d, %d) = %d, want %d", test.a, test.b, got, test.want)
        }
    })
}
```

Subtests get independent pass/fail reporting and can be selected
individually with `-run Parent/child`. Since Go 1.22, each loop iteration
gets its own `test` variable, so capturing it into a local variable before
the closure (`test := test`) is no longer necessary.

### Helpers, `t.TempDir`, and `t.Cleanup` (`03_subtests_helpers`)

- `t.Helper()` marks a function as a test helper so failure line numbers
  reported by `t.Fatalf`/`t.Errorf` point at the **caller**, not at the
  helper's own line.
- `t.TempDir()` returns a directory unique to the current test (or
  subtest) that is removed automatically when the test ends — no manual
  `os.RemoveAll`.
- `t.Cleanup(func)` registers a function to run when the test ends, even if
  the test fails or calls `t.Fatal` partway through. It composes: helpers
  can register their own cleanup and callers do not need to remember to
  call it.

### `net/http/httptest` (`04_httptest`)

Two complementary styles:

- `httptest.NewRecorder()` + `handler.ServeHTTP(recorder, request)` calls
  the handler directly with no real network listener — fast, ideal for
  exhaustive table-driven cases.
- `httptest.NewServer(handler)` starts a real listener on `127.0.0.1` and
  gives you a `*http.Client` through `server.Client()` — closer to a real
  round trip, always paired with `t.Cleanup(server.Close)` or
  `defer server.Close()`.

### `Example` functions (`05_examples`)

An `ExampleXxx` function is compiled and, if it contains a trailing
`// Output:` comment, executed by `go test`, which captures stdout and
compares it to the comment text. Naming conventions:

- `ExampleFoo` documents function/type `Foo`.
- `ExampleFoo_bar` documents a specific scenario for `Foo`.
- `Example` (no suffix) documents the package as a whole.
- `// Unordered output:` compares output as an unordered set of lines.

Examples double as documentation: `go doc` displays them next to the
symbol they document.

### Benchmarks (`06_benchmarks`)

A `BenchmarkXxx(b *testing.B)` function runs its body `b.N` times; the
testing framework increases `b.N` until the timing is stable. Benchmarks
only run with `-bench` (a plain `go test` skips them):

```bash
go test -bench=. -benchmem ./lessons/08_testing/06_benchmarks
go test -bench=FibIterative ./lessons/08_testing/06_benchmarks
```

Keep correctness tests for the same code (`TestFibImplementationsAgree`) so
a benchmark suite cannot silently drift from correct behavior.

### Coverage (`07_coverage`)

```bash
go test -coverprofile=coverage.out ./lessons/08_testing/07_coverage
go tool cover -func=coverage.out    # per-function percentages in the terminal
go tool cover -html=coverage.out    # opens an annotated HTML view
```

`07_coverage`'s tests intentionally skip one branch (`Classify`'s "D"
grade). Run the commands above and find the uncovered line yourself before
looking at the source again. Coverage measures **statements executed**, not
correctness — 100% coverage does not guarantee bug-free code, and less than
100% is not automatically wrong, but an unexplained gap deserves a look.

### Fuzzing (`08_fuzzing`)

```bash
go test ./lessons/08_testing/08_fuzzing                        # runs seeds + regressions only
go test -fuzz=FuzzParseKeyValue -fuzztime=30s ./lessons/08_testing/08_fuzzing
```

- A fuzz target is `FuzzXxx(f *testing.F)`; `f.Add(...)` registers seed
  inputs that always run under plain `go test`.
- `f.Fuzz(func(t *testing.T, input string) { ... })` defines the property
  to check. Under `-fuzz`, the runtime also generates random inputs derived
  from the seed corpus.
- If fuzzing finds a failing input, Go writes it under
  `testdata/fuzz/FuzzXxx/`. This module ships two such files
  (`regression_multiple_equals`, `regression_whitespace_only_key`) as
  **regression seeds** — inputs a fuzzer found interesting in the past.
  `go test` (with or without `-fuzz`) always replays every file in that
  directory, so a fixed bug can never silently regress.

## ⚠️ Common mistakes

- Forgetting the `_test.go` suffix or the `TestXxx`/`BenchmarkXxx`/`FuzzXxx`
  naming convention — the tool simply will not find the function.
- Using `t.Fatal` inside a **goroutine** spawned by a test: only the
  goroutine that calls it stops; the test can hang or panic. Send failures
  back to the main test goroutine instead (for example over a channel).
- Comparing floating-point results with `==` in a table-driven test instead
  of an epsilon-based comparison — fine for the exact fractions used here,
  risky for arbitrary computed floats.
- Assuming `-cover` proves correctness. It only proves a statement ran, not
  that its result was checked.
- Writing a fuzz target with no invariant to check (`f.Fuzz(func(t
  *testing.T, input string) { ParseKeyValue(input) })`): this only proves
  the code does not panic. Prefer checking a property (like the round-trip
  invariant in `08_fuzzing`) so real bugs surface, not only crashes.
- Deleting files under `testdata/fuzz/` "to clean up" — that discards
  regression coverage for previously discovered bugs.

## ❓ Review questions

1. What is the difference between `t.Error` and `t.Fatal`? When would using
   `t.Fatal` in the wrong place hide information from you?
2. Why does `t.Run` give each case its own name in `go test -v` output, and
   why is that useful when only one case fails out of twenty?
3. What does `t.Helper()` change about failure output, and why does calling
   it matter more as helper functions get reused across many tests?
4. Why does `t.TempDir()` remove the need for manual cleanup, and what
   happens to that directory if the test panics?
5. When would you reach for `httptest.NewRecorder` instead of
   `httptest.NewServer`, and vice versa?
6. What makes an `Example` function fail: a wrong `// Output:` comment, a
   compile error, or both? What happens if you omit the `Output:` comment
   entirely?
7. Why do benchmarks need a separate correctness test instead of trusting
   that a fast benchmark implies a correct one?
8. If `go test -cover` reports 100%, does that mean the package has no
   bugs? Why or why not?
9. What is the purpose of a regression seed file under `testdata/fuzz/`,
   and what would happen to that guarantee if it were deleted?
10. Why does a fuzz target benefit from checking an invariant (like a
    round trip) rather than only checking "did it panic"?

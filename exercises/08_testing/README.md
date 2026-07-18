# 🧪 Exercises: Testing

This exercise applies [Module 8](../../lessons/08_testing/README.md).
`textproc` is a small, already-implemented text-processing package. Unlike
earlier modules, you will not implement its functions — you will write the
**tests, benchmark, and fuzz target** that prove they work. This mirrors real
Go work: production code is often reviewed through the tests it ships with.

A `TestRequiredTestsExist` check (`meta_test.go`, do not edit) statically
inspects this package's test source. It exists because `go test` can look
green even when nothing was written: `Benchmark*` functions only run with
`-bench`, `Fuzz*` functions only run with `-fuzz`, and a leftover `t.Skip`
reports as skipped, not failed — none of that fails a plain `go test`. The
meta-test parses the source itself, so it fails loudly, with a precise
per-item report, until every required test is genuinely implemented.

## 🧩 Tasks

1. `TestNormalize` — table-driven test. Cover leading/trailing whitespace,
   multiple interior spaces, mixed case, and an empty string.
2. `TestWordFrequency` — table-driven test. Cover a normal sentence, repeated
   words, an empty string (expect an empty map), and mixed-case words that
   must be counted together.
3. `TestReverse` — table-driven test. Cover empty string, single rune, an
   ASCII word, and a multi-byte string (e.g. `"héllo"`) to prove runes, not
   bytes, are reversed.
4. `TestSafeSlice` — table-driven test. Cover at least one success case and
   at least two distinct failure cases (negative start, negative length,
   start beyond the rune length, end beyond the rune length). Assert an
   error is returned; `SafeSlice` must never panic.
5. `BenchmarkWordFrequency` — build a representative input once, call
   `b.ResetTimer`, then call `WordFrequency` inside a `for i := 0; i < b.N;
   i++` loop.
6. `FuzzSafeSlice` — seed the corpus with `f.Add` (include inputs you expect
   to fail), then use `f.Fuzz` to assert `SafeSlice` never panics and that a
   `nil` error always corresponds to a result whose rune count equals
   `length`.

## ▶️ Commands

```bash
go test ./exercises/08_testing/...
go test -run '^$' ./exercises/08_testing
go test ./exercises/08_testing/solution
go test -bench . -benchtime 10x ./exercises/08_testing/solution
go test -fuzz FuzzSafeSlice -fuzztime 5s ./exercises/08_testing/solution
gofmt -l exercises/08_testing
```

## 📝 Notes

- A table-driven test is a `[]struct{...}` of cases run with
  `t.Run(tt.name, func(t *testing.T) {...})` inside a loop — not a handful of
  standalone `if` checks.
- `go test` alone does not execute benchmarks or fuzzing; use `-bench` and
  `-fuzz` (or `go test -run '^$' -bench .` to skip regular tests) to actually
  run them, in addition to relying on the meta-test's static check.
- `f.Fuzz`'s function signature must start with `*testing.T` followed by the
  same argument types you passed to `f.Add`.
- `meta_test.go` is scaffolding, not a task: read it if you are curious how
  it works, but do not edit it.
- Compare with `solution/` only after a genuine attempt.

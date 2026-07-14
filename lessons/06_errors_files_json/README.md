# 🧯 Module 6 — Errors, Files and JSON

This module treats failure as data. You will model errors as ordinary values,
wrap them with context while keeping them inspectable, manage resources
deterministically with `defer`, and exchange structured data safely through
JSON.

## 🎯 Learning goals

By the end of this module you will be able to:

- treat errors as values instead of exceptions, using sentinel and typed
  errors where each fits best;
- wrap errors with `fmt.Errorf` and `%w` without losing the original cause;
- inspect wrapped error chains with `errors.Is` and `errors.As`;
- use `defer` to pair resource acquisition with guaranteed cleanup;
- read and write files with `os`, `io`, and `bufio`, and build paths safely
  with `path/filepath`; and
- encode and decode JSON with struct tags, and validate decoded data at a
  program's boundary.

## 📦 Lessons

1. [`01_error_values_and_sentinels/`](01_error_values_and_sentinels/) —
   errors as values, sentinel errors, typed errors.
2. [`02_wrapping_and_inspecting_errors/`](02_wrapping_and_inspecting_errors/)
   — `fmt.Errorf("...: %w", err)`, `errors.Is`, `errors.As`, `errors.Unwrap`,
   `errors.Join`.
3. [`03_files_and_defer/`](03_files_and_defer/) — `defer` order and use,
   `os`/`bufio` file I/O, `path/filepath`.
4. [`04_json_encoding_and_validation/`](04_json_encoding_and_validation/) —
   JSON struct tags, `json.Marshal`/`Unmarshal`, `json.Decoder`, boundary
   validation of decoded values.

## ▶️ How to run a lesson

From the repository root:

```bash
go run ./lessons/06_errors_files_json/01_error_values_and_sentinels
go run ./lessons/06_errors_files_json/02_wrapping_and_inspecting_errors
go run ./lessons/06_errors_files_json/03_files_and_defer
go run ./lessons/06_errors_files_json/04_json_encoding_and_validation
```

Predict the output of each program first, especially the `defer` ordering and
the strict-versus-lenient JSON decoding cases, then run it to check.

## 🚧 Common mistakes

- **Comparing wrapped errors with `==`.** Once an error has passed through
  `fmt.Errorf("...: %w", err)`, the outer value is never `==` to the
  original. Use `errors.Is` for sentinel comparisons and `errors.As` for
  typed errors, both of which walk the whole wrap chain.
- **Using `%v` instead of `%w` when wrapping.** `%v` renders the inner error
  as a plain string and discards the chain, so `errors.Is`/`errors.As` can no
  longer find it. Use `%w` whenever the caller might need to inspect the
  cause.
- **Forgetting to check `scanner.Err()`.** `bufio.Scanner`'s `Scan()` returns
  `false` both at normal end-of-input and on a real read error; always check
  `Err()` after the loop to tell them apart.
- **Not flushing a `bufio.Writer`.** Buffered writes can sit in memory until
  `Flush()` runs (or the writer is closed through a type that flushes on
  close). A missing `Flush()` silently loses the tail of the output.
- **Building paths with string concatenation.** `dir + "/" + name` breaks on
  Windows and mishandles trailing slashes; use `filepath.Join` instead.
- **Trusting decoded JSON without validation.** A successful `Unmarshal`
  only proves the JSON's shape matched the struct; it says nothing about
  whether the values are sensible. Validate required fields and ranges
  before using decoded data.
- **Not deciding whether unknown JSON fields are acceptable.** Plain
  `json.Unmarshal` silently ignores fields it does not recognize, which
  hides typos. Use `json.NewDecoder(...).DisallowUnknownFields()` when input
  should be rejected instead of silently truncated.

## ❓ Review questions

1. When should an error be a sentinel value versus a typed error with
   fields?
2. What does `%w` do differently from `%v` inside `fmt.Errorf`, and why does
   that matter for `errors.Is`/`errors.As`?
3. In what order do multiple `defer` statements in the same function run,
   and why does that make `defer` well suited to resource cleanup?
4. Why must a file opened for writing generally be closed (or its writer
   flushed) before the program relies on its contents?
5. What is the difference in behavior between `json.Unmarshal` and a
   `json.Decoder` configured with `DisallowUnknownFields`?
6. Why is validating decoded JSON a separate step from decoding itself?

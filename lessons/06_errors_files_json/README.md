# 🧯 Module 6 — Errors, Files, Directories, JSON and Time

This module treats failure as data. You will model errors as ordinary values,
wrap them with context while keeping them inspectable, manage resources
deterministically with `defer`, work with files and directory trees, exchange
structured data safely through JSON, and distinguish elapsed durations from
wall-clock timestamps.

## 🎯 Learning goals

By the end of this module you will be able to:

- treat errors as values instead of exceptions, using sentinel and typed
  errors where each fits best;
- wrap errors with `fmt.Errorf` and `%w` without losing the original cause;
- inspect wrapped error chains with `errors.Is` and `errors.As`;
- use `defer` to pair resource acquisition with guaranteed cleanup;
- read and write files with `os`, `io`, and `bufio`, build paths with
  `path/filepath`, and create, inspect, traverse, move, and remove directories
  with explicit ownership boundaries;
- encode and decode JSON with struct tags, and validate decoded data at a
  program's boundary;
- parse and format RFC 3339 timestamps, normalize instants to UTC, work with
  `time.Duration`, and inject a clock when tests need deterministic time.

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
5. [`05_time_durations_and_clocks/`](05_time_durations_and_clocks/) —
   durations, instants and locations, RFC 3339, UTC normalization, elapsed
   time, and a testable clock seam.
6. [`06_directory_operations/`](06_directory_operations/) — `os.MkdirAll`,
   `os.ReadDir`, `os.Stat`, `filepath.WalkDir`, deterministic relative paths,
   same-filesystem moves, and safe `Remove`/owned-`RemoveAll` boundaries.

## ▶️ How to run a lesson

From the repository root:

```bash
go run ./lessons/06_errors_files_json/01_error_values_and_sentinels
go run ./lessons/06_errors_files_json/02_wrapping_and_inspecting_errors
go run ./lessons/06_errors_files_json/03_files_and_defer
go run ./lessons/06_errors_files_json/04_json_encoding_and_validation
go run ./lessons/06_errors_files_json/05_time_durations_and_clocks
go run ./lessons/06_errors_files_json/06_directory_operations
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
- **Using `os.RemoveAll` on an unchecked path.** Recursive deletion is
  appropriate for a temporary tree the program just created and owns, not for
  a path assembled from untrusted or weakly validated input.
- **Treating `filepath.Clean` as a sandbox.** It normalizes path syntax
  lexically; it does not prove that a path remains below an intended root or
  account for symlink escapes.
- **Expecting `filepath.WalkDir` to follow symlinked directories.** It reports
  the symlink entry but does not traverse its target unless the program adds an
  explicit, loop-safe policy.
- **Assuming `os.Rename` is a universal move.** It can fail across filesystems,
  and replacement behavior for an existing destination varies by operating
  system.
- **Comparing `time.Time` values with `==`.** That operator also compares the
  location and monotonic-clock metadata. Use `t.Equal(u)` when the question is
  whether two values represent the same instant.
- **Persisting local display time without an offset.** Use a standard such as
  RFC 3339 and normally normalize machine-facing timestamps to UTC; convert to
  a user's location only at the presentation boundary.

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
7. What is the difference between a `time.Time` instant and the location used
   to display it, and why does `Time.Equal` usually express intent better than
   `==`?
8. Why should elapsed work use durations/monotonic clock readings while
   persisted timestamps use a wall-clock representation such as RFC 3339?
9. When should you use `os.ReadDir` instead of `filepath.WalkDir`, and what path
   information does each callback or entry provide?
10. Why is `os.Remove` a safer default than `os.RemoveAll`, and when is
    recursive cleanup justified?
11. Why does `filepath.Clean` not establish a security boundary, and what
    additional policy is needed for untrusted paths?
12. What limitations of `os.Rename` matter when a program moves files?

# 🧯 Module 6 — Errors, Files, Directories, JSON and Time

This module treats failure as data. You will model errors as ordinary values,
wrap them with context while keeping them inspectable, manage resources
deterministically with `defer`, work with files and directory trees, exchange
structured data safely through JSON, and distinguish elapsed durations from
wall-clock timestamps.

**Prerequisites:** Modules 1–5 (Basics, Control Flow, Functions and Pointers,
Collections, Structs/Methods/Interfaces). In particular, the error and clock
types here are ordinary structs with methods and small interfaces, so this
module leans directly on Module 5's mental models.

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

## 📖 Errors, resources, data and time, explained

An **error** in Go is just a value satisfying the one-method `error`
interface (`Error() string`) — not a separate exception mechanism with its
own control flow. That lets errors be compared, wrapped, stored, and passed
around like any other value. A **sentinel error** (`var ErrNotFound = ...`)
identifies a specific condition by identity; a **typed error**
(`*ValidationError`) additionally carries structured detail. `fmt.Errorf`
with `%w` builds an **error chain**: each wrap keeps a reference to the error
it wraps instead of flattening it into a string, so `errors.Is` can walk the
whole chain looking for a matching sentinel and `errors.As` can walk it
looking for a matching type, both succeeding no matter how many layers of
wrapping sit in between. Using `%v` instead of `%w`, or comparing an error to
a string, discards that chain — the check might work today and silently stop
working the moment someone adds a wrapping layer.

Go's resource-management idiom is **acquire, then immediately defer the
release**: call `os.Open`, check the error, and put `defer file.Close()` on
the very next line, before doing anything else with the file. `defer`
schedules a call to run when the *enclosing function* returns, not when a
block or loop iteration ends, so deferred calls in a loop all pile up until
the function exits — a reason to factor loop bodies that open resources into
their own function. A deferred cleanup can itself fail (a deferred `Close`
on a file opened for writing can return a real error, and `RemoveAll` can
fail to remove everything); when that failure matters, capture it with a
named return value and `errors.Join` rather than discarding it, as lesson 6's
`run` function does for the temporary root it owns.

Data crossing an I/O boundary passes through distinct layers, and it is worth
keeping them separate in your head: **raw bytes** (what `os.File` and
`io.Reader` hand you, with no assumed structure), **decoded text** (bytes
interpreted under an encoding, typically UTF-8, and often split into lines
with `bufio.Scanner`), and **structured data** (JSON decoded into a Go value
by `encoding/json`, obeying the shape a struct's fields and tags describe).
Each layer only guarantees what its own step promises. Successful JSON decoding
proves the input was valid JSON and that encountered values could be assigned
under the decoder's rules; by default, missing fields become zero values and
unknown fields may be ignored. It says nothing about whether required values
are present or sensible, so strict shape rules and semantic validation are
separate steps before decoded data is trusted.

Files, paths, and directories carry an implicit **ownership** boundary: a
program should only delete or overwrite what it created or was explicitly
told to manage. `filepath.Clean` is purely lexical — it rewrites `a/../b`
and repeated separators into a canonical form, but it does not check that a
path stays under an intended root, and it does nothing about symlinks, so it
is not an authorization or sandboxing mechanism. `os.Remove` deletes exactly
one file or empty directory and fails loudly on anything else, making it the
safe default; `os.RemoveAll` recurses without asking, so reserve it for a
directory the program just created and unambiguously owns (as in lesson 6's
temporary workspace), never for a path built from unvalidated input.
`os.Rename` is a same-filesystem move: it can fail across filesystems, and
its behavior when the destination already exists varies by operating
system, so treat it as a boundary to check rather than a universal move.

Finally, `time.Time` mixes several concepts that are easy to blur: an
**instant** (a specific point in time), a **location** (the time zone used
to display it, which `==` also compares — use `Time.Equal` to compare
instants alone), a **wall-clock** reading (suitable for formatting,
persistence, and RFC 3339 timestamps), and an optional **monotonic** reading
attached by `time.Now` (used internally by `Sub`/`time.Since` for reliable
elapsed-time measurement, but not preserved by formatting, JSON, or storage).
A `time.Duration` measures elapsed time, independent of any calendar date or
location. Code that needs the current instant should depend on a small
`Clock` interface rather than call `time.Now` directly, so tests can inject
a fixed time and assert exact results instead of tolerating timing noise.

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

## 🔬 Experiment

Combine lesson 4's strict JSON boundary with a small address-book record: add a
`CreatedAt string` field, parse it the way lesson 5 parses RFC 3339 timestamps
(`time.Parse(time.RFC3339, raw)`), wrapping a failure with `%w` so
`errors.As` can still recover the underlying `*time.ParseError`, then save
valid contacts into a `records/` directory created with `os.MkdirAll` (as in
lesson 6). Open the
JSON file with `os.Create`/`os.Open` and `defer file.Close()` on the very
next line in every function that touches it. Before overwriting an existing
address book, rename the old file to `records/archive/<name>.bak` with
`os.Rename`, wrapping any error with `%w`, then confirm with `errors.Is` and
`errors.As` that a missing source file and a malformed timestamp are both
still recoverable through every layer of wrapping.

Practice the same ideas with tests behind them in the matching exercise:
[`exercises/06_errors_files_json/`](../../exercises/06_errors_files_json/).

## 🔗 Related reading

- <https://go.dev/blog/errors-are-values>, <https://pkg.go.dev/errors>
- <https://pkg.go.dev/os>, <https://pkg.go.dev/io>,
  <https://pkg.go.dev/path/filepath>
- <https://pkg.go.dev/encoding/json>
- <https://pkg.go.dev/time>

# 🗂️ 06 — Errors, Files, Directories, JSON, and Time

Persist a small address book to a JSON file to practice typed errors, error
wrapping, `defer`-based resource cleanup, file I/O, and JSON validation.
The final three helpers practice timestamps and elapsed durations without
changing the address-book format. The directory helpers then practice creating,
walking, moving within, and safely cleaning up a temporary workspace.

## ▶️ Workflow

```bash
go test ./exercises/06_errors_files_json
go test ./exercises/06_errors_files_json/solution
```

The starter files in this package fail until you implement every function and
method. Compare with the matching files under `solution/` only after a genuine
attempt.

## 🧩 Tasks

1. Implement `(*ValidationError).Error() string`, returning `"<Field>:
   <Message>"`.
2. Implement `Contact.Validate() error`. Check, in order: `Name` is
   non-empty, `Email` contains `"@"`, `Age` is not negative. Return the first
   violation as a `*ValidationError`, or `nil` if the contact is valid.
3. Implement `LoadContacts(path string) ([]Contact, error)`. Open the file,
   `defer` closing it, and decode a `[]Contact` from its JSON contents. Wrap
   every error with `fmt.Errorf` and `%w` so a missing file still satisfies
   `errors.Is(err, os.ErrNotExist)` and a malformed file still lets callers
   recover the underlying `*json.SyntaxError` with `errors.As`.
4. Implement `SaveContacts(path string, contacts []Contact) error`. Validate
   every contact *before* touching the filesystem. Create (or truncate) the
   file, `defer` closing it, and encode the contacts as indented JSON,
   wrapping any error with `%w`.
5. Using the provided `ErrDuplicateEmail`, implement
   `AddContact(path string, c Contact) error`. Treat a missing file as an
   empty address book (check with `errors.Is(err, os.ErrNotExist)`), reject
   duplicate emails by wrapping `ErrDuplicateEmail` with `%w`, and otherwise
   append and save.
6. Using the provided `ErrNotFound`, implement
   `FindByEmail(contacts []Contact, email string) (Contact, error)`, wrapping
   `ErrNotFound` with `%w` and including the email in the message when nothing
   matches.
7. Implement `ParseTimestamp(raw string) (time.Time, error)` with
   `time.Parse(time.RFC3339, raw)`. Wrap parse failures with `%w` so callers can
   recover `*time.ParseError` with `errors.As`.
8. Implement `FormatTimestampUTC(value time.Time) string`, normalizing the
   instant with `UTC()` before formatting it with `time.RFC3339`.
9. Using the provided `ErrEndBeforeStart`, implement
   `Elapsed(start, end time.Time) (time.Duration, error)`. Return `end.Sub(start)`
   when the interval is non-negative; otherwise return
   `ErrEndBeforeStart`.
10. Implement `EnsureWorkspace(root string) error`, creating `inbox`, `archive`,
    and `reports/daily` with `os.MkdirAll`.
11. Implement `ListRegularFiles(root string) ([]string, error)` with
    `filepath.WalkDir`. Return regular files as deterministic, relative,
    slash-separated paths; wrap `ErrNotDirectory` when root is a file.
12. Implement `MoveFile(source, destination string) error`, creating the
    destination parent directory before calling `os.Rename`.
13. Implement `RemoveEmptyDirectory(path string) error` with `os.Stat` and
    `os.Remove`. Reject files with `ErrNotDirectory` and let `os.Remove` reject
    non-empty directories.

## 🔍 What this covers

- Typed errors (`*ValidationError`) recovered with `errors.As`.
- Sentinel errors (`ErrDuplicateEmail`, `ErrNotFound`) recovered with
  `errors.Is`, even through `fmt.Errorf("...: %w", ...)` wrapping.
- `defer` for deterministic file cleanup, including on early-return error
  paths.
- Reading and writing files and JSON with the standard library.
- Validating data before performing I/O.
- RFC 3339 timestamp parsing, UTC normalization, elapsed durations, and
  `Time.Equal`-based instant comparisons in tests.
- Directory creation, recursive deterministic listing, same-filesystem moves,
  and empty-directory removal under `t.TempDir`.

## ⚠️ Common mistakes

- Comparing errors with `==` or string matching instead of `errors.Is` /
  `errors.As`, which breaks once an error is wrapped.
- Wrapping with `fmt.Errorf("...: %v", err)` instead of `%w`, which loses the
  ability to unwrap.
- Forgetting `defer file.Close()` immediately after a successful `Open` or
  `Create`, which leaks the file handle on early returns.
- Validating contacts only after creating the output file, leaving a
  truncated or partial file behind on a validation failure.
- Comparing `time.Time` with `==` when the intent is to compare instants, or
  persisting a local display time without an offset.
- Returning absolute or platform-specific paths from `ListRegularFiles` instead
  of relative paths normalized with `filepath.ToSlash`.
- Using `os.RemoveAll` where the contract only calls for removing one verified
  empty directory.
- Assuming `os.Rename` copies data across filesystems or has portable overwrite
  behavior for an existing destination.

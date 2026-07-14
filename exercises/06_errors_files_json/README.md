# 🗂️ 06 — Errors, Files, and JSON

Persist a small address book to a JSON file to practice typed errors, error
wrapping, `defer`-based resource cleanup, file I/O, and JSON validation.

## ▶️ Workflow

```bash
go test ./exercises/06_errors_files_json
go test ./exercises/06_errors_files_json/solution
```

The starter package in this folder (`contacts.go`) fails until you implement
every function and method. Compare with `solution/contacts.go` only after a
genuine attempt.

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
5. Declare `ErrDuplicateEmail` and implement
   `AddContact(path string, c Contact) error`. Treat a missing file as an
   empty address book (check with `errors.Is(err, os.ErrNotExist)`), reject
   duplicate emails by wrapping `ErrDuplicateEmail` with `%w`, and otherwise
   append and save.
6. Declare `ErrNotFound` and implement
   `FindByEmail(contacts []Contact, email string) (Contact, error)`,
   wrapping `ErrNotFound` with `%w` and including the email in the message
   when nothing matches.

## 🔍 What this covers

- Typed errors (`*ValidationError`) recovered with `errors.As`.
- Sentinel errors (`ErrDuplicateEmail`, `ErrNotFound`) recovered with
  `errors.Is`, even through `fmt.Errorf("...: %w", ...)` wrapping.
- `defer` for deterministic file cleanup, including on early-return error
  paths.
- Reading and writing files and JSON with the standard library.
- Validating data before performing I/O.

## ⚠️ Common mistakes

- Comparing errors with `==` or string matching instead of `errors.Is` /
  `errors.As`, which breaks once an error is wrapped.
- Wrapping with `fmt.Errorf("...: %v", err)` instead of `%w`, which loses the
  ability to unwrap.
- Forgetting `defer file.Close()` immediately after a successful `Open` or
  `Create`, which leaks the file handle on early returns.
- Validating contacts only after creating the output file, leaving a
  truncated or partial file behind on a validation failure.

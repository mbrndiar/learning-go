# 🛠️ Exercises: Tooling, CLI, and Observability

This exercise applies
[Module 9](../../lessons/09_tooling_cli_observability/README.md). `cliapp` is
the testable core of a tiny greeting CLI; `cmd/taskcli` is the thin `main` that
wires it to the real process. Keep that split: `main` should only translate
`os.Args` into a call and an error into an exit code, so almost everything else
can be unit tested without spawning a subprocess.

## 🧩 Tasks

1. `ParseArgs` — parse `-name`, `-count`, `-verbose`, and `-log-format`
   using a dedicated `flag.NewFlagSet` (never the global
   `flag.CommandLine`), so it is safe to call repeatedly in tests. Validate
   that `-count >= 1` and `-log-format` is `"text"` or `"json"`, returning a
   descriptive error otherwise. Return `flag.ErrHelp` unchanged when `-h`/
   `-help` is requested, so callers can tell "print usage" apart from a real
   failure.
2. `NewLogger` — build an `*slog.Logger` writing to the given `io.Writer`:
   a JSON handler when `LogFormat == "json"`, a text handler otherwise, at
   `slog.LevelDebug` when `Verbose` else `slog.LevelInfo`.
3. `Greeting` — pure function returning `Count` copies of `"Hello, <Name>!"`.
4. `Diagnostics` — return a `map[string]string` with `go_version`, `os`,
   `arch`, `num_cpu`, and `num_goroutine`, built from the `runtime` package.
5. `Run` — wire the four functions above together: parse args, build a
   logger writing to `stderr`, print the greeting to `stdout`, log the
   parsed config at debug level and diagnostics at info level, and return an
   error instead of calling `os.Exit` (so `Run` itself stays testable).

## ▶️ Commands

```bash
go test ./exercises/09_tooling_cli_observability/...
go test -run '^$' ./exercises/09_tooling_cli_observability
go test ./exercises/09_tooling_cli_observability/solution
gofmt -l exercises/09_tooling_cli_observability
go run ./exercises/09_tooling_cli_observability/solution/cmd/taskcli -name Ada -count 2 -log-format json -verbose
```

## 📝 Notes

- A `main` function that only calls one function and checks its error is a
  deliberate pattern, not laziness: it lets every other behavior be reached
  by ordinary table-driven tests with buffers standing in for `stdout` and
  `stderr`.
- `flag.NewFlagSet(name, flag.ContinueOnError)` returns parse errors instead
  of calling `os.Exit`; pair it with `fs.SetOutput(io.Discard)` in library
  code so a library never prints to the real process's stderr on its own.
- `errors.Is(err, flag.ErrHelp)` is the idiomatic way to detect `-h`/`-help`
  without string-matching the error message.
- `slog.HandlerOptions{Level: ...}` controls the minimum level a handler
  emits; `slog.NewTextHandler` and `slog.NewJSONHandler` share the same
  options type.
- Compare with `solution/` only after a genuine attempt.

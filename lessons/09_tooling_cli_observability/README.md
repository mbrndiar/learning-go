# 🛠️ Module 09 — Tooling, CLI, and Observability

Go ships almost everything you need for day-to-day development inside one
`go` binary: builder, formatter, dependency manager, test runner, profiler,
and documentation browser. This module tours that toolchain and then builds
two small runnable programs that show how production Go code takes command
-line input and emits structured logs.

## 🎯 Learning goals

By the end of this module you will be able to:

- use `go run`, `go build`, `go install`, `go list`, and `go doc` to build,
  inspect, and read about Go code;
- manage dependencies with `go.mod` and `go mod tidy`;
- keep code formatted with `gofmt`/`go fmt` and catch suspicious code with
  `go vet`;
- describe what `staticcheck` and `govulncheck` add beyond `go vet`, and how
  to install and run them;
- build a command-line interface with the standard library's `flag`
  package;
- emit structured, leveled logs with `log/slog`;
- find data races with `go test -race` / `go run -race`; and
- describe how `go tool pprof` and Delve (`dlv`) fit into a debugging and
  performance workflow.

## 🗂️ Lesson map

1. [`01_cli_flags/`](01_cli_flags/) — a CLI built with the `flag` package.
2. [`02_structured_logging/`](02_structured_logging/) — structured, leveled
   logging with `log/slog`.
3. [`03_race_detector/`](03_race_detector/) — a real data race, a
   synchronized fix, and `-race` in action.
4. [`04_pprof_delve/`](04_pprof_delve/) — writing CPU/heap profiles with
   `runtime/pprof`, plus Delve orientation.

## ▶️ How to run these lessons

Unlike module 08, every lesson here is an independent `package main` you run
directly:

```bash
go run ./lessons/09_tooling_cli_observability/01_cli_flags -name Ada -times 2 -shout
go run ./lessons/09_tooling_cli_observability/02_structured_logging -format=json -level=debug
go run ./lessons/09_tooling_cli_observability/03_race_detector -mode=safe
go run -race ./lessons/09_tooling_cli_observability/03_race_detector -mode=race
go run ./lessons/09_tooling_cli_observability/04_pprof_delve
```

## 🔬 Topic notes

### Core `go` commands

```bash
go run ./path/to/package     # compile + run, discarding the binary afterward
go build ./path/to/package   # compile, write a binary in the current directory
go build ./...               # compile every package in the module (catches build errors early)
go install ./path/to/cmd     # build and copy the binary into $GOBIN (on your PATH)
go list ./...                 # list every package import path in the module
go list -f '{{.ImportPath}}: {{.GoFiles}}' ./lessons/09_tooling_cli_observability/01_cli_flags
go doc log/slog               # read package documentation in the terminal
go doc log/slog.Logger.With    # read documentation for one symbol
```

`go run` never leaves a binary behind; `go build`/`go install` do. Use
`go build ./...` (without running anything) as a fast "does everything still
compile" check across a whole module - exactly what CI normally runs first.

### `go.mod` and `go mod tidy`

- `go.mod` declares the module path and the minimum Go version
  (`go 1.25.0` at the root of this course) plus any dependencies.
- `go mod tidy` adds missing requirements and removes unused ones so
  `go.mod`/`go.sum` match what the code actually imports.
- This course's lessons use only the standard library, so `go.mod` needs no
  changes to run them; `go mod tidy` matters once you add a third-party
  import.

> This module's lessons must not touch the repository's root `go.mod` -
> everything here builds against the standard library.

### `gofmt` and `go vet`

```bash
gofmt -l .                    # list files that are not formatted (no changes made)
gofmt -w .                    # rewrite files in place
go fmt ./...                  # gofmt -l -w, module-aware, run per package
go vet ./...                  # static analysis for suspicious constructs
```

`gofmt` defines the one true formatting for Go source: never hand-align
code or argue about brace placement. `go vet` catches real mistakes that
still compile - a `Printf` verb that does not match its argument, a
copied `sync.Mutex`, an unreachable `case` - and CI should always run it.

### `staticcheck` and `govulncheck` orientation

Neither ships with the `go` binary; both are official or widely trusted
tools you install separately with `go install`:

```bash
go install honnef.co/go/tools/cmd/staticcheck@v0.7.0
staticcheck ./...
```

`staticcheck` finds a much larger set of bugs and style issues than
`go vet` - unused struct fields, simplifiable expressions, deprecated API
usage, and more.

```bash
go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
govulncheck ./...
```

`govulncheck` cross-references your dependencies (and which of their
functions you actually call) against the Go vulnerability database, and
reports only vulnerabilities reachable from your code. Run it after
`go mod tidy` whenever you add or update a dependency.

Both tools require installing a separate binary and network access to fetch
it. The versions above match CI so local and automated results stay
reproducible; update pins deliberately after reviewing release notes.

### A real local-to-CI workflow

The course workflow in [`.github/workflows/course.yml`](../../.github/workflows/course.yml)
runs on Go 1.25 and 1.26. A practical local sequence is:

```bash
go mod download
go mod tidy
git diff --exit-code -- go.mod go.sum
go fmt ./...
go vet ./...
go test ./lessons/...
go list ./exercises/... | grep -v '/solution$' | xargs go test -run '^$'
go list ./exercises/... | grep '/solution$' | xargs go test
go test ./project/...
go test -race ./project/... ./lessons/10_concurrency/... ./lessons/11_application_integration/... \
  ./exercises/10_concurrency/solution ./exercises/11_application_integration/solution
go test -coverprofile=coverage.out ./project/taskmanager ./project/taskapi ./project/taskclient
bash scripts/check-coverage.sh coverage.out 85
{ go list ./lessons/...; go list ./exercises/... | grep '/solution$'; go list ./project/...; go list ./tools/...; } |
  xargs staticcheck
govulncheck ./...
go run ./tools/checklinks
```

Starter exercises intentionally fail their behavior tests until completed, so
CI compiles them with `go test -run '^$'` and runs the separate `solution`
packages. The workflow also runs a short fuzz smoke test. GitHub Actions uses
the official [`actions/checkout`](https://github.com/actions/checkout) and
[`actions/setup-go`](https://github.com/actions/setup-go) actions.

### CLI design with `flag` (`01_cli_flags`)

- `flag.String`, `flag.Int`, `flag.Bool`, etc. register a flag and return a
  pointer to its value; call `flag.Parse()` (or `fs.Parse(args)` for a
  `flag.NewFlagSet`) once, after registering every flag.
- `flag.NewFlagSet(name, flag.ContinueOnError)` avoids the top-level
  `flag.CommandLine`'s default behavior of calling `os.Exit` on a parse
  error, which makes the parsing logic callable from a test.
- After `Parse`, `fs.Args()` returns the remaining non-flag arguments.
- Overriding `fs.Usage` controls the `-h`/`-help` message and the message
  printed on a parse error.
- Structuring `main` as a thin wrapper around a `run(args, stdout, stderr)`
  function (as in `01_cli_flags`) keeps `os.Exit` in exactly one place and
  makes the logic itself testable without spawning a subprocess.

### Structured logging with `log/slog` (`02_structured_logging`)

- `slog.New(handler)` builds a logger; `slog.NewTextHandler` and
  `slog.NewJSONHandler` are the two standard handlers, both writing to any
  `io.Writer`.
- Log with key-value pairs, not formatted strings:
  `logger.Info("server starting", "addr", ":8080")` - this keeps fields
  machine-parseable, unlike `fmt.Sprintf("server starting on %s", addr)`.
- `logger.With(...)` returns a child logger that always includes given
  attributes - handy for attaching request IDs or component names once.
- `slog.Group(...)` nests related attributes under one key in the output.
- `slog.LevelVar` holds a level that can change at runtime (for example
  from a config reload), unlike a plain `slog.Level` constant.
- `slog.SetDefault(logger)` makes package-level calls (`slog.Info`, ...)
  use your configured logger instead of the built-in default.
- Prefer the `*Context` methods (`InfoContext`, `WarnContext`, ...) in real
  services so a handler can pull trace/span IDs out of `context.Context`.

### The race detector (`03_race_detector`)

```bash
go run -race ./path/to/package         # race-check a single run
go test -race ./...                    # race-check while testing (most common use)
go build -race -o app ./cmd/app        # build a race-instrumented binary
```

The race detector instruments memory accesses and reports a `WARNING: DATA
RACE` whenever two goroutines access the same memory concurrently without
synchronization and at least one access is a write - exactly what
`racyCount` does in this lesson's `-mode=race` path. It only reports races
that actually happen during that run, so a race-detector-clean run does not
prove a program is race-free; it proves that run did not trigger one.
`-race` roughly doubles memory use and slows execution significantly, so it
is normally used in CI and local testing, not in production binaries.

This example intentionally previews goroutines, `sync.WaitGroup`, and
`sync.Mutex`; module 10 teaches those mechanisms in order. Here, treat them as
the setup needed to compare a racy execution with a synchronized one.

### Profiling and debugging orientation (`04_pprof_delve`)

`runtime/pprof` (used in `04_pprof_delve`) writes CPU and heap profiles to
files you inspect afterward - a good fit for short-lived programs and
tests:

```bash
go tool pprof -top cpu.pprof            # top functions by CPU time, in the terminal
go tool pprof -http=:0 cpu.pprof        # interactive graph in a browser
go tool pprof -top heap.pprof           # top allocations by current heap usage
```

For a long-running server, `net/http/pprof` is the usual alternative: importing it
for its side effect (`import _ "net/http/pprof"`) registers `/debug/pprof/*`
handlers on the default mux, and you fetch profiles over HTTP instead of
writing them from inside the program, e.g.
`go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30`.

Delve (`dlv`) is Go's interactive source-level debugger. It is a separate
tool (`go install github.com/go-delve/delve/cmd/dlv@latest`), so this
lesson documents it rather than depending on it at runtime:

```bash
dlv debug ./lessons/09_tooling_cli_observability/04_pprof_delve
(dlv) break main.isPrime
(dlv) continue
(dlv) print n
(dlv) next
(dlv) continue
```

`dlv debug` compiles with debugging information and starts an interactive
session; `break` sets a breakpoint, `continue` runs to the next breakpoint,
`next`/`step` advance line by line, and `print` inspects a variable. Editors
with the official Go extension (VS Code, GoLand) usually wrap the same
`dlv` binary behind a graphical debugger.

## ⚠️ Common mistakes

- Calling `flag.Parse()` before every flag is registered - flags declared
  afterward never get a value from the command line.
- Logging with `fmt.Printf`/string concatenation in a service that other
  tools are supposed to parse - structured fields survive log aggregation
  and querying; formatted strings do not.
- Treating a clean `-race` run as proof of no races: the detector only
  reports races that were actually triggered by that specific execution
  and inputs.
- Running `go vet`/`staticcheck` only in CI and never locally - both are
  fast enough to run before every commit.
- Forgetting `go mod tidy` after adding or removing an import, leaving
  `go.mod`/`go.sum` out of sync with the code.
- Profiling a program for far too short a duration (`04_pprof_delve` runs
  a few hundred milliseconds) and then over-interpreting a handful of
  samples - real profiling needs a representative, sufficiently long
  workload.
- Leaving `_ "net/http/pprof"` imported (and its endpoints reachable) in a
  publicly exposed production server without additional access control.

## ❓ Review questions

1. What is the practical difference between `go run` and `go build`, and
   why might CI prefer `go build ./...` over running every package?
2. What does `go mod tidy` do, and why should you run it right after
   adding a new import?
3. Why does `go vet` catch some bugs that still compile cleanly? Give an
   example.
4. What does `staticcheck` add on top of `go vet`? What does `govulncheck`
   check that neither of the other two does?
5. Why does structuring `main` as `os.Exit(run(args, stdout, stderr))`
   make a CLI's logic easier to test than putting everything directly in
   `func main()`?
6. Why prefer `logger.Info("msg", "key", value)` over
   `fmt.Sprintf("msg key=%v", value)` for logs a machine will later parse?
7. What problem does `slog.LevelVar` solve that a plain `slog.Level`
   constant does not?
8. What exactly does the race detector detect, and why is a race-free test
   run not a guarantee that the program has no data races?
9. When would you reach for `runtime/pprof` versus `net/http/pprof`?
10. What is the difference in workflow between reading a `go tool pprof`
    report after the fact and stepping through a program live with
    `dlv debug`?

## 🔗 Related reading

- <https://pkg.go.dev/cmd/go>
- <https://pkg.go.dev/flag>
- <https://pkg.go.dev/log/slog>
- <https://go.dev/doc/articles/race_detector>
- <https://pkg.go.dev/runtime/pprof>
- <https://pkg.go.dev/net/http/pprof>
- <https://github.com/go-delve/delve/tree/master/Documentation>
- <https://staticcheck.dev/>
- <https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck>

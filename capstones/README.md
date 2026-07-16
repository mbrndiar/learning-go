# 🏆 Capstones

This course has two independent capstones:

- [`comparative/`](comparative/README.md) implements the shared, versioned
  SQLite key/value specification used by every learning repository.
- [`idiomatic/`](idiomatic/README.md) builds a Go-specific service health
  monitor around interfaces, goroutines, contexts, and `net/http`.

Both capstones use matching `starter/` and `solution/` trees. Their exported
boundaries intentionally match so shared contract helpers and milestone tests
can target either implementation without changing behavior. Both solutions are
complete; both starters remain explicit guided TODOs.

Complete the required
[`Task REST API and clients`](../projects/tasks/README.md) project after Module
12 and before starting either capstone. Its finished three-server-by-two-client
solution applies the SQL, repository, HTTP, lifecycle, CLI, and testing
boundaries used here. Module 11 leads into the comparative SQLite capstone;
Modules 10 and 12 supply the concurrency, HTTP client, and shutdown skills used
by the idiomatic monitor.

These are the current course capstones. A different, superseded connected Task
codebase is available only in the last pre-removal snapshot, commit
[`b3211f9`](https://github.com/mbrndiar/learning-go/tree/b3211f99fc2ce5da54b88c59da3f12aacbed30ff/project)
at path `project/`. Use it as historical migration context rather than adding
new work there; the current required project lives under `projects/tasks/`.

## Targeted workflow

Compile every starter package without running learner tests:

```bash
go test -run '^$' \
  ./capstones/comparative/starter/... \
  ./capstones/idiomatic/starter/...
```

Run the starter harnesses, both complete solutions, and the API boundary check:

```bash
go test \
  ./capstones/comparative/starter/kvstore \
  ./capstones/comparative/solution/kvstore \
  ./capstones/idiomatic/starter/monitor \
  ./capstones/idiomatic/solution/monitor \
  ./capstones
```

Run all current solution and reusable test-support packages:

```bash
go test \
  ./capstones/comparative/solution/... \
  ./capstones/comparative/tests/... \
  ./capstones/idiomatic/solution/... \
  ./capstones/idiomatic/tests/...
```

The starter compile command remains useful after milestone tests are added:
starter tests may intentionally fail while unfinished, but every starter
package must always load and type-check.

For the repository-wide CI split—and why raw `go test ./...` intentionally runs
failing exercise starters—see [`docs/QUALITY.md`](../docs/QUALITY.md).

## Harness rules

- Keep corresponding starter and solution exported declarations identical.
- Put reusable assertions in each capstone's `tests/contract` package and call
  them from thin implementation-specific wrappers.
- Keep commands thin; test behavior through importable packages.
- Return `domain.ErrNotImplemented` from unfinished error-returning boundaries.
- Keep placeholder command/API output visibly non-conforming so it cannot be
  mistaken for a completed capstone.
- Change `domain.Implemented` to `true` only when replacing every placeholder
  covered by the harness.

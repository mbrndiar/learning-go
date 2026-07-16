# 🏆 Capstones

This course has two independent capstones:

- [`comparative/`](comparative/README.md) implements the shared, versioned
  SQLite key/value specification used by every learning repository.
- [`idiomatic/`](idiomatic/README.md) builds a Go-specific service health
  monitor around interfaces, goroutines, contexts, and `net/http`.

Both capstones use matching `starter/` and `solution/` trees. Their exported
boundaries intentionally match so shared contract helpers and later milestone
tests can target either implementation without changing imports or test logic.
At this harness stage both trees compile but deliberately report
`not implemented`; no key/value or monitoring behavior has been added yet.

## Targeted workflow

Compile every starter package without running learner tests:

```bash
go test -run '^$' \
  ./capstones/comparative/starter/... \
  ./capstones/idiomatic/starter/...
```

Run the harness smoke tests for both targets:

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

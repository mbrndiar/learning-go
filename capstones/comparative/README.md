# 🔁 Comparative Go capstone — versioned configuration store

The normative language-neutral contract is frozen under [`spec/`](spec/).
Read [`spec/SPEC.md`](spec/SPEC.md) and
[`spec/SCENARIOS.md`](spec/SCENARIOS.md) before implementing a milestone.

## Go layout

```text
starter/kvstore/       learner implementation target
solution/kvstore/      reference implementation target
  domain/              values, expectations, results, and structured errors
  storage/             consumer-facing storage and opener interfaces
  app/                 testable process boundary
  cmd/kvstore/         thin os.Args/os.Exit wiring
tests/contract/        reusable target-independent harness assertions
```

The two implementation trees expose the same Go API. Keep domain types free of
SQLite details, define storage behavior through the small `storage.Store`
interface, and keep `main` free of parsing or persistence logic. This prevents
import cycles and lets future milestone contracts inject a fake opener before
real SQLite process tests are introduced.

## Commands

Compile the starter:

```bash
go test -run '^$' ./capstones/comparative/starter/...
```

Run only its harness smoke test:

```bash
go test ./capstones/comparative/starter/kvstore
```

Run the solution target and shared support:

```bash
go test \
  ./capstones/comparative/solution/... \
  ./capstones/comparative/tests/...
```

Run either command target from the repository root:

```bash
go run ./capstones/comparative/starter/kvstore/cmd/kvstore --db store.db list
go run ./capstones/comparative/solution/kvstore/cmd/kvstore --db store.db list
```

Until implementation begins, both commands intentionally emit a non-normative
`not_implemented` envelope and exit `1`. They do not create a database.

Verify the copied shared specification:

```bash
(cd capstones/comparative/spec && sha256sum -c MANIFEST.sha256)
```

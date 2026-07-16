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
import cycles, lets milestone contracts inject a fake opener, and still leaves
the real SQLite/process boundary available to the conformance tests.

## Commands

Compile the starter:

```bash
go test -run '^$' ./capstones/comparative/starter/...
```

Run only its harness smoke test:

```bash
go test ./capstones/comparative/starter/kvstore
```

Run the complete solution and all five shared milestone suites:

```bash
go test ./capstones/comparative/solution/... ./capstones/comparative/tests/...
```

The solution test builds one repository-local `kvstore` executable, executes the
shared sequential fixtures, then launches independent child processes for
initialization, migration, CAS/delete races, and the 10-second busy cases.

Run one milestone while iterating:

```bash
go test ./capstones/comparative/solution/kvstore \
  -run 'TestMilestones/m3-storage'
```

Run either command target from the repository root:

```bash
go run ./capstones/comparative/starter/kvstore/cmd/kvstore --db store.db list
go run ./capstones/comparative/solution/kvstore/cmd/kvstore --db store.db list
```

The starter remains compileable and intentionally emits a non-normative
`not_implemented` envelope with exit `1`; it does not create a database. The
solution implements the frozen version-1 contract with the pinned pure-Go
`modernc.org/sqlite` driver.

For a race-instrumented real-process run, prebuild the launcher and point the
harness at it:

```bash
mkdir -p capstones/comparative/.conformance/race
go build -race \
  -o capstones/comparative/.conformance/race/kvstore \
  ./capstones/comparative/solution/kvstore/cmd/kvstore
COMPARATIVE_KV_PROGRAM="$PWD/capstones/comparative/.conformance/race/kvstore" \
  go test -race ./capstones/comparative/solution/kvstore
rm -rf capstones/comparative/.conformance/race
```

Verify the copied shared specification:

```bash
(cd capstones/comparative/spec && sha256sum -c MANIFEST.sha256)
```

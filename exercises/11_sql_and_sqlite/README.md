# 🗄️ Exercises: SQL and SQLite

This exercise applies [Module 11](../../lessons/11_sql_and_sqlite/README.md).
Implement the TODOs in `sqlite.go`. The tests use a real SQLite database file
under `t.TempDir()` and close it with `t.Cleanup`; there is no fake SQL
driver, so every task must actually satisfy real constraints and real
transaction semantics.

## 🔍 What this practices

- Schema constraints (`PRIMARY KEY`, `FOREIGN KEY`, `UNIQUE`, `CHECK`, an
  index) as durable rules the database enforces, not conventions your code
  must remember.
- Parameterized values versus fixed, trusted query structure.
- Row and nullable-safe mapping, and translating `sql.ErrNoRows` into a
  domain-specific `ErrNotFound`.
- Rows ownership: closing `*sql.Rows` and checking `rows.Err()` after a scan
  loop.
- The `BeginTx` / `defer tx.Rollback()` / `tx.Commit()` transaction shape,
  so a multi-statement change is all-or-nothing.
- Hiding persistence behind the narrow `TaskRepository` interface.

## 🧩 Tasks

1. Create relational schema with keys, constraints, and an index.
2. Implement parameterized task creation and row mapping, including
   not-found translation.
3. List all tasks or filter by done state, always in deterministic ID order.
4. Produce one project/task join and aggregate, including projects with no
   tasks.
5. Use a transaction so project rename, task rename, and audit insertion
   commit or roll back together.
6. Keep callers dependent on the narrow `TaskRepository` interface.

## ▶️ Commands

```bash
# Starter tests intentionally fail until the TODOs are implemented.
go test ./exercises/11_sql_and_sqlite
go test ./exercises/11_sql_and_sqlite/solution
go test -race ./exercises/11_sql_and_sqlite/solution
```

## ⚠️ Common mistakes

- Building SQL by formatting a caller-supplied value into the query string
  instead of passing it as an argument — the first test seeds a project
  named to look like a SQL injection attempt specifically to catch this.
- Forgetting `defer tx.Rollback()` right after `BeginTx`, which leaves a
  half-finished transaction (and its connection) open on any early return
  instead of undoing it.
- Returning the raw `sql.ErrNoRows` from `Get` instead of wrapping
  `ErrNotFound`, which breaks the `errors.Is(err, ErrNotFound)` checks in
  the tests.
- Leaving out `rows.Err()` after the `List` scan loop, which can hide a
  truncated result as if every row had been read.

## 📮 Feedback

Compare with the matching files under `solution/` only after a genuine
attempt; a failing `go test` output tells you exactly which task and
assertion still need work.

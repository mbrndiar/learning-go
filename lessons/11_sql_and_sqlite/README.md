# 🗄️ Module 11: Relational Databases and SQL with SQLite

**Prerequisites:** Modules 1–10, especially error wrapping and deterministic
resource cleanup with `defer` (Module 6), interfaces as narrow boundaries
(Modules 5 and 7), table-driven tests and `t.Cleanup`/`t.TempDir` (Module 8),
and `context.Context` cancellation (Module 10). Persistence code leans on all
four: every query can fail, every `*sql.DB`/`*sql.Rows` is a resource that
must be closed, a repository is only useful behind an interface, tests need
disposable databases, and every call takes a `context.Context` so a caller can
bound or cancel it.

This module teaches portable relational and SQL concepts through Go's
`database/sql` package and the pinned pure-Go `modernc.org/sqlite` driver.
Examples use isolated in-memory or temporary-file databases and close every
`*sql.DB` deterministically, so the emphasis stays on the `database/sql` API
and SQL itself rather than on running or administering a database server.

## 🎯 Objectives

By the end of this module you will be able to:

- model one-to-many relations with primary, foreign, unique, and check
  constraints, and explain why those constraints are durable rules enforced
  by the database rather than conventions your Go code must remember;
- explain `*sql.DB` as a handle to a pool of connections, not a single
  connection, and configure that pool deliberately (including why the
  in-memory SQLite lessons pin it to one connection);
- execute parameterized CRUD, distinguish query arguments from query
  structure, and map result rows — including nullable columns — into Go
  values;
- own the lifetime of a `*sql.Rows`: close it, drain it, and check
  `rows.Err()` after the loop;
- use joins, aggregates, ordering, filtering, and indexes to answer questions
  that span more than one table;
- make multi-statement changes atomic with `BeginTx`, a deferred `Rollback`,
  and an explicit `Commit`, and explain why that ordering is safe even after
  a successful commit;
- separate portable SQL and `database/sql` usage from SQLite-specific
  behavior, so the parts of this module that transfer to other databases are
  clear from the parts that do not;
- hide persistence behind a narrow repository interface, and explain what
  that boundary buys a caller (and what it costs).

## 📖 SQL and `database/sql`, explained

A relational **schema** is a set of durable rules the database enforces on
every write, not just documentation: a `PRIMARY KEY` guarantees uniqueness
and gives every row an identity other rows can reference, a `FOREIGN KEY`
guarantees a referenced row actually exists (an *orphan* insert fails instead
of silently succeeding), `UNIQUE` rejects duplicate values, and `CHECK`
rejects values that violate a business rule such as a non-negative balance.
Because the database enforces these constraints on every statement,
regardless of which code path performed the write, they are more reliable
than an equivalent check duplicated across every Go function that touches
the table.

`database/sql` draws a hard line between **values** and **query structure**.
A placeholder (`?` for SQLite) always represents one value passed as a
separate argument (`db.ExecContext(ctx, "... WHERE id = ?", id)`); the driver
sends it to the database apart from the SQL text, so it can never change
what the query does — only what it matches. Table names, column names,
`ORDER BY` direction, and other structural choices are not values and cannot
be parameterized; when a program needs to vary them, it must choose among a
small set of fixed, known-safe query strings written directly by the
programmer, never by formatting caller-supplied text into SQL.

`sql.Open` does not open a connection; it returns a `*sql.DB`, which is a
**concurrent-safe handle to a pool of connections** that Go opens, reuses,
and closes as needed. Multiple goroutines can share one `*sql.DB` and call
it concurrently without a mutex — that safety is part of its contract. Pool
size is configurable with `SetMaxOpenConns`, `SetMaxIdleConns`, and
`SetConnMaxLifetime`. This module's lessons call `db.SetMaxOpenConns(1)`
specifically because each *new* connection to SQLite's `:memory:` database
is a *separate, empty* database — the pool sharing one connection is what
keeps every statement in a lesson visible to every other statement. That
setting is a workaround for `:memory:`'s scope, not a general recommendation;
a file-backed SQLite database or another SQL database does not need it.

Scanning a row copies each column into a destination variable with
`Scan`; the number, order, and type compatibility of destinations must match
the `SELECT` list exactly. A column that can be SQL `NULL` cannot be scanned
directly into a plain Go type such as `string` or `time.Time` — use a
nullable wrapper (`sql.NullString`, `sql.NullTime`, ...), check its `Valid`
field, and convert it (often into a pointer) for the rest of the program.
`db.QueryRowContext(...).Scan(...)` reports `sql.ErrNoRows` when no row
matched; that is normal, expected control flow to translate into a
domain-specific "not found" error, not a condition to `log.Fatal` on.

`db.QueryContext` returns a `*sql.Rows`, which reserves a connection from the
pool while results are active. Exhausting all result sets closes it
automatically, but `defer rows.Close()` immediately after checking the error is
still the robust pattern because early returns may stop before exhaustion.
Call `rows.Next()` in a loop and `Scan` inside it; after the loop ends, always
check `rows.Err()`, since a `false` return from `Next()` can mean either "no
more rows" or "an error interrupted iteration," and only `rows.Err()` tells
them apart.

A **transaction**, started with `db.BeginTx(ctx, nil)`, groups several
statements so they commit or roll back together. The idiomatic shape is
`tx, err := db.BeginTx(...)`, then immediately `defer tx.Rollback()`, then the
statements, then `tx.Commit()`. Calling `Rollback` after a successful
`Commit` is documented to be a safe no-op (it returns `sql.ErrTxDone`, which
the deferred call discards), so the `defer` reliably undoes a half-finished
transaction on any early return — an error, a panic, or a business-rule
failure such as an insufficient balance — without also needing an `if err !=
nil { tx.Rollback() }` at every exit point.

Some of what this module teaches is portable and some is not. The relational
model, constraints, parameter binding, CRUD, joins, aggregates, indexes,
transactions, and the repository boundary all transfer to PostgreSQL, MySQL,
and other SQL databases; only placeholder spelling and how a driver reports a
generated ID tend to differ. `:memory:`, `PRAGMA foreign_keys`,
`sqlite_master`, SQLite's dynamic column typing, and SQLite's single-writer
locking behavior are SQLite-specific and called out as such in the lessons
below.

Finally, a **repository** is an interface — such as `TaskRepository` in
lesson 5 — that expresses persistence as domain operations (`Create`, `Find`,
`List`) instead of exposing `*sql.DB` to callers. Code written against the
interface can be tested against a real, disposable SQLite database instead of
a hand-written fake, and swapping the underlying store later touches only the
implementation. The trade-off is a layer of indirection: for a small program
talking to one store, a narrow interface can be more ceremony than benefit,
which is why this module introduces the pattern last, after the raw
`database/sql` mechanics it wraps.

## 🧭 Lessons

1. [`01_relational_model_and_database_sql/`](01_relational_model_and_database_sql/) —
   tables, keys, relations, constraints, and connection pools.
2. [`02_parameterized_crud_and_row_mapping/`](02_parameterized_crud_and_row_mapping/) —
   safe values, CRUD, `Scan`, nullable columns, and `sql.ErrNoRows`.
3. [`03_joins_aggregates_and_indexes/`](03_joins_aggregates_and_indexes/) —
   joins, grouped summaries, and index metadata.
4. [`04_transactions_and_sqlite/`](04_transactions_and_sqlite/) — commit,
   rollback, and SQLite connection pragmas.
5. [`05_repository_pattern/`](05_repository_pattern/) — a small domain
   interface backed by `database/sql`.

## ▶️ Running the lessons

Each lesson is its own runnable package:

```bash
go run ./lessons/11_sql_and_sqlite/01_relational_model_and_database_sql
go run ./lessons/11_sql_and_sqlite/02_parameterized_crud_and_row_mapping
go run ./lessons/11_sql_and_sqlite/03_joins_aggregates_and_indexes
go run ./lessons/11_sql_and_sqlite/04_transactions_and_sqlite
go run ./lessons/11_sql_and_sqlite/05_repository_pattern
```

Run every lesson's tests, then again with the race detector, from the
repository root:

```bash
go test ./lessons/11_sql_and_sqlite/...
go test -race ./lessons/11_sql_and_sqlite/...
```

## 🧪 Try it yourself

- In lesson 1, insert a task with a `project_id` that does not exist and
  observe the foreign-key violation; then run the same insert after removing
  `PRAGMA foreign_keys = ON` and see it silently succeed instead.
- In lesson 2, replace `sql.NullTime` temporarily with a plain `time.Time` and
  scan a row whose `completed_at` is `NULL`; observe the conversion error, then
  restore `sql.NullTime` and print `completed.Valid` before and after
  `completeTask`.
- In lesson 4, call `transfer` with a destination name that is absent, print both
  balances after the returned error, and confirm that the deferred rollback
  preserved the original values despite the first `UPDATE` having run inside
  the transaction.

## ⚠️ Common mistakes

- **Concatenating a value into SQL text instead of passing it as an
  argument.** Even values that "look safe," like an ID from a URL path,
  should always be passed through a placeholder; building SQL from
  user-controlled strings reopens the exact vulnerability parameters exist
  to close.
- **Treating `*sql.DB` as one connection.** It is a pool; a program does not
  need — and should not create — one `*sql.DB` per query or per goroutine.
  Open it once, keep it, and close it once when the program or test is done.
- **Scanning a nullable column straight into a non-nullable Go type.** This
  fails at runtime the first time the column is actually `NULL`; use the
  `sql.Null*` wrapper (or scan into a pointer) whenever a column allows
  `NULL`.
- **Forgetting `rows.Close()` or `rows.Err()`.** Fully exhausting rows closes
  them automatically, but an early return without `Close` can retain the pooled
  connection; a missing `rows.Err()` check can silently treat a truncated
  result set as if every row had been read successfully.
- **Checking a transaction's error but not deferring `Rollback` first.** If
  a statement after `BeginTx` panics or returns before an explicit
  `Rollback`/`Commit` decision is reached, the transaction (and the
  connection it holds) is left open until it times out on its own.
- **Believing a passing `go test` run without `-race` proves a repository is
  safe for concurrent callers.** `*sql.DB` is safe for concurrent use, but
  code layered on top of it (a cache, a counter, an in-memory index) may not
  be; run `go test -race` the same way Module 10 does.
- **Confusing SQLite quirks with SQL in general.** `PRAGMA foreign_keys`,
  `:memory:`'s per-connection scope, and dynamic column typing are SQLite
  behaviors; do not assume another database enforces foreign keys the same
  way or shares those defaults.

## ❓ Review questions

1. Why is a `FOREIGN KEY` constraint more reliable than a Go function that
   checks a project exists before inserting a task?
2. What does `*sql.DB` actually represent, and why can several goroutines
   safely share one without a mutex?
3. Why does this module's lessons call `db.SetMaxOpenConns(1)` for SQLite's
   `:memory:` database, and why would a file-backed database not need it?
4. What is the difference between a query argument and part of the query's
   structure, and why can only the former be a placeholder?
5. What does `sql.ErrNoRows` mean, and why is translating it into a
   domain-specific "not found" error usually better than treating it as an
   unexpected failure?
6. Why must a nullable column be scanned into a wrapper like `sql.NullTime`
   instead of directly into `time.Time`?
7. Why does the loop pattern for `*sql.Rows` need to check `rows.Err()`
   after `rows.Next()` returns `false`, instead of assuming iteration
   finished cleanly?
8. Why is `defer tx.Rollback()` immediately after `BeginTx` safe to leave in
   place even on the success path, after `tx.Commit()` has already run?
9. Which parts of this module's SQL and `database/sql` usage would carry
   over unchanged to PostgreSQL or MySQL, and which parts are SQLite-specific?
10. What does hiding `*sql.DB` behind a `TaskRepository` interface let a
    caller do that calling `database/sql` directly does not, and what does
    that boundary cost?

## 🏁 Checkpoint

Continue with [`exercises/11_sql_and_sqlite/`](../../exercises/11_sql_and_sqlite/README.md)
to build a schema, parameterized CRUD, a join/aggregate, and an atomic
multi-statement transaction behind a repository interface of your own.

## 🔗 Related reading

- <https://pkg.go.dev/database/sql>
- <https://go.dev/doc/database/cancel-operations>
- <https://www.sqlite.org/foreignkeys.html>
- <https://www.sqlite.org/inmemorydb.html>
- <https://www.sqlite.org/datatype3.html>
- <https://pkg.go.dev/modernc.org/sqlite>

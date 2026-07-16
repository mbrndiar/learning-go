# 🗄️ Module 11: Relational Databases and SQL with SQLite

This module teaches portable relational and SQL concepts through Go's
`database/sql` package and the pinned pure-Go `modernc.org/sqlite` driver.
Examples use isolated in-memory databases and close every `*sql.DB`
deterministically.

## Objectives

- model one-to-many relations with primary, foreign, unique, and check constraints;
- execute parameterized CRUD and map rows, including nullable values;
- use joins, aggregates, ordering, filtering, and indexes;
- make multi-statement changes atomic with transactions and rollback;
- hide persistence behind a narrow repository interface.

## Lessons

1. [`01_relational_model_and_database_sql/`](01_relational_model_and_database_sql/) — tables, keys, relations, constraints, and connection pools.
2. [`02_parameterized_crud_and_row_mapping/`](02_parameterized_crud_and_row_mapping/) — safe values, CRUD, `Scan`, and `sql.ErrNoRows`.
3. [`03_joins_aggregates_and_indexes/`](03_joins_aggregates_and_indexes/) — joins, grouped summaries, and index metadata.
4. [`04_transactions_and_sqlite/`](04_transactions_and_sqlite/) — commit, rollback, and SQLite connection pragmas.
5. [`05_repository_pattern/`](05_repository_pattern/) — a small domain interface backed by `database/sql`.

## Portable SQL versus SQLite specifics

The relational model, parameter binding, CRUD, joins, aggregates,
transactions, and repository boundary transfer to other SQL databases.
Placeholder spelling and generated-ID APIs can vary by driver. `:memory:`,
`PRAGMA foreign_keys`, `sqlite_master`, dynamic typing, and SQLite's locking
behavior are SQLite-specific. The examples set `MaxOpenConns(1)` because each
SQLite `:memory:` connection otherwise owns a different database.

## Run

```bash
go test ./lessons/11_sql_and_sqlite/...
go test -race ./lessons/11_sql_and_sqlite/...
go run ./lessons/11_sql_and_sqlite/01_relational_model_and_database_sql
```

Always pass values as query arguments, close rows, check `rows.Err()`, defer
`tx.Rollback()` immediately after `BeginTx`, and close the database when its
owner is done.

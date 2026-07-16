# 🗄️ Exercises: SQL and SQLite

Implement the TODOs in `sqlite.go`. The tests use a real SQLite file under
`t.TempDir()` and close each database with `t.Cleanup`; there is no fake SQL
driver.

## Tasks

1. Create relational schema with keys, constraints, and an index.
2. Implement parameterized task creation and row mapping, including not-found translation.
3. List all tasks or filter by done state, always in deterministic ID order.
4. Produce one project/task join and aggregate, including projects with no tasks.
5. Use a transaction so project rename, task rename, and audit insertion commit or roll back together.
6. Keep callers dependent on the narrow `TaskRepository` interface.

```bash
# Starter tests intentionally fail until the TODOs are implemented.
go test ./exercises/11_sql_and_sqlite
go test ./exercises/11_sql_and_sqlite/solution
go test -race ./exercises/11_sql_and_sqlite/solution
```

Use parameters for values, fixed query text for ordering/filter choices, close
rows, check `rows.Err()`, and defer rollback immediately after `BeginTx`.

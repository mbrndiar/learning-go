package taskapi

// This file is test/exercise infrastructure, not a task: do not edit it.
//
// It implements a minimal, in-memory database/sql driver so SQLTaskStore
// (which you implement in store.go) can be exercised through the real
// database/sql API -- Exec, Query, Scan, sql.ErrNoRows -- without a real
// database or a SQLite dependency. It only needs to understand the three
// fixed queries this package uses (see the query* constants below); it does
// not parse arbitrary SQL.

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sort"
	"sync"
)

// Fixed queries recognized by the fake driver. SQLTaskStore must use these
// exact strings so the fake conn can recognize which operation to perform.
const (
	queryInsertTask = "INSERT INTO tasks (title, done, due_date) VALUES (?, ?, ?)"
	queryGetTask    = "SELECT id, title, done, due_date FROM tasks WHERE id = ?"
	queryListTasks  = "SELECT id, title, done, due_date FROM tasks ORDER BY id"
)

// OpenFakeDB returns a *sql.DB backed by a brand-new, isolated in-memory
// database: every call returns an independent database, even if called
// twice with the same name (name is purely a label used in error messages,
// e.g. in case a future version of this fake surfaces it in diagnostics).
// It uses sql.OpenDB with a driver.Connector instead of the global
// sql.Register/sql.Open registry, so no process-wide name collisions are
// possible between tests, including repeated runs of the same test name via
// `go test -count`.
func OpenFakeDB(name string) (*sql.DB, error) {
	return sql.OpenDB(&fakeConnector{db: &fakeDB{}}), nil
}

// fakeConnector implements driver.Connector, handing out connections that
// all share the single *fakeDB captured when OpenFakeDB created it.
type fakeConnector struct {
	db *fakeDB
}

func (c *fakeConnector) Connect(context.Context) (driver.Conn, error) {
	return &fakeConn{db: c.db}, nil
}

func (c *fakeConnector) Driver() driver.Driver {
	return fakeDriverStateless{}
}

// fakeDriverStateless is only present to satisfy driver.Connector.Driver;
// database/sql never calls Open on it because every connection is created
// through fakeConnector.Connect instead.
type fakeDriverStateless struct{}

func (fakeDriverStateless) Open(name string) (driver.Conn, error) {
	return nil, fmt.Errorf("taskapi fake driver: direct sql.Open is not supported, use OpenFakeDB")
}

// fakeRow is one stored row: id is the primary key, title/done/dueDate hold
// the column values. dueDate is nil when the column is NULL.
type fakeRow struct {
	id      int64
	title   string
	done    bool
	dueDate *string
}

// fakeDB is the in-memory table shared by every connection opened with the
// same DSN.
type fakeDB struct {
	mu     sync.Mutex
	rows   []fakeRow
	nextID int64
}

// fakeConn implements driver.Conn plus the context-aware ExecerContext and
// QueryerContext interfaces, so database/sql calls ExecContext/QueryContext
// on it directly without ever going through driver.Stmt.
type fakeConn struct {
	db *fakeDB
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("taskapi fake driver: Prepare is not supported, use ExecContext/QueryContext (query: %s)", query)
}

func (c *fakeConn) Close() error { return nil }

func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("taskapi fake driver: transactions are not supported")
}

func (c *fakeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if query != queryInsertTask {
		return nil, fmt.Errorf("taskapi fake driver: unsupported exec query: %s", query)
	}
	if len(args) != 3 {
		return nil, fmt.Errorf("taskapi fake driver: expected 3 args for insert, got %d", len(args))
	}

	title, ok := args[0].Value.(string)
	if !ok {
		return nil, fmt.Errorf("taskapi fake driver: title arg must be string, got %T", args[0].Value)
	}
	done, ok := args[1].Value.(bool)
	if !ok {
		return nil, fmt.Errorf("taskapi fake driver: done arg must be bool, got %T", args[1].Value)
	}
	var dueDate *string
	switch v := args[2].Value.(type) {
	case nil:
		dueDate = nil
	case string:
		dueDate = &v
	default:
		return nil, fmt.Errorf("taskapi fake driver: due_date arg must be string or nil, got %T", args[2].Value)
	}

	c.db.mu.Lock()
	defer c.db.mu.Unlock()
	c.db.nextID++
	id := c.db.nextID
	c.db.rows = append(c.db.rows, fakeRow{id: id, title: title, done: done, dueDate: dueDate})

	return fakeResult{lastInsertID: id, rowsAffected: 1}, nil
}

func (c *fakeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	switch query {
	case queryGetTask:
		if len(args) != 1 {
			return nil, fmt.Errorf("taskapi fake driver: expected 1 arg for get, got %d", len(args))
		}
		id, ok := args[0].Value.(int64)
		if !ok {
			return nil, fmt.Errorf("taskapi fake driver: id arg must be int64, got %T", args[0].Value)
		}
		for _, r := range c.db.rows {
			if r.id == id {
				return &fakeRows{rows: []fakeRow{r}}, nil
			}
		}
		return &fakeRows{rows: nil}, nil

	case queryListTasks:
		rows := make([]fakeRow, len(c.db.rows))
		copy(rows, c.db.rows)
		sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })
		return &fakeRows{rows: rows}, nil

	default:
		return nil, fmt.Errorf("taskapi fake driver: unsupported query: %s", query)
	}
}

// fakeResult implements driver.Result.
type fakeResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r fakeResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

// fakeRows implements driver.Rows over a fixed, already-materialized slice.
type fakeRows struct {
	rows []fakeRow
	pos  int
}

func (r *fakeRows) Columns() []string { return []string{"id", "title", "done", "due_date"} }
func (r *fakeRows) Close() error      { return nil }

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.pos]
	r.pos++
	dest[0] = row.id
	dest[1] = row.title
	dest[2] = row.done
	if row.dueDate == nil {
		dest[3] = nil
	} else {
		dest[3] = *row.dueDate
	}
	return nil
}

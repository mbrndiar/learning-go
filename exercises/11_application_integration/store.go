package taskapi

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SQLTaskStore is a TaskStore backed by database/sql. In production it
// would wrap a *sql.DB opened against a real database driver; exercises and
// tests instead open it against the in-memory fake driver via OpenFakeDB,
// so the same Create/Get/List code is exercised either way.
type SQLTaskStore struct {
	db *sql.DB
}

// NewSQLTaskStore wraps db as a TaskStore.
func NewSQLTaskStore(db *sql.DB) *SQLTaskStore {
	return &SQLTaskStore{db: db}
}

// dueDateArg converts t's optional DueDate into a value database/sql can
// bind as a query argument: nil when unset, or an RFC 3339 string when set.
func dueDateArg(t Task) any {
	if t.DueDate == nil {
		return nil
	}
	return t.DueDate.UTC().Format(time.RFC3339)
}

// scanDueDate converts a nullable due_date column value back into *time.Time.
func scanDueDate(raw sql.NullString) (*time.Time, error) {
	if !raw.Valid {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw.String)
	if err != nil {
		return nil, fmt.Errorf("parsing due_date %q: %w", raw.String, err)
	}
	return &parsed, nil
}

// Create persists t (ignoring t.ID) and returns t with its assigned ID.
//
// TODO(task 2): implement Create using s.db.ExecContext with the
// queryInsertTask statement (title, done, and dueDateArg(t) as arguments),
// then res.LastInsertId to fill in the returned Task's ID. Wrap any error
// with fmt.Errorf including %w.
func (s *SQLTaskStore) Create(ctx context.Context, t Task) (Task, error) {
	panic("not implemented")
}

// Get returns the task with the given id, or ErrNotFound if none exists.
//
// TODO(task 3): implement Get using s.db.QueryRowContext with the
// queryGetTask statement and row.Scan into (&id, &title, &done, &dueDate)
// where dueDate is a sql.NullString. Translate sql.ErrNoRows into
// ErrNotFound with errors.Is; use scanDueDate to convert the scanned column.
func (s *SQLTaskStore) Get(ctx context.Context, id int64) (Task, error) {
	panic("not implemented")
}

// List returns every task, ordered by ID ascending.
//
// TODO(task 4): implement List using s.db.QueryContext with the
// queryListTasks statement. Remember to defer rows.Close(), Scan each row
// the same way as Get, and check rows.Err() after the loop.
func (s *SQLTaskStore) List(ctx context.Context) ([]Task, error) {
	panic("not implemented")
}

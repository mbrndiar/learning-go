package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	_ "modernc.org/sqlite"
)

const schema = `CREATE TABLE tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	completed INTEGER NOT NULL CHECK (completed IN (0, 1))
)`

const initializeSchema = `CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	completed INTEGER NOT NULL CHECK (completed IN (0, 1))
)`

// Repository stores tasks in a process-owned SQLite connection pool.
type Repository struct {
	db        *sql.DB
	closeOnce sync.Once
	closeErr  error
}

var _ task.Repository = (*Repository)(nil)

// Open opens path, initializes its schema, and returns its owning
// repository. It is a compatibility wrapper around OpenContext for callers
// that have no context to propagate.
func Open(path string) (*Repository, error) {
	return OpenContext(context.Background(), path)
}

// OpenContext opens path and initializes its schema using ctx, and returns
// its owning repository. ctx governs the schema initialization statement and
// its read-back; canceling it aborts a slow initialize instead of blocking
// indefinitely, and a context that is already done when OpenContext is
// called fails immediately without creating or altering the database file.
func OpenContext(ctx context.Context, path string) (*Repository, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, task.WrapStorage("open sqlite", err)
	}
	dsn := sqliteDSN(absolute)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, task.WrapStorage("open sqlite", err)
	}
	// Cap the pool at one connection: modernc.org/sqlite serializes access
	// per connection, and concurrent connections to the same file risk
	// SQLITE_BUSY errors despite the busy_timeout pragma.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	repository := &Repository{db: db}
	if err := repository.initialize(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repository, nil
}

func sqliteDSN(absolute string) string {
	uriPath := filepath.ToSlash(absolute)
	if len(uriPath) >= 3 &&
		((uriPath[0] >= 'A' && uriPath[0] <= 'Z') || (uriPath[0] >= 'a' && uriPath[0] <= 'z')) &&
		uriPath[1] == ':' && uriPath[2] == '/' {
		uriPath = "/" + uriPath
	}
	dsn := &url.URL{Scheme: "file", Path: uriPath}
	query := url.Values{}
	query.Set("_pragma", "busy_timeout(5000)")
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

// Close closes the repository's database exactly once.
func (r *Repository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	r.closeOnce.Do(func() {
		r.closeErr = r.db.Close()
	})
	return r.closeErr
}

func (r *Repository) initialize(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, initializeSchema); err != nil {
		return task.WrapStorage("initialize sqlite schema", err)
	}

	// A pre-existing database file may already define a "tasks" table with
	// an incompatible shape (e.g. missing columns or constraints). Read back
	// its actual definition and compare it against the schema this package
	// expects, rather than trusting that CREATE TABLE IF NOT EXISTS succeeded
	// against a matching table.
	var statement string
	err := r.db.QueryRowContext(ctx, `
		SELECT sql
		FROM sqlite_master
		WHERE type = 'table' AND name = 'tasks'
	`).Scan(&statement)
	if err != nil {
		return task.WrapStorage("inspect sqlite schema", err)
	}
	if canonicalSQL(statement) != canonicalSQL(schema) {
		return task.WrapStorage("inspect sqlite schema", fmt.Errorf("incompatible tasks schema"))
	}
	return nil
}

// Create inserts one incomplete task atomically.
func (r *Repository) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO tasks (title, completed) VALUES (?, ?)`,
		input.Title, 0,
	)
	if err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	created, err := queryTask(ctx, tx, id)
	if err != nil {
		return task.Task{}, mapQueryError("create task", id, err)
	}
	if err := tx.Commit(); err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	return created, nil
}

// List returns tasks in ascending ID order, optionally filtered by completion.
func (r *Repository) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	statement := `SELECT id, title, completed FROM tasks`
	var args []any
	if filter.Completed != nil {
		statement += ` WHERE completed = ?`
		args = append(args, boolInteger(*filter.Completed))
	}
	statement += ` ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, task.WrapStorage("list tasks", err)
	}
	defer rows.Close()

	tasks := make([]task.Task, 0)
	for rows.Next() {
		value, err := scanTask(rows)
		if err != nil {
			return nil, task.WrapStorage("list tasks", err)
		}
		tasks = append(tasks, value)
	}
	if err := rows.Err(); err != nil {
		return nil, task.WrapStorage("list tasks", err)
	}
	return tasks, nil
}

// Get returns one task by ID.
func (r *Repository) Get(ctx context.Context, id int64) (task.Task, error) {
	value, err := queryTask(ctx, r.db, id)
	if err != nil {
		return task.Task{}, mapQueryError("get task", id, err)
	}
	return value, nil
}

// Update atomically applies the supplied fields to one task.
func (r *Repository) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.Task{}, task.WrapStorage("update task", err)
	}
	defer tx.Rollback()

	current, err := queryTask(ctx, tx, id)
	if err != nil {
		return task.Task{}, mapQueryError("update task", id, err)
	}
	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Completed != nil {
		current.Completed = *input.Completed
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE tasks SET title = ?, completed = ? WHERE id = ?`,
		current.Title, boolInteger(current.Completed), id,
	); err != nil {
		return task.Task{}, task.WrapStorage("update task", err)
	}
	updated, err := queryTask(ctx, tx, id)
	if err != nil {
		return task.Task{}, mapQueryError("update task", id, err)
	}
	if err := tx.Commit(); err != nil {
		return task.Task{}, task.WrapStorage("update task", err)
	}
	return updated, nil
}

// Delete atomically removes one task.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return task.WrapStorage("delete task", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return task.WrapStorage("delete task", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return task.WrapStorage("delete task", err)
	}
	if affected == 0 {
		return task.NewNotFoundError(id)
	}
	if err := tx.Commit(); err != nil {
		return task.WrapStorage("delete task", err)
	}
	return nil
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type scanner interface {
	Scan(...any) error
}

func queryTask(ctx context.Context, queryer queryer, id int64) (task.Task, error) {
	return scanTask(queryer.QueryRowContext(ctx,
		`SELECT id, title, completed FROM tasks WHERE id = ?`,
		id,
	))
}

func scanTask(row scanner) (task.Task, error) {
	var value task.Task
	var completed int64
	if err := row.Scan(&value.ID, &value.Title, &completed); err != nil {
		return task.Task{}, err
	}
	switch completed {
	case 0:
		value.Completed = false
	case 1:
		value.Completed = true
	default:
		return task.Task{}, fmt.Errorf("invalid completed value %d for task %d", completed, value.ID)
	}
	if err := task.ValidateTask(value); err != nil {
		return task.Task{}, fmt.Errorf("invalid task %d: %w", value.ID, err)
	}
	return value, nil
}

func mapQueryError(operation string, id int64, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return task.NewNotFoundError(id)
	}
	return task.WrapStorage(operation, err)
}

func boolInteger(value bool) int {
	if value {
		return 1
	}
	return 0
}

// canonicalSQL strips whitespace, semicolons, and case so that equivalent
// but differently formatted CREATE TABLE statements compare as equal.
func canonicalSQL(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == ';' {
			return -1
		}
		return unicode.ToLower(r)
	}, value)
}

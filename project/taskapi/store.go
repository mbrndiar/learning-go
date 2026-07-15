// Package taskapi persists tasks in SQLite and serves them as JSON over HTTP.
//
// The package is split into two collaborating parts: SQLiteStore owns
// persistence through database/sql and the pure-Go modernc.org/sqlite driver,
// while API owns HTTP transport, request validation, and structured errors.
// Both are context-aware so callers control cancellation and timeouts.
package taskapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

// MaxTitleLength bounds task titles in Unicode code points.
const MaxTitleLength = 256

// ErrNotFound reports that a requested task does not exist. Callers should
// branch with errors.Is(err, ErrNotFound) rather than inspecting SQL errors.
var ErrNotFound = errors.New("task not found")

// ErrEmptyTitle reports that a task title was empty after trimming whitespace.
var ErrEmptyTitle = errors.New("task title must not be empty")

// ErrTitleTooLong reports that a task title exceeded MaxTitleLength.
var ErrTitleTooLong = errors.New("task title is too long")

// ErrInvalidTitle reports that a task title contained invalid UTF-8.
var ErrInvalidTitle = errors.New("task title is not valid UTF-8")

// Task is the persisted task record.
type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// SQLiteStore stores tasks in a SQLite database. The zero value is not usable;
// construct it with OpenSQLiteStore. A SQLiteStore is safe for concurrent use
// because database/sql manages the underlying connection pool.
type SQLiteStore struct {
	db *sql.DB
}

// OpenSQLiteStore opens (creating if necessary) the SQLite database at dsn and
// ensures the schema exists. Use the path ":memory:" for an ephemeral store or
// a file path for durable storage. The context bounds schema initialization.
func OpenSQLiteStore(ctx context.Context, dsn string) (*SQLiteStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("taskapi: data source name must not be empty")
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("taskapi: open database: %w", err)
	}

	// SQLite permits only one writer at a time. Restricting database/sql to one
	// connection serializes this teaching application's reads and writes,
	// avoiding competing transactions and "database is locked" errors while
	// the HTTP server itself remains safe to call concurrently.
	db.SetMaxOpenConns(1)

	store := &SQLiteStore{db: db}
	if err := store.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

// Close releases the underlying database resources.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// init creates the tasks table if it does not already exist.
func (s *SQLiteStore) init(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS tasks (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT    NOT NULL,
    done  INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1))
);`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("taskapi: initialize schema: %w", err)
	}
	return nil
}

// List returns every task ordered by identifier.
func (s *SQLiteStore) List(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, done FROM tasks ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("taskapi: query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Done); err != nil {
			return nil, fmt.Errorf("taskapi: scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("taskapi: iterate tasks: %w", err)
	}
	return tasks, nil
}

// Get returns the task with the given identifier, or ErrNotFound.
func (s *SQLiteStore) Get(ctx context.Context, id int64) (Task, error) {
	var task Task
	err := s.db.QueryRowContext(ctx, `SELECT id, title, done FROM tasks WHERE id = ?`, id).
		Scan(&task.ID, &task.Title, &task.Done)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Task{}, fmt.Errorf("taskapi: get task %d: %w", id, ErrNotFound)
	case err != nil:
		return Task{}, fmt.Errorf("taskapi: get task %d: %w", id, err)
	}
	return task, nil
}

// Add validates and stores a new task, returning it with its assigned id.
func (s *SQLiteStore) Add(ctx context.Context, title string) (Task, error) {
	normalized, err := normalizeTitle(title)
	if err != nil {
		return Task{}, err
	}

	result, err := s.db.ExecContext(ctx, `INSERT INTO tasks (title, done) VALUES (?, 0)`, normalized)
	if err != nil {
		return Task{}, fmt.Errorf("taskapi: insert task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("taskapi: read inserted id: %w", err)
	}
	return Task{ID: id, Title: normalized, Done: false}, nil
}

// Complete marks the task with the given identifier as done and returns it.
// It reports ErrNotFound when no such task exists.
func (s *SQLiteStore) Complete(ctx context.Context, id int64) (Task, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET done = 1 WHERE id = ?`, id)
	if err != nil {
		return Task{}, fmt.Errorf("taskapi: complete task %d: %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Task{}, fmt.Errorf("taskapi: complete task %d: %w", id, err)
	}
	if affected == 0 {
		return Task{}, fmt.Errorf("taskapi: complete task %d: %w", id, ErrNotFound)
	}
	return s.Get(ctx, id)
}

// Remove deletes the task with the given identifier. It reports ErrNotFound
// when no such task exists.
func (s *SQLiteStore) Remove(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("taskapi: remove task %d: %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("taskapi: remove task %d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("taskapi: remove task %d: %w", id, ErrNotFound)
	}
	return nil
}

// normalizeTitle trims and validates a candidate title.
func normalizeTitle(title string) (string, error) {
	if !utf8.ValidString(title) {
		return "", ErrInvalidTitle
	}
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", ErrEmptyTitle
	}
	length := utf8.RuneCountInString(trimmed)
	if length > MaxTitleLength {
		return "", fmt.Errorf("%w: %d runes", ErrTitleTooLong, length)
	}
	return trimmed, nil
}

package solution

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SQLTaskStore is a TaskStore backed by database/sql.
type SQLTaskStore struct {
	db *sql.DB
}

// NewSQLTaskStore wraps db as a TaskStore.
func NewSQLTaskStore(db *sql.DB) *SQLTaskStore {
	return &SQLTaskStore{db: db}
}

func dueDateArg(t Task) any {
	if t.DueDate == nil {
		return nil
	}
	return t.DueDate.UTC().Format(time.RFC3339)
}

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
func (s *SQLTaskStore) Create(ctx context.Context, t Task) (Task, error) {
	res, err := s.db.ExecContext(ctx, queryInsertTask, t.Title, t.Done, dueDateArg(t))
	if err != nil {
		return Task{}, fmt.Errorf("creating task: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("creating task: reading assigned id: %w", err)
	}
	t.ID = id
	return t, nil
}

// Get returns the task with the given id, or ErrNotFound if none exists.
func (s *SQLTaskStore) Get(ctx context.Context, id int64) (Task, error) {
	row := s.db.QueryRowContext(ctx, queryGetTask, id)

	var t Task
	var dueDate sql.NullString
	if err := row.Scan(&t.ID, &t.Title, &t.Done, &dueDate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("getting task %d: %w", id, err)
	}

	due, err := scanDueDate(dueDate)
	if err != nil {
		return Task{}, fmt.Errorf("getting task %d: %w", id, err)
	}
	t.DueDate = due
	return t, nil
}

// List returns every task, ordered by ID ascending.
func (s *SQLTaskStore) List(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, queryListTasks)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var dueDate sql.NullString
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &dueDate); err != nil {
			return nil, fmt.Errorf("listing tasks: %w", err)
		}
		due, err := scanDueDate(dueDate)
		if err != nil {
			return nil, fmt.Errorf("listing tasks: %w", err)
		}
		t.DueDate = due
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	return tasks, nil
}

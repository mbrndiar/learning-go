// Package solution is the reference implementation for the SQL exercises.
package solution

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("record not found")

type Task struct {
	ID        int64
	ProjectID int64
	Title     string
	Done      bool
}

type ProjectSummary struct {
	Project   string
	TaskCount int
	DoneCount int
}

type TaskRepository interface {
	Create(context.Context, Task) (Task, error)
	Get(context.Context, int64) (Task, error)
	List(context.Context, *bool) ([]Task, error)
}

type SQLiteRepository struct{ db *sql.DB }

func OpenDatabase(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func CreateSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE projects (
 id INTEGER PRIMARY KEY,
 name TEXT NOT NULL UNIQUE
);
CREATE TABLE tasks (
 id INTEGER PRIMARY KEY,
 project_id INTEGER NOT NULL REFERENCES projects(id),
 title TEXT NOT NULL CHECK(length(trim(title)) > 0),
 done BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_tasks_project_done ON tasks(project_id, done);
CREATE TABLE task_audit (
 id INTEGER PRIMARY KEY,
 task_id INTEGER NOT NULL REFERENCES tasks(id),
 action TEXT NOT NULL
);`)
	return err
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository { return &SQLiteRepository{db: db} }

func (r *SQLiteRepository) Create(ctx context.Context, task Task) (Task, error) {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO tasks(project_id, title, done) VALUES (?, ?, ?)",
		task.ProjectID, task.Title, task.Done)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("created task id: %w", err)
	}
	return r.Get(ctx, id)
}

func (r *SQLiteRepository) Get(ctx context.Context, id int64) (Task, error) {
	var task Task
	err := r.db.QueryRowContext(ctx,
		"SELECT id, project_id, title, done FROM tasks WHERE id = ?", id,
	).Scan(&task.ID, &task.ProjectID, &task.Title, &task.Done)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, fmt.Errorf("task %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return Task{}, fmt.Errorf("get task %d: %w", id, err)
	}
	return task, nil
}

func (r *SQLiteRepository) List(ctx context.Context, done *bool) ([]Task, error) {
	query := "SELECT id, project_id, title, done FROM tasks ORDER BY id"
	var rows *sql.Rows
	var err error
	if done == nil {
		rows, err = r.db.QueryContext(ctx, query)
	} else {
		rows, err = r.db.QueryContext(ctx,
			"SELECT id, project_id, title, done FROM tasks WHERE done = ? ORDER BY id", *done)
	}
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Done); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}

func CreateProject(ctx context.Context, db *sql.DB, name string) (int64, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO projects(name) VALUES (?)", name)
	if err != nil {
		return 0, fmt.Errorf("create project: %w", err)
	}
	return result.LastInsertId()
}

func ProjectSummaries(ctx context.Context, db *sql.DB) ([]ProjectSummary, error) {
	rows, err := db.QueryContext(ctx, `
SELECT p.name, COUNT(t.id), COALESCE(SUM(CASE WHEN t.done THEN 1 ELSE 0 END), 0)
FROM projects AS p
LEFT JOIN tasks AS t ON t.project_id = p.id
GROUP BY p.id, p.name
ORDER BY p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var summaries []ProjectSummary
	for rows.Next() {
		var summary ProjectSummary
		if err := rows.Scan(&summary.Project, &summary.TaskCount, &summary.DoneCount); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, rows.Err()
}

func RenameProjectAndTask(ctx context.Context, db *sql.DB, projectID int64, projectName string, taskID int64, taskTitle string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := updateExactlyOne(ctx, tx, "UPDATE projects SET name = ? WHERE id = ?", projectName, projectID); err != nil {
		return err
	}
	if err := updateExactlyOne(ctx, tx, "UPDATE tasks SET title = ? WHERE id = ? AND project_id = ?", taskTitle, taskID, projectID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO task_audit(task_id, action) VALUES (?, ?)", taskID, "renamed"); err != nil {
		return err
	}
	return tx.Commit()
}

func updateExactlyOne(ctx context.Context, tx *sql.Tx, query string, args ...any) error {
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrNotFound
	}
	return nil
}

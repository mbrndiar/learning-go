// Command 05_repository_pattern keeps SQL behind a narrow domain interface.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID    int64
	Title string
	Done  bool
}

type TaskRepository interface {
	Create(context.Context, string) (Task, error)
	Find(context.Context, int64) (Task, error)
	List(context.Context, bool) ([]Task, error)
}

type SQLTaskRepository struct{ db *sql.DB }

func NewSQLTaskRepository(db *sql.DB) *SQLTaskRepository {
	return &SQLTaskRepository{db: db}
}

func (r *SQLTaskRepository) Create(ctx context.Context, title string) (Task, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO tasks(title) VALUES (?)", title)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("created task id: %w", err)
	}
	return r.Find(ctx, id)
}

func (r *SQLTaskRepository) Find(ctx context.Context, id int64) (Task, error) {
	var task Task
	err := r.db.QueryRowContext(ctx,
		"SELECT id, title, done FROM tasks WHERE id = ?", id,
	).Scan(&task.ID, &task.Title, &task.Done)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, fmt.Errorf("task %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return Task{}, fmt.Errorf("find task %d: %w", id, err)
	}
	return task, nil
}

func (r *SQLTaskRepository) List(ctx context.Context, done bool) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, title, done FROM tasks WHERE done = ? ORDER BY id", done)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Done); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}

func openDatabase(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, "CREATE TABLE tasks (id INTEGER PRIMARY KEY, title TEXT NOT NULL, done BOOLEAN NOT NULL DEFAULT FALSE)"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func pendingTitles(ctx context.Context, repository TaskRepository) ([]string, error) {
	tasks, err := repository.List(ctx, false)
	if err != nil {
		return nil, err
	}
	titles := make([]string, len(tasks))
	for i, task := range tasks {
		titles[i] = task.Title
	}
	return titles, nil
}

func main() {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var repository TaskRepository = NewSQLTaskRepository(db)
	if _, err := repository.Create(ctx, "hide SQL behind an interface"); err != nil {
		log.Fatal(err)
	}
	titles, err := pendingTitles(ctx, repository)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(titles)
}

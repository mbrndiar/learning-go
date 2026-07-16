// Package sqlsqlite contains focused database/sql exercises backed by real SQLite.
package sqlsqlite

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

// TODO: create projects, tasks, and task_audit tables with keys and constraints,
// plus an index supporting task filtering by project and done state.
func CreateSchema(context.Context, *sql.DB) error {
	return errors.New("TODO: implement CreateSchema")
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// TODO: insert using parameters, read LastInsertId, and return the mapped row.
func (r *SQLiteRepository) Create(context.Context, Task) (Task, error) {
	return Task{}, errors.New("TODO: implement Create")
}

// TODO: map one row and translate sql.ErrNoRows to ErrNotFound.
func (r *SQLiteRepository) Get(context.Context, int64) (Task, error) {
	return Task{}, errors.New("TODO: implement Get")
}

// TODO: use fixed parameterized queries for optional done filtering and always
// order by ID. Do not build SQL from caller-provided text.
func (r *SQLiteRepository) List(context.Context, *bool) ([]Task, error) {
	return nil, errors.New("TODO: implement List")
}

// TODO: create a project with a parameterized INSERT.
func CreateProject(context.Context, *sql.DB, string) (int64, error) {
	return 0, errors.New("TODO: implement CreateProject")
}

// TODO: write a LEFT JOIN/GROUP BY aggregate that includes empty projects.
func ProjectSummaries(context.Context, *sql.DB) ([]ProjectSummary, error) {
	return nil, errors.New("TODO: implement ProjectSummaries")
}

// RenameProjectAndTask must update both rows and append one audit row atomically.
// If any row is missing or any statement fails, roll the whole transaction back.
func RenameProjectAndTask(context.Context, *sql.DB, int64, string, int64, string) error {
	return fmt.Errorf("TODO: implement RenameProjectAndTask")
}

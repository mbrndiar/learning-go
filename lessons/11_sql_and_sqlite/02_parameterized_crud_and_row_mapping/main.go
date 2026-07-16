// Command 02_parameterized_crud_and_row_mapping demonstrates safe values,
// CRUD statements, nullable columns, and QueryRow/Rows scanning.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID          int64
	Title       string
	Done        bool
	CompletedAt *time.Time
}

func openDatabase(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(ctx, `CREATE TABLE tasks (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE,
		completed_at DATETIME
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func createTask(ctx context.Context, db *sql.DB, title string) (Task, error) {
	// Values are separate arguments. Never concatenate user input into SQL.
	result, err := db.ExecContext(ctx, "INSERT INTO tasks(title) VALUES (?)", title)
	if err != nil {
		return Task{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, err
	}
	return getTask(ctx, db, id)
}

func getTask(ctx context.Context, db *sql.DB, id int64) (Task, error) {
	var task Task
	var completed sql.NullTime
	err := db.QueryRowContext(ctx,
		"SELECT id, title, done, completed_at FROM tasks WHERE id = ?", id,
	).Scan(&task.ID, &task.Title, &task.Done, &completed)
	if err != nil {
		return Task{}, err
	}
	if completed.Valid {
		task.CompletedAt = &completed.Time
	}
	return task, nil
}

func completeTask(ctx context.Context, db *sql.DB, id int64, at time.Time) error {
	result, err := db.ExecContext(ctx,
		"UPDATE tasks SET done = ?, completed_at = ? WHERE id = ?", true, at, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func deleteTask(ctx context.Context, db *sql.DB, id int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	return err
}

func main() {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	task, err := createTask(ctx, db, "parameter text: ' OR 1=1 --")
	if err != nil {
		log.Fatal(err)
	}
	if err := completeTask(ctx, db, task.ID, time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)); err != nil {
		log.Fatal(err)
	}
	task, err = getTask(ctx, db, task.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", task)
	fmt.Println("missing row:", errors.Is(getOnlyError(ctx, db, 999), sql.ErrNoRows))
}

func getOnlyError(ctx context.Context, db *sql.DB, id int64) error {
	_, err := getTask(ctx, db, id)
	return err
}

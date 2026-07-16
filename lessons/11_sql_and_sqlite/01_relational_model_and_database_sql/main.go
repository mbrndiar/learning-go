// Command 01_relational_model_and_database_sql introduces tables, keys,
// constraints, and the database/sql connection-pool API using real SQLite.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE
);`

type Task struct {
	ID      int64
	Project string
	Title   string
	Done    bool
}

func openDatabase(ctx context.Context) (*sql.DB, error) {
	// :memory: is SQLite-specific. Restricting the pool to one connection keeps
	// every database/sql operation on the same in-memory database.
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func seed(ctx context.Context, db *sql.DB) error {
	result, err := db.ExecContext(ctx, "INSERT INTO projects(name) VALUES (?)", "course")
	if err != nil {
		return err
	}
	projectID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx,
		"INSERT INTO tasks(project_id, title) VALUES (?, ?), (?, ?)",
		projectID, "learn relations", projectID, "practice SQL")
	return err
}

func listTasks(ctx context.Context, db *sql.DB) ([]Task, error) {
	rows, err := db.QueryContext(ctx, `
SELECT tasks.id, projects.name, tasks.title, tasks.done
FROM tasks JOIN projects ON projects.id = tasks.project_id
ORDER BY tasks.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Project, &task.Title, &task.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func main() {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := seed(ctx, db); err != nil {
		log.Fatal(err)
	}
	tasks, err := listTasks(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	for _, task := range tasks {
		fmt.Printf("%d %s: %s (done=%t)\n", task.ID, task.Project, task.Title, task.Done)
	}
}

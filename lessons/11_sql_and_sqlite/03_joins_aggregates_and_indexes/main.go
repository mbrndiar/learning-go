// Command 03_joins_aggregates_and_indexes combines related tables, groups
// rows, and inspects a deliberately created index.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type ProjectSummary struct {
	Name      string
	TaskCount int
	DoneCount int
}

func openDatabase(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.ExecContext(ctx, `
CREATE TABLE projects (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE);
CREATE TABLE tasks (id INTEGER PRIMARY KEY, project_id INTEGER NOT NULL, title TEXT NOT NULL, done BOOLEAN NOT NULL);
CREATE INDEX idx_tasks_project_done ON tasks(project_id, done);
INSERT INTO projects(id, name) VALUES (1, 'course'), (2, 'website');
INSERT INTO tasks(project_id, title, done) VALUES
 (1, 'write', TRUE), (1, 'review', FALSE), (2, 'deploy', TRUE);`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func summaries(ctx context.Context, db *sql.DB) ([]ProjectSummary, error) {
	rows, err := db.QueryContext(ctx, `
SELECT p.name, COUNT(t.id), COALESCE(SUM(CASE WHEN t.done THEN 1 ELSE 0 END), 0)
FROM projects AS p
LEFT JOIN tasks AS t ON t.project_id = p.id
GROUP BY p.id, p.name
ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ProjectSummary
	for rows.Next() {
		var summary ProjectSummary
		if err := rows.Scan(&summary.Name, &summary.TaskCount, &summary.DoneCount); err != nil {
			return nil, err
		}
		result = append(result, summary)
	}
	return result, rows.Err()
}

func hasIndex(ctx context.Context, db *sql.DB, name string) (bool, error) {
	// sqlite_master is SQLite-specific metadata. Other databases expose indexes differently.
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?", name,
	).Scan(&count)
	return count == 1, err
}

func main() {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	items, err := summaries(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		fmt.Printf("%s: %d tasks, %d done\n", item.Name, item.TaskCount, item.DoneCount)
	}
}

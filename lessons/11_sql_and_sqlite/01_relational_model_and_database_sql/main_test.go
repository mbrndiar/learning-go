package main

import (
	"context"
	"testing"
)

func TestSchemaRelationsAndQuery(t *testing.T) {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := seed(ctx, db); err != nil {
		t.Fatal(err)
	}
	tasks, err := listTasks(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 || tasks[0].Project != "course" {
		t.Fatalf("tasks = %+v", tasks)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO tasks(project_id, title) VALUES (?, ?)", 999, "orphan"); err == nil {
		t.Fatal("foreign-key constraint accepted an orphan task")
	}
}

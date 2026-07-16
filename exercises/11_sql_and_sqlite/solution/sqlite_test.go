package solution

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func newTestRepository(t *testing.T) (*sql.DB, TaskRepository) {
	t.Helper()
	ctx := context.Background()
	db, err := OpenDatabase(ctx, filepath.Join(t.TempDir(), "exercise.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := CreateSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	return db, NewSQLiteRepository(db)
}

func TestSchemaParameterizedCRUDAndRowMapping(t *testing.T) {
	db, repository := newTestRepository(t)
	ctx := context.Background()
	projectID, err := CreateProject(ctx, db, "Robert'); DROP TABLE tasks;--")
	if err != nil {
		t.Fatal(err)
	}
	created, err := repository.Create(ctx, Task{ProjectID: projectID, Title: "map this row"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != created || got.Title != "map this row" {
		t.Fatalf("Get = %+v, want %+v", got, created)
	}
	if _, err := repository.Get(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing error = %v, want ErrNotFound", err)
	}
}

func TestListFiltersAndOrders(t *testing.T) {
	db, repository := newTestRepository(t)
	ctx := context.Background()
	projectID, err := CreateProject(ctx, db, "course")
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range []Task{
		{ProjectID: projectID, Title: "first", Done: true},
		{ProjectID: projectID, Title: "second", Done: false},
		{ProjectID: projectID, Title: "third", Done: true},
	} {
		if _, err := repository.Create(ctx, task); err != nil {
			t.Fatal(err)
		}
	}
	done := true
	tasks, err := repository.List(ctx, &done)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 || tasks[0].Title != "first" || tasks[1].Title != "third" {
		t.Fatalf("filtered tasks = %+v", tasks)
	}
	all, err := repository.List(ctx, nil)
	if err != nil || len(all) != 3 || all[0].ID >= all[1].ID {
		t.Fatalf("all tasks = %+v, %v", all, err)
	}
}

func TestJoinAggregateIncludesEmptyProjects(t *testing.T) {
	db, repository := newTestRepository(t)
	ctx := context.Background()
	courseID, _ := CreateProject(ctx, db, "course")
	_, _ = CreateProject(ctx, db, "empty")
	_, _ = repository.Create(ctx, Task{ProjectID: courseID, Title: "done", Done: true})
	_, _ = repository.Create(ctx, Task{ProjectID: courseID, Title: "pending"})
	summaries, err := ProjectSummaries(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 2 || summaries[0] != (ProjectSummary{"course", 2, 1}) || summaries[1] != (ProjectSummary{"empty", 0, 0}) {
		t.Fatalf("summaries = %+v", summaries)
	}
}

func TestTransactionRollsBackEveryStep(t *testing.T) {
	db, repository := newTestRepository(t)
	ctx := context.Background()
	projectID, _ := CreateProject(ctx, db, "original project")
	task, _ := repository.Create(ctx, Task{ProjectID: projectID, Title: "original task"})
	if err := RenameProjectAndTask(ctx, db, projectID, "changed project", task.ID+999, "changed task"); err == nil {
		t.Fatal("RenameProjectAndTask succeeded for a missing task")
	}
	var name string
	if err := db.QueryRowContext(ctx, "SELECT name FROM projects WHERE id = ?", projectID).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "original project" {
		t.Fatalf("project name = %q; first update was not rolled back", name)
	}
	var auditCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM task_audit").Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 0 {
		t.Fatalf("audit count = %d after rollback", auditCount)
	}
}

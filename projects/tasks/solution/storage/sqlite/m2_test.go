package sqlite_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m2"
	_ "modernc.org/sqlite"
)

func TestRepositoryContract(t *testing.T) {
	m2.Run(t, ".db", func(path string) (task.Repository, func() error, error) {
		repository, err := sqlite.Open(path)
		if err != nil {
			return nil, nil, err
		}
		return repository, repository.Close, nil
	})
}

func TestIncompatibleSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE tasks (id INTEGER PRIMARY KEY, title TEXT)`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	repository, err := sqlite.Open(path)
	if repository != nil {
		repository.Close()
		t.Fatal("Open returned a repository for an incompatible schema")
	}
	if !errors.Is(err, task.ErrStorage) {
		t.Fatalf("Open error = %v; want ErrStorage", err)
	}
}

func TestClosedRepositoryWrapsStorageFailure(t *testing.T) {
	repository, err := sqlite.Open(filepath.Join(t.TempDir(), "tasks.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Close(); err != nil {
		t.Fatal(err)
	}
	if err := repository.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	_, err = repository.List(context.Background(), task.ListFilter{})
	if !errors.Is(err, task.ErrStorage) {
		t.Fatalf("List error = %v; want ErrStorage", err)
	}
	var storageError *task.StorageError
	if !errors.As(err, &storageError) || storageError.Operation == "" || storageError.Err == nil {
		t.Fatalf("List error = %#v; want operation and underlying cause", err)
	}
}

func TestPathCharactersAreNotInterpretedAsDSNOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks #1?.db")
	repository, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	created, err := repository.Create(context.Background(), task.CreateInput{Title: "safe path"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 1 {
		t.Fatalf("created ID = %d; want 1", created.ID)
	}
}

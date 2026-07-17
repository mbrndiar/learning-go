package sqlite_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

func TestOpenIsExplicitlyIncomplete(t *testing.T) {
	repository, err := sqlite.Open(filepath.Join(t.TempDir(), "tasks.db"))
	if repository != nil {
		t.Fatal("Open returned a repository")
	}
	if !errors.Is(err, task.ErrNotImplemented) {
		t.Fatalf("Open error = %v; want ErrNotImplemented", err)
	}
}

func TestOpenContextIsExplicitlyIncomplete(t *testing.T) {
	repository, err := sqlite.OpenContext(context.Background(), filepath.Join(t.TempDir(), "tasks.db"))
	if repository != nil {
		t.Fatal("OpenContext returned a repository")
	}
	if !errors.Is(err, task.ErrNotImplemented) {
		t.Fatalf("OpenContext error = %v; want ErrNotImplemented", err)
	}
}

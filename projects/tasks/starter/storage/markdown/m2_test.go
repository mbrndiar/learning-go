package markdown_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

func TestOpenIsExplicitlyIncomplete(t *testing.T) {
	repository, err := markdown.Open(filepath.Join(t.TempDir(), "tasks.md"))
	if repository != nil {
		t.Fatal("Open returned a repository")
	}
	if !errors.Is(err, task.ErrNotImplemented) {
		t.Fatalf("Open error = %v; want ErrNotImplemented", err)
	}
}

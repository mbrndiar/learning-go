package markdown_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m2"
)

func TestRepositoryContract(t *testing.T) {
	m2.Run(t, ".md", func(path string) (task.Repository, func() error, error) {
		repository, err := markdown.Open(path)
		if err != nil {
			return nil, nil, err
		}
		return repository, func() error { return nil }, nil
	})
}

func TestOpenContextPreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	path := filepath.Join(t.TempDir(), "tasks.md")
	repository, err := markdown.OpenContext(ctx, path)
	if repository != nil {
		t.Fatal("OpenContext returned a repository for a pre-canceled context")
	}
	if !errors.Is(err, task.ErrStorage) {
		t.Fatalf("OpenContext error = %v; want ErrStorage", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("OpenContext error = %v; want context.Canceled", err)
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("OpenContext created %s on a pre-canceled context", path)
	}
}

func TestMissingFileInitializesExactDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.md")
	if _, err := markdown.Open(path); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "<!-- rest-task-api:v1 next-id=1 -->\n# Tasks\n\n"
	if string(content) != want {
		t.Fatalf("initialized content = %q; want %q", content, want)
	}
}

func TestSerializationIsDeterministic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.md")
	repository, err := markdown.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	first, err := repository.Create(context.Background(), task.CreateInput{Title: "literal *Markdown*"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.Create(context.Background(), task.CreateInput{Title: "second"})
	if err != nil {
		t.Fatal(err)
	}
	completed := true
	if _, err := repository.Update(context.Background(), second.ID, task.UpdateInput{Completed: &completed}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "<!-- rest-task-api:v1 next-id=3 -->\n# Tasks\n\n" +
		"- [ ] 1: literal *Markdown*\n" +
		"- [x] 2: second\n"
	if string(content) != want {
		t.Fatalf("content = %q; want %q", content, want)
	}
	if first.ID != 1 {
		t.Fatalf("first ID = %d; want 1", first.ID)
	}
}

func TestMalformedDocuments(t *testing.T) {
	cases := map[string]string{
		"empty":                 "",
		"invalid UTF-8":         string([]byte{0xff, '\n'}),
		"missing final newline": "<!-- rest-task-api:v1 next-id=1 -->\n# Tasks\n",
		"extra final newline":   "<!-- rest-task-api:v1 next-id=1 -->\n# Tasks\n\n\n",
		"missing metadata":      "# Tasks\n\n",
		"unsupported version":   "<!-- rest-task-api:v2 next-id=1 -->\n# Tasks\n\n",
		"noncanonical version":  "<!-- rest-task-api:v01 next-id=1 -->\n# Tasks\n\n",
		"zero next ID":          "<!-- rest-task-api:v1 next-id=0 -->\n# Tasks\n\n",
		"invalid heading":       "<!-- rest-task-api:v1 next-id=1 -->\n# Task\n\n",
		"malformed row":         "<!-- rest-task-api:v1 next-id=2 -->\n# Tasks\n\n- [X] 1: title\n",
		"zero ID":               "<!-- rest-task-api:v1 next-id=2 -->\n# Tasks\n\n- [ ] 0: title\n",
		"duplicate ID":          "<!-- rest-task-api:v1 next-id=3 -->\n# Tasks\n\n- [ ] 1: one\n- [x] 1: two\n",
		"out of order":          "<!-- rest-task-api:v1 next-id=3 -->\n# Tasks\n\n- [ ] 2: two\n- [x] 1: one\n",
		"invalid title":         "<!-- rest-task-api:v1 next-id=2 -->\n# Tasks\n\n- [ ] 1: trailing \n",
		"next ID not greater":   "<!-- rest-task-api:v1 next-id=2 -->\n# Tasks\n\n- [ ] 2: title\n",
		"unexpected blank row":  "<!-- rest-task-api:v1 next-id=2 -->\n# Tasks\n\n- [ ] 1: title\n\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tasks.md")
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			repository, err := markdown.Open(path)
			if repository != nil {
				t.Fatal("Open returned repository for malformed document")
			}
			if !errors.Is(err, task.ErrStorage) {
				t.Fatalf("Open error = %v; want ErrStorage", err)
			}
			var storageError *task.StorageError
			if !errors.As(err, &storageError) || storageError.Operation == "" || storageError.Err == nil {
				t.Fatalf("Open error = %#v; want operation and underlying cause", err)
			}
		})
	}
}

func TestStorageFailureLeavesNoTempArtifacts(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "missing", "tasks.md")
	repository, err := markdown.Open(path)
	if repository != nil {
		t.Fatal("Open unexpectedly returned a repository")
	}
	if !errors.Is(err, task.ErrStorage) {
		t.Fatalf("Open error = %v; want ErrStorage", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			t.Fatalf("temporary artifact remains: %s", entry.Name())
		}
	}
}

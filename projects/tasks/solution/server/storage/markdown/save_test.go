package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveRemovesTemporaryFileWhenRenameFails(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "tasks.md")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}

	repository := &Repository{path: target}
	if err := repository.save(document{NextID: 1}); err == nil {
		t.Fatal("save succeeded with a directory as its target")
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

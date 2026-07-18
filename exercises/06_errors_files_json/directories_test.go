package contacts

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEnsureWorkspace(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	if err := EnsureWorkspace(root); err != nil {
		t.Fatalf("EnsureWorkspace() error = %v", err)
	}

	for _, relative := range []string{"inbox", "archive", filepath.Join("reports", "daily")} {
		info, err := os.Stat(filepath.Join(root, relative))
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", relative, err)
		}
		if !info.IsDir() {
			t.Fatalf("%q is not a directory", relative)
		}
	}
}

func TestListRegularFiles(t *testing.T) {
	root := t.TempDir()
	for path, content := range map[string]string{
		filepath.Join("inbox", "beta.txt"):           "beta",
		filepath.Join("inbox", "alpha.txt"):          "alpha",
		filepath.Join("reports", "daily", "one.log"): "one",
	} {
		fullPath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", path, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatalf("MkdirAll(empty) error = %v", err)
	}

	got, err := ListRegularFiles(root)
	if err != nil {
		t.Fatalf("ListRegularFiles() error = %v", err)
	}
	want := []string{
		"inbox/alpha.txt",
		"inbox/beta.txt",
		"reports/daily/one.log",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListRegularFiles() = %v, want %v", got, want)
	}
}

func TestListRegularFilesErrors(t *testing.T) {
	t.Run("missing root", func(t *testing.T) {
		_, err := ListRegularFiles(filepath.Join(t.TempDir(), "missing"))
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("ListRegularFiles() error = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("root is a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		_, err := ListRegularFiles(path)
		if !errors.Is(err, ErrNotDirectory) {
			t.Fatalf("ListRegularFiles() error = %v, want ErrNotDirectory", err)
		}
	})
}

func TestMoveFile(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "inbox", "report.txt")
	destination := filepath.Join(root, "archive", "2026", "report.txt")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatalf("MkdirAll(source) error = %v", err)
	}
	if err := os.WriteFile(source, []byte("report"), 0o600); err != nil {
		t.Fatalf("WriteFile(source) error = %v", err)
	}

	if err := MoveFile(source, destination); err != nil {
		t.Fatalf("MoveFile() error = %v", err)
	}
	if _, err := os.Stat(source); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("source Stat error = %v, want fs.ErrNotExist", err)
	}
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("ReadFile(destination) error = %v", err)
	}
	if string(content) != "report" {
		t.Fatalf("destination content = %q, want %q", content, "report")
	}
}

func TestMoveFileMissingSource(t *testing.T) {
	root := t.TempDir()
	err := MoveFile(
		filepath.Join(root, "missing.txt"),
		filepath.Join(root, "archive", "missing.txt"),
	)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("MoveFile() error = %v, want fs.ErrNotExist", err)
	}
}

func TestRemoveEmptyDirectory(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatalf("Mkdir() error = %v", err)
		}
		if err := RemoveEmptyDirectory(path); err != nil {
			t.Fatalf("RemoveEmptyDirectory() error = %v", err)
		}
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("Stat() error = %v, want fs.ErrNotExist", err)
		}
	})

	t.Run("non-empty directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "non-empty")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatalf("Mkdir() error = %v", err)
		}
		if err := os.WriteFile(filepath.Join(path, "keep.txt"), []byte("keep"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		if err := RemoveEmptyDirectory(path); err == nil {
			t.Fatal("RemoveEmptyDirectory() error = nil, want non-empty directory error")
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("non-empty directory was removed: %v", err)
		}
	})

	t.Run("path is a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		err := RemoveEmptyDirectory(path)
		if !errors.Is(err, ErrNotDirectory) {
			t.Fatalf("RemoveEmptyDirectory() error = %v, want ErrNotDirectory", err)
		}
	})
}

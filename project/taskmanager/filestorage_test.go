package taskmanager

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func newFileStorage(t *testing.T) Storage {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	storage, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	return storage
}

func TestFileStorageContract(t *testing.T) {
	runStorageContract(t, newFileStorage)
}

func TestNewFileStorageRejectsEmptyPath(t *testing.T) {
	if _, err := NewFileStorage(""); err == nil {
		t.Fatal("NewFileStorage(\"\") error = nil, want error")
	}
}

func TestFileStorageMissingFileIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	storage, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	tasks, err := storage.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("List() = %d, want 0", len(tasks))
	}
}

func TestFileStorageWritesSchemaDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	storage, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	if _, err := storage.Add(context.Background(), "write me"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if doc.Version != currentSchemaVersion {
		t.Fatalf("version = %d, want %d", doc.Version, currentSchemaVersion)
	}
	if doc.NextID != 2 {
		t.Fatalf("next_id = %d, want 2", doc.NextID)
	}
	if len(doc.Tasks) != 1 || doc.Tasks[0].ID != 1 {
		t.Fatalf("tasks = %+v, want single task id=1", doc.Tasks)
	}
}

func TestFileStorageMonotonicIDsAcrossReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	ctx := context.Background()

	first, err := mustFileStorage(t, path).Add(ctx, "one")
	if err != nil {
		t.Fatalf("Add(one) error = %v", err)
	}
	if err := mustFileStorage(t, path).Remove(ctx, first.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// A fresh FileStorage over the same file must continue the id sequence,
	// proving next_id is persisted rather than derived from current tasks.
	next, err := mustFileStorage(t, path).Add(ctx, "two")
	if err != nil {
		t.Fatalf("Add(two) error = %v", err)
	}
	if next.ID <= first.ID {
		t.Fatalf("reloaded Add id = %d, want > %d", next.ID, first.ID)
	}
}

func TestFileStorageReadsLegacyArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	legacy := `[{"id":1,"title":"legacy one","done":false},{"id":2,"title":"legacy two","done":true}]`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	storage := mustFileStorage(t, path)
	ctx := context.Background()

	tasks, err := storage.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("List() = %d, want 2", len(tasks))
	}

	// The next id must exceed the largest legacy id and the file must migrate
	// to the object schema on the next write.
	added, err := storage.Add(ctx, "migrated")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if added.ID != 3 {
		t.Fatalf("Add() id = %d, want 3", added.ID)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "\"next_id\"") {
		t.Fatalf("migrated file missing next_id: %s", data)
	}
}

func TestFileStorageRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{"invalid json", []byte("{not json}")},
		{"unknown field", []byte(`{"version":1,"next_id":1,"tasks":[],"extra":true}`)},
		{"duplicate ids", []byte(`{"version":1,"next_id":3,"tasks":[{"id":1,"title":"a"},{"id":1,"title":"b"}]}`)},
		{"invalid task id", []byte(`{"version":1,"next_id":3,"tasks":[{"id":0,"title":"a"}]}`)},
		{"empty task title", []byte(`{"version":1,"next_id":3,"tasks":[{"id":1,"title":"  "}]}`)},
		{"future schema", []byte(`{"version":999,"next_id":1,"tasks":[]}`)},
		{"invalid utf8", []byte("{\"version\":1,\"next_id\":1,\"tasks\":[{\"id\":1,\"title\":\"\xff\xfe\"}]}")},
		{"trailing data", []byte(`{"version":1,"next_id":1,"tasks":[]}{}`)},
		{"malformed trailing data", []byte(`{"version":1,"next_id":1,"tasks":[]}x`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tasks.json")
			if err := os.WriteFile(path, test.content, 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			storage := mustFileStorage(t, path)
			if _, err := storage.List(context.Background()); err == nil {
				t.Fatal("List() error = nil, want error for invalid data")
			}
		})
	}
}

func TestFileStorageAtomicSaveLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	storage := mustFileStorage(t, path)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if _, err := storage.Add(ctx, "task"); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("temp file left behind: %s", entry.Name())
		}
	}
	if len(entries) != 1 {
		t.Fatalf("directory has %d entries, want 1 (tasks.json only)", len(entries))
	}
}

func TestFileStorageAtomicSaveReplacesContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	storage := mustFileStorage(t, path)
	ctx := context.Background()

	added, err := storage.Add(ctx, "original")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if _, err := storage.Complete(ctx, added.ID); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	// Read back through a fresh instance to confirm the rename published the
	// completed state, not a partial write.
	got, err := mustFileStorage(t, path).Get(ctx, added.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !got.Done {
		t.Fatalf("persisted done = false, want true")
	}
}

func TestFileStorageConcurrentAddsAreSerialized(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	storage := mustFileStorage(t, path)
	ctx := context.Background()

	const workers = 10
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := storage.Add(ctx, "concurrent"); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent Add() error = %v", err)
	}

	tasks, err := storage.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != workers {
		t.Fatalf("List() = %d tasks, want %d", len(tasks), workers)
	}

	ids := make(map[int64]struct{}, len(tasks))
	for _, task := range tasks {
		if _, dup := ids[task.ID]; dup {
			t.Fatalf("duplicate id %d after concurrent adds", task.ID)
		}
		ids[task.ID] = struct{}{}
	}
}

func TestFileStorageContextCancellation(t *testing.T) {
	storage := mustFileStorage(t, filepath.Join(t.TempDir(), "tasks.json"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := storage.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("List(cancelled) error = %v, want context.Canceled", err)
	}
	if _, err := storage.Add(ctx, "task"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Add(cancelled) error = %v, want context.Canceled", err)
	}
}

func mustFileStorage(t *testing.T, path string) *FileStorage {
	t.Helper()
	storage, err := NewFileStorage(path)
	if err != nil {
		t.Fatalf("NewFileStorage() error = %v", err)
	}
	return storage
}

package taskapi

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func newMemoryStore(t *testing.T) *SQLiteStore {
	t.Helper()
	store, err := OpenSQLiteStore(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return store
}

func TestOpenSQLiteStoreRejectsEmptyDSN(t *testing.T) {
	if _, err := OpenSQLiteStore(context.Background(), "  "); err == nil {
		t.Fatal("OpenSQLiteStore(\"\") error = nil, want error")
	}
}

func TestSQLiteStoreCRUD(t *testing.T) {
	store := newMemoryStore(t)
	ctx := context.Background()

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("List() = %d, want 0", len(tasks))
	}

	added, err := store.Add(ctx, "  write tests  ")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if added.ID <= 0 {
		t.Fatalf("Add() id = %d, want positive", added.ID)
	}
	if added.Title != "write tests" {
		t.Fatalf("Add() title = %q, want trimmed %q", added.Title, "write tests")
	}
	if added.Done {
		t.Fatal("Add() done = true, want false")
	}

	got, err := store.Get(ctx, added.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != added {
		t.Fatalf("Get() = %+v, want %+v", got, added)
	}

	completed, err := store.Complete(ctx, added.ID)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if !completed.Done {
		t.Fatal("Complete() done = false, want true")
	}

	if err := store.Remove(ctx, added.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := store.Get(ctx, added.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() after remove error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteStoreNotFound(t *testing.T) {
	store := newMemoryStore(t)
	ctx := context.Background()

	if _, err := store.Get(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
	if _, err := store.Complete(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Complete(missing) error = %v, want ErrNotFound", err)
	}
	if err := store.Remove(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Remove(missing) error = %v, want ErrNotFound", err)
	}
}

func TestSQLiteStoreRejectsInvalidTitles(t *testing.T) {
	store := newMemoryStore(t)
	ctx := context.Background()

	if _, err := store.Add(ctx, "   "); !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("Add(blank) error = %v, want ErrEmptyTitle", err)
	}
	if _, err := store.Add(ctx, strings.Repeat("x", MaxTitleLength+1)); !errors.Is(err, ErrTitleTooLong) {
		t.Fatalf("Add(long) error = %v, want ErrTitleTooLong", err)
	}
	if _, err := store.Add(ctx, strings.Repeat("🐹", MaxTitleLength)); err != nil {
		t.Fatalf("Add(max-length Unicode title) error = %v", err)
	}
	if _, err := store.Add(ctx, strings.Repeat("🐹", MaxTitleLength+1)); !errors.Is(err, ErrTitleTooLong) {
		t.Fatalf("Add(long Unicode title) error = %v, want ErrTitleTooLong", err)
	}
	if _, err := store.Add(ctx, "\xff\xfe"); !errors.Is(err, ErrInvalidTitle) {
		t.Fatalf("Add(invalid UTF-8) error = %v, want ErrInvalidTitle", err)
	}
}

func TestSQLiteStoreMonotonicIDs(t *testing.T) {
	store := newMemoryStore(t)
	ctx := context.Background()

	first, err := store.Add(ctx, "one")
	if err != nil {
		t.Fatalf("Add(one) error = %v", err)
	}
	second, err := store.Add(ctx, "two")
	if err != nil {
		t.Fatalf("Add(two) error = %v", err)
	}
	if err := store.Remove(ctx, second.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	third, err := store.Add(ctx, "three")
	if err != nil {
		t.Fatalf("Add(three) error = %v", err)
	}
	if third.ID == first.ID || third.ID == second.ID {
		t.Fatalf("Add(three) id = %d reused earlier id (first=%d second=%d)", third.ID, first.ID, second.ID)
	}
}

func TestSQLiteStorePersistsToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.db")
	ctx := context.Background()

	first, err := OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}
	added, err := first.Add(ctx, "durable")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	second, err := OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("reopen OpenSQLiteStore() error = %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })

	got, err := second.Get(ctx, added.ID)
	if err != nil {
		t.Fatalf("Get() after reopen error = %v", err)
	}
	if got.Title != "durable" {
		t.Fatalf("Get() title = %q, want %q", got.Title, "durable")
	}
}

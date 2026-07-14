package taskmanager

import (
	"context"
	"errors"
	"testing"
)

// storageContract exercises the behavior every Storage implementation must
// provide. It is applied to FileStorage and RESTStorage so both backends are
// held to the same guarantees, including monotonic identifiers and uniform
// not-found errors.
func runStorageContract(t *testing.T, factory func(t *testing.T) Storage) {
	t.Helper()

	t.Run("empty store lists nothing", func(t *testing.T) {
		storage := factory(t)
		tasks, err := storage.List(context.Background())
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(tasks) != 0 {
			t.Fatalf("List() = %d tasks, want 0", len(tasks))
		}
	})

	t.Run("add assigns positive increasing ids", func(t *testing.T) {
		storage := factory(t)
		ctx := context.Background()

		first, err := storage.Add(ctx, "first")
		if err != nil {
			t.Fatalf("Add(first) error = %v", err)
		}
		if first.ID <= 0 {
			t.Fatalf("Add(first) id = %d, want positive", first.ID)
		}
		if first.Title != "first" || first.Done {
			t.Fatalf("Add(first) = %+v, want title=first done=false", first)
		}

		second, err := storage.Add(ctx, "second")
		if err != nil {
			t.Fatalf("Add(second) error = %v", err)
		}
		if second.ID <= first.ID {
			t.Fatalf("Add(second) id = %d, want > %d", second.ID, first.ID)
		}
	})

	t.Run("get returns stored task and not-found", func(t *testing.T) {
		storage := factory(t)
		ctx := context.Background()

		added, err := storage.Add(ctx, "lookup")
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		got, err := storage.Get(ctx, added.ID)
		if err != nil {
			t.Fatalf("Get(%d) error = %v", added.ID, err)
		}
		if got != added {
			t.Fatalf("Get(%d) = %+v, want %+v", added.ID, got, added)
		}

		if _, err := storage.Get(ctx, added.ID+1000); !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("Get(missing) error = %v, want ErrTaskNotFound", err)
		}
	})

	t.Run("complete marks done and reports not-found", func(t *testing.T) {
		storage := factory(t)
		ctx := context.Background()

		added, err := storage.Add(ctx, "complete me")
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		completed, err := storage.Complete(ctx, added.ID)
		if err != nil {
			t.Fatalf("Complete(%d) error = %v", added.ID, err)
		}
		if !completed.Done {
			t.Fatalf("Complete(%d) done = false, want true", added.ID)
		}

		got, err := storage.Get(ctx, added.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if !got.Done {
			t.Fatalf("Get() done = false after complete, want true")
		}

		if _, err := storage.Complete(ctx, added.ID+1000); !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("Complete(missing) error = %v, want ErrTaskNotFound", err)
		}
	})

	t.Run("remove deletes and reports not-found", func(t *testing.T) {
		storage := factory(t)
		ctx := context.Background()

		added, err := storage.Add(ctx, "remove me")
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		if err := storage.Remove(ctx, added.ID); err != nil {
			t.Fatalf("Remove(%d) error = %v", added.ID, err)
		}
		if _, err := storage.Get(ctx, added.ID); !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("Get() after remove error = %v, want ErrTaskNotFound", err)
		}
		if err := storage.Remove(ctx, added.ID); !errors.Is(err, ErrTaskNotFound) {
			t.Fatalf("Remove(missing) error = %v, want ErrTaskNotFound", err)
		}
	})

	t.Run("removed ids are never reused", func(t *testing.T) {
		storage := factory(t)
		ctx := context.Background()

		first, err := storage.Add(ctx, "first")
		if err != nil {
			t.Fatalf("Add(first) error = %v", err)
		}
		second, err := storage.Add(ctx, "second")
		if err != nil {
			t.Fatalf("Add(second) error = %v", err)
		}
		if err := storage.Remove(ctx, second.ID); err != nil {
			t.Fatalf("Remove(second) error = %v", err)
		}

		third, err := storage.Add(ctx, "third")
		if err != nil {
			t.Fatalf("Add(third) error = %v", err)
		}
		if third.ID == first.ID || third.ID == second.ID {
			t.Fatalf("Add(third) id = %d reused an earlier id (first=%d second=%d)", third.ID, first.ID, second.ID)
		}
	})

	t.Run("rejects empty titles", func(t *testing.T) {
		storage := factory(t)
		if _, err := storage.Add(context.Background(), "   "); err == nil {
			t.Fatal("Add(blank) error = nil, want validation error")
		}
	})
}

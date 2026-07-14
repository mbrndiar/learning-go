package main

import (
	"context"
	"errors"
	"testing"
)

func TestInsertAndGet(t *testing.T) {
	t.Parallel()

	store := newMemoryTaskStore()
	ctx := context.Background()

	created, err := store.Insert(ctx, "write tests")
	if err != nil {
		t.Fatalf("Insert() error = %v, want nil", err)
	}
	if created.ID != 1 || created.Title != "write tests" {
		t.Fatalf("Insert() = %+v, want ID=1 Title=%q", created, "write tests")
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got != created {
		t.Fatalf("Get() = %+v, want %+v", got, created)
	}
}

func TestGetMissingReturnsErrNotFound(t *testing.T) {
	t.Parallel()

	store := newMemoryTaskStore()
	_, err := store.Get(context.Background(), 42)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want it to wrap ErrNotFound", err)
	}
}

func TestListReturnsInsertionOrderByID(t *testing.T) {
	t.Parallel()

	store := newMemoryTaskStore()
	ctx := context.Background()

	store.Insert(ctx, "first")
	store.Insert(ctx, "second")
	store.Insert(ctx, "third")

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}

	want := []Task{
		{ID: 1, Title: "first"},
		{ID: 2, Title: "second"},
		{ID: 3, Title: "third"},
	}
	if len(tasks) != len(want) {
		t.Fatalf("List() = %+v, want %+v", tasks, want)
	}
	for i := range want {
		if tasks[i] != want[i] {
			t.Fatalf("List()[%d] = %+v, want %+v", i, tasks[i], want[i])
		}
	}
}

func TestOperationsRespectCanceledContext(t *testing.T) {
	t.Parallel()

	store := newMemoryTaskStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := store.Insert(ctx, "too late"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Insert() error = %v, want context.Canceled", err)
	}
	if _, err := store.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("List() error = %v, want context.Canceled", err)
	}
}

func TestMemoryTaskStoreSatisfiesTaskStore(t *testing.T) {
	t.Parallel()

	var _ TaskStore = newMemoryTaskStore()
}

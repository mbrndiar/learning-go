package main

import (
	"context"
	"errors"
	"testing"
)

func TestSQLTaskRepository(t *testing.T) {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	var repository TaskRepository = NewSQLTaskRepository(db)
	created, err := repository.Create(ctx, "repository lesson")
	if err != nil {
		t.Fatal(err)
	}
	got, err := repository.Find(ctx, created.ID)
	if err != nil || got.Title != "repository lesson" {
		t.Fatalf("Find = %+v, %v", got, err)
	}
	if _, err := repository.Find(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
	titles, err := pendingTitles(ctx, repository)
	if err != nil || len(titles) != 1 {
		t.Fatalf("pendingTitles = %v, %v", titles, err)
	}
}

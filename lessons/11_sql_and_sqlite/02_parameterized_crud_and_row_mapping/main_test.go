package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestParameterizedCRUDAndMapping(t *testing.T) {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	task, err := createTask(ctx, db, "quotes ' stay data")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2030, 2, 3, 4, 5, 6, 0, time.UTC)
	if err := completeTask(ctx, db, task.ID, at); err != nil {
		t.Fatal(err)
	}
	got, err := getTask(ctx, db, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Done || got.CompletedAt == nil || !got.CompletedAt.Equal(at) {
		t.Fatalf("task = %+v", got)
	}
	if err := deleteTask(ctx, db, task.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := getTask(ctx, db, task.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("error = %v, want sql.ErrNoRows", err)
	}
}

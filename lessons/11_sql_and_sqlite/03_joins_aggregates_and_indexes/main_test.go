package main

import (
	"context"
	"testing"
)

func TestJoinAggregateAndIndex(t *testing.T) {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	got, err := summaries(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != (ProjectSummary{Name: "course", TaskCount: 2, DoneCount: 1}) {
		t.Fatalf("summaries = %+v", got)
	}
	ok, err := hasIndex(ctx, db, "idx_tasks_project_done")
	if err != nil || !ok {
		t.Fatalf("hasIndex = %t, %v", ok, err)
	}
}

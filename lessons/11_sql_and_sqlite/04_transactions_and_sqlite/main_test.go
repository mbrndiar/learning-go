package main

import (
	"context"
	"errors"
	"testing"
)

func TestTransferCommitAndRollback(t *testing.T) {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := transfer(ctx, db, "alice", "bob", 30); err != nil {
		t.Fatal(err)
	}
	alice, _ := balance(ctx, db, "alice")
	bob, _ := balance(ctx, db, "bob")
	if alice != 70 || bob != 50 {
		t.Fatalf("balances = %d, %d", alice, bob)
	}
	if err := transfer(ctx, db, "alice", "missing", 10); err == nil {
		t.Fatal("transfer to missing account succeeded")
	}
	alice, _ = balance(ctx, db, "alice")
	if alice != 70 {
		t.Fatalf("rollback left alice balance = %d", alice)
	}
	if err := transfer(ctx, db, "alice", "bob", 1000); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("error = %v", err)
	}
}

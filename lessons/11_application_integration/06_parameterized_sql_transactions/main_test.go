package main

import (
	"context"
	"errors"
	"testing"
)

func TestTransferMovesBalance(t *testing.T) {
	t.Parallel()

	l := newLedger(map[string]int{"alice": 100, "bob": 20})

	if err := l.Transfer(context.Background(), "alice", "bob", 30); err != nil {
		t.Fatalf("Transfer() error = %v, want nil", err)
	}

	if got, want := l.Balance("alice"), 70; got != want {
		t.Fatalf("alice balance = %d, want %d", got, want)
	}
	if got, want := l.Balance("bob"), 50; got != want {
		t.Fatalf("bob balance = %d, want %d", got, want)
	}
}

func TestTransferFailsAtomicallyOnInsufficientFunds(t *testing.T) {
	t.Parallel()

	l := newLedger(map[string]int{"alice": 10, "bob": 20})

	err := l.Transfer(context.Background(), "alice", "bob", 1000)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Transfer() error = %v, want ErrInsufficientFunds", err)
	}

	// Neither balance should have moved: the failed transfer must leave
	// no partial effect, the same guarantee a rolled-back transaction
	// gives against a real database.
	if got, want := l.Balance("alice"), 10; got != want {
		t.Fatalf("alice balance after failed transfer = %d, want unchanged %d", got, want)
	}
	if got, want := l.Balance("bob"), 20; got != want {
		t.Fatalf("bob balance after failed transfer = %d, want unchanged %d", got, want)
	}
}

func TestTransferRespectsCanceledContext(t *testing.T) {
	t.Parallel()

	l := newLedger(map[string]int{"alice": 100, "bob": 0})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := l.Transfer(ctx, "alice", "bob", 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("Transfer() error = %v, want context.Canceled", err)
	}
	if got, want := l.Balance("alice"), 100; got != want {
		t.Fatalf("alice balance = %d, want unchanged %d", got, want)
	}
}

func TestNewLedgerCopiesInitialBalances(t *testing.T) {
	t.Parallel()

	initial := map[string]int{"alice": 5}
	l := newLedger(initial)

	initial["alice"] = 999 // mutating the caller's map must not affect the ledger

	if got, want := l.Balance("alice"), 5; got != want {
		t.Fatalf("Balance() = %d, want %d (ledger should own its own copy)", got, want)
	}
}

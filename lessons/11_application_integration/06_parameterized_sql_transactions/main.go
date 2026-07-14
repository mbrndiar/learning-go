// Command 06_parameterized_sql_transactions explains why SQL parameters
// prevent injection, and demonstrates the transaction pattern
// (Begin/Commit/Rollback) that keeps a multi-step change atomic. It uses
// an in-memory teaching fake shaped like database/sql so the lesson stays
// standard-library-only; the capstone connects the same pattern to a real
// database/sql handle with a SQLite driver.
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Why parameters instead of string concatenation:
//
// Never build SQL by concatenating user input:
//
//	query := "SELECT * FROM accounts WHERE name = '" + name + "'"
//
// If name is attacker-controlled and contains `' OR '1'='1`, the query's
// meaning changes completely. A real database/sql call instead uses a
// placeholder and passes the value separately, so the driver sends it as
// data, never as part of the SQL text:
//
//	db.QueryContext(ctx, "SELECT * FROM accounts WHERE name = ?", name)
//
// The placeholder syntax (`?`, `$1`, `:name`, ...) depends on the driver,
// but the principle is the same for every database/sql driver: pass
// values as arguments, never format them into the query string.

// ErrInsufficientFunds signals a transfer that would overdraw an account.
var ErrInsufficientFunds = errors.New("insufficient funds")

// ledger is a teaching fake for an accounts table, guarded by a Mutex the
// same way a real *sql.DB serializes access to shared connections.
type ledger struct {
	mu       sync.Mutex
	balances map[string]int
}

func newLedger(initial map[string]int) *ledger {
	balances := make(map[string]int, len(initial))
	for k, v := range initial {
		balances[k] = v
	}
	return &ledger{balances: balances}
}

// Balance models `SELECT balance FROM accounts WHERE name = ?`.
func (l *ledger) Balance(name string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.balances[name]
}

// Transfer models a real transaction:
//
//	tx, err := db.BeginTx(ctx, nil)
//	if err != nil { return err }
//	defer tx.Rollback() // no-op if Commit already succeeded
//
//	if _, err := tx.ExecContext(ctx,
//	    "UPDATE accounts SET balance = balance - ? WHERE name = ?", amount, from); err != nil {
//	    return err // deferred Rollback undoes nothing yet committed
//	}
//	if _, err := tx.ExecContext(ctx,
//	    "UPDATE accounts SET balance = balance + ? WHERE name = ?", amount, to); err != nil {
//	    return err
//	}
//	return tx.Commit()
//
// Both updates must succeed together, or neither should take effect: a
// crash or error between them must never leave money "created" or
// "destroyed". Transfer reproduces that all-or-nothing guarantee against
// the in-memory ledger: it validates the full operation before mutating
// any state, so a failure never leaves a partial update applied.
func (l *ledger) Transfer(ctx context.Context, from, to string, amount int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.balances[from] < amount {
		return fmt.Errorf("transfer %d from %s to %s: %w", amount, from, to, ErrInsufficientFunds)
	}

	// Only mutate state after every check has passed, so a returned error
	// above this point guarantees zero side effects, the same guarantee
	// tx.Rollback() gives you against a real database.
	l.balances[from] -= amount
	l.balances[to] += amount
	return nil
}

func main() {
	l := newLedger(map[string]int{"alice": 100, "bob": 20})

	ctx := context.Background()
	if err := l.Transfer(ctx, "alice", "bob", 30); err != nil {
		fmt.Println("transfer failed:", err)
	}
	fmt.Println("alice:", l.Balance("alice"), "bob:", l.Balance("bob"))

	err := l.Transfer(ctx, "bob", "alice", 1000)
	fmt.Println("oversized transfer error:", err)
	fmt.Println("alice unchanged:", l.Balance("alice"), "bob unchanged:", l.Balance("bob"))
}

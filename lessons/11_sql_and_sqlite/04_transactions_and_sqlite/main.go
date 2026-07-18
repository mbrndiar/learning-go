// Command 04_transactions_and_sqlite demonstrates an all-or-nothing transfer
// and calls out SQLite-specific connection settings.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

func openDatabase(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	// Foreign keys are enabled per SQLite connection; many other databases
	// enforce them without a PRAGMA.
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
	// modernc.org/sqlite accepts multiple statements in one ExecContext. This
	// convenience is driver-specific; other drivers may require separate calls.
	_, err = db.ExecContext(ctx, `
CREATE TABLE accounts (name TEXT PRIMARY KEY, balance INTEGER NOT NULL CHECK(balance >= 0));
INSERT INTO accounts(name, balance) VALUES ('alice', 100), ('bob', 20);`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func transfer(ctx context.Context, db *sql.DB, from, to string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - ? WHERE name = ? AND balance >= ?",
		amount, from, amount)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrInsufficientFunds
	}
	result, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + ? WHERE name = ?", amount, to)
	if err != nil {
		return err
	}
	changed, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return fmt.Errorf("destination account %q not found", to)
	}
	return tx.Commit()
}

func balance(ctx context.Context, db *sql.DB, name string) (int64, error) {
	var value int64
	err := db.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE name = ?", name).Scan(&value)
	return value, err
}

func main() {
	ctx := context.Background()
	db, err := openDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := transfer(ctx, db, "alice", "bob", 30); err != nil {
		log.Fatal(err)
	}
	alice, _ := balance(ctx, db, "alice")
	bob, _ := balance(ctx, db, "bob")
	fmt.Println(alice, bob)
}

// Package storage defines persistence capabilities consumed by the application.
package storage

import (
	"context"

	"github.com/mbrndiar/learning-go/capstones/comparative/starter/kvstore/domain"
)

// Store is the persistence boundary used by the comparative application.
type Store interface {
	Set(context.Context, string, domain.Value, domain.Expectation) (domain.SetResult, error)
	Get(context.Context, string) (domain.Entry, error)
	Delete(context.Context, string, domain.Expectation) (domain.DeleteResult, error)
	List(context.Context) (domain.ListResult, error)
	Close() error
}

// Opener creates a store for one literal database path.
type Opener interface {
	Open(context.Context, string) (Store, error)
}

// SQLiteOpener is the future production opener.
type SQLiteOpener struct{}

// NewSQLiteOpener constructs the SQLite opener boundary.
func NewSQLiteOpener() *SQLiteOpener {
	return &SQLiteOpener{}
}

// Open is intentionally incomplete in the harness.
func (*SQLiteOpener) Open(context.Context, string) (Store, error) {
	return nil, domain.ErrNotImplemented
}

// Package sqlite persists tasks through database/sql and SQLite. Each opened
// repository owns one process-scoped connection pool and must be closed by its
// composition root.
package sqlite

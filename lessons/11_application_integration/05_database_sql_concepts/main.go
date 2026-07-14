// Command 05_database_sql_concepts introduces the shape of database/sql
// (queries, scanning rows, sentinel not-found errors, context-aware calls)
// through a small interface and an in-memory teaching fake, so the lesson
// stays standard-library-only. The capstone project wires the same
// interface to a real database/sql connection using a SQLite driver.
package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// ErrNotFound plays the same role database/sql's sql.ErrNoRows plays for a
// real query: a single sentinel value callers can check with errors.Is,
// instead of parsing an error string.
var ErrNotFound = errors.New("record not found")

// Task is the row shape we will eventually map to and from a real table.
// Field names here mirror what a `SELECT id, title, done FROM tasks` would
// scan into.
type Task struct {
	ID    int
	Title string
	Done  bool
}

// TaskStore is the seam between business logic and storage. Keeping it
// narrow, and expressed in terms of the domain (Insert, Get, List) rather
// than raw SQL, means callers do not need to change when the backing
// implementation changes from this in-memory fake to a real
// database/sql-backed type.
type TaskStore interface {
	Insert(ctx context.Context, title string) (Task, error)
	Get(ctx context.Context, id int) (Task, error)
	List(ctx context.Context) ([]Task, error)
}

// memoryTaskStore is a teaching fake standing in for database/sql plus a
// driver. A real implementation would hold a *sql.DB connection pool and
// run parameterized SQL (see the next lesson) instead of touching a map
// directly; the method signatures here are deliberately shaped the same
// way so swapping the implementation later does not change callers.
type memoryTaskStore struct {
	mu     sync.Mutex
	nextID int
	rows   map[int]Task
}

// newMemoryTaskStore returns an empty store, analogous to a freshly
// migrated, empty database table.
func newMemoryTaskStore() *memoryTaskStore {
	return &memoryTaskStore{nextID: 1, rows: make(map[int]Task)}
}

// Insert models what `INSERT INTO tasks(title) VALUES (?)` followed by
// reading back the generated ID would do.
func (s *memoryTaskStore) Insert(ctx context.Context, title string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err // a real driver call would also fail fast on a canceled context
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task := Task{ID: s.nextID, Title: title}
	s.rows[task.ID] = task
	s.nextID++
	return task, nil
}

// Get models `SELECT id, title, done FROM tasks WHERE id = ?`, returning
// ErrNotFound instead of a zero Task when no row matches, the same way a
// real implementation would translate sql.ErrNoRows.
func (s *memoryTaskStore) Get(ctx context.Context, id int) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.rows[id]
	if !ok {
		return Task{}, fmt.Errorf("task %d: %w", id, ErrNotFound)
	}
	return task, nil
}

// List models `SELECT id, title, done FROM tasks ORDER BY id`. A real
// implementation would run rows, err := db.QueryContext(ctx, query), defer
// rows.Close(), then loop `for rows.Next() { rows.Scan(&t.ID, &t.Title,
// &t.Done) }` and finally check rows.Err() after the loop.
func (s *memoryTaskStore) List(ctx context.Context) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make([]Task, 0, len(s.rows))
	for _, t := range s.rows {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, nil
}

func main() {
	var store TaskStore = newMemoryTaskStore() // callers depend only on the interface

	ctx := context.Background()
	store.Insert(ctx, "write lesson")
	store.Insert(ctx, "review lesson")

	all, _ := store.List(ctx)
	fmt.Println("all tasks:", all)

	_, err := store.Get(ctx, 999)
	fmt.Println("missing task error:", err, "is ErrNotFound:", errors.Is(err, ErrNotFound))
}

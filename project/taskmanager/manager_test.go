package taskmanager

import (
	"context"
	"errors"
	"testing"
)

// stubStorage records the last identifier it was asked about so tests can
// assert that Manager validates before delegating.
type stubStorage struct {
	addCalls     int
	lastGetID    int
	returnErr    error
	returnedTask Task
}

func (s *stubStorage) List(context.Context) ([]Task, error) {
	return nil, s.returnErr
}

func (s *stubStorage) Get(_ context.Context, id int) (Task, error) {
	s.lastGetID = id
	return s.returnedTask, s.returnErr
}

func (s *stubStorage) Add(_ context.Context, title string) (Task, error) {
	s.addCalls++
	if s.returnErr != nil {
		return Task{}, s.returnErr
	}
	return Task{ID: 1, Title: title}, nil
}

func (s *stubStorage) Complete(_ context.Context, id int) (Task, error) {
	s.lastGetID = id
	return s.returnedTask, s.returnErr
}

func (s *stubStorage) Remove(_ context.Context, id int) error {
	s.lastGetID = id
	return s.returnErr
}

func TestNewManagerRejectsNilStorage(t *testing.T) {
	if _, err := NewManager(nil); err == nil {
		t.Fatal("NewManager(nil) error = nil, want error")
	}
}

func TestManagerAddValidatesBeforeStorage(t *testing.T) {
	stub := &stubStorage{}
	manager, err := NewManager(stub)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if _, err := manager.Add(context.Background(), "   "); !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("Add(blank) error = %v, want ErrEmptyTitle", err)
	}
	if stub.addCalls != 0 {
		t.Fatalf("storage.Add called %d times, want 0 for invalid title", stub.addCalls)
	}
}

func TestManagerAddNormalizesTitle(t *testing.T) {
	stub := &stubStorage{}
	manager, _ := NewManager(stub)

	task, err := manager.Add(context.Background(), "  trimmed  ")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if task.Title != "trimmed" {
		t.Fatalf("Add() title = %q, want %q", task.Title, "trimmed")
	}
}

func TestManagerRejectsNonPositiveIDs(t *testing.T) {
	stub := &stubStorage{}
	manager, _ := NewManager(stub)
	ctx := context.Background()

	if _, err := manager.Get(ctx, 0); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Get(0) error = %v, want ErrInvalidID", err)
	}
	if _, err := manager.Complete(ctx, -1); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Complete(-1) error = %v, want ErrInvalidID", err)
	}
	if err := manager.Remove(ctx, 0); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Remove(0) error = %v, want ErrInvalidID", err)
	}
	if stub.lastGetID != 0 {
		t.Fatalf("storage was called with id %d, want no call for invalid ids", stub.lastGetID)
	}
}

func TestManagerPropagatesStorageErrors(t *testing.T) {
	sentinel := errors.New("boom")
	stub := &stubStorage{returnErr: sentinel}
	manager, _ := NewManager(stub)

	if _, err := manager.List(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("List() error = %v, want wrapped sentinel", err)
	}
}

func TestManagerSuccessPaths(t *testing.T) {
	stub := &stubStorage{returnedTask: Task{ID: 5, Title: "x", Done: true}}
	manager, _ := NewManager(stub)
	ctx := context.Background()

	if _, err := manager.List(ctx); err != nil {
		t.Fatalf("List() error = %v", err)
	}

	got, err := manager.Get(ctx, 5)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != 5 {
		t.Fatalf("Get() id = %d, want 5", got.ID)
	}

	completed, err := manager.Complete(ctx, 5)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if !completed.Done {
		t.Fatal("Complete() done = false, want true")
	}

	if err := manager.Remove(ctx, 5); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if stub.lastGetID != 5 {
		t.Fatalf("storage saw id %d, want 5", stub.lastGetID)
	}
}

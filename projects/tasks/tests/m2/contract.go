package m2

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// Factory opens one repository path and returns its owner cleanup.
type Factory func(path string) (task.Repository, func() error, error)

// Run exercises the behavior shared by every persistence backend.
func Run(t *testing.T, extension string, factory Factory) {
	t.Helper()
	t.Run("CRUD filters and ordering", func(t *testing.T) {
		repository, cleanup := open(t, extension, factory)
		defer cleanup()
		ctx := context.Background()

		first := create(t, ctx, repository, "first")
		second := create(t, ctx, repository, "second")
		third := create(t, ctx, repository, "third")
		if first.ID != 1 || second.ID != 2 || third.ID != 3 {
			t.Fatalf("created IDs = %d, %d, %d; want 1, 2, 3", first.ID, second.ID, third.ID)
		}
		if first.Completed {
			t.Fatal("new task is completed; want incomplete")
		}

		completed := true
		renamed := "second updated"
		updated, err := repository.Update(ctx, second.ID, task.UpdateInput{
			Title:     &renamed,
			Completed: &completed,
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Title != renamed || !updated.Completed {
			t.Fatalf("updated task = %#v", updated)
		}

		incomplete := false
		got, err := repository.List(ctx, task.ListFilter{Completed: &incomplete})
		if err != nil {
			t.Fatalf("List(false): %v", err)
		}
		assertIDs(t, got, first.ID, third.ID)
		got, err = repository.List(ctx, task.ListFilter{Completed: &completed})
		if err != nil {
			t.Fatalf("List(true): %v", err)
		}
		assertIDs(t, got, second.ID)

		explicitFalse, err := repository.Update(ctx, second.ID, task.UpdateInput{Completed: &incomplete})
		if err != nil {
			t.Fatalf("Update(false): %v", err)
		}
		if explicitFalse.Completed {
			t.Fatal("explicit false update was lost")
		}
		noOp, err := repository.Update(ctx, second.ID, task.UpdateInput{Title: &renamed})
		if err != nil {
			t.Fatalf("no-op Update: %v", err)
		}
		if noOp != explicitFalse {
			t.Fatalf("no-op update = %#v; want %#v", noOp, explicitFalse)
		}

		gotOne, err := repository.Get(ctx, second.ID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if gotOne != noOp {
			t.Fatalf("Get = %#v; want %#v", gotOne, noOp)
		}
		if err := repository.Delete(ctx, second.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		got, err = repository.List(ctx, task.ListFilter{})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		assertIDs(t, got, first.ID, third.ID)
	})

	t.Run("missing IDs", func(t *testing.T) {
		repository, cleanup := open(t, extension, factory)
		defer cleanup()
		ctx := context.Background()
		title := "missing"
		completed := false

		if _, err := repository.Get(ctx, 99); !errors.Is(err, task.ErrNotFound) {
			t.Fatalf("Get error = %v; want ErrNotFound", err)
		}
		if _, err := repository.Update(ctx, 99, task.UpdateInput{Title: &title, Completed: &completed}); !errors.Is(err, task.ErrNotFound) {
			t.Fatalf("Update error = %v; want ErrNotFound", err)
		}
		if err := repository.Delete(ctx, 99); !errors.Is(err, task.ErrNotFound) {
			t.Fatalf("Delete error = %v; want ErrNotFound", err)
		}
		got, err := repository.List(ctx, task.ListFilter{})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("List after missing mutations = %#v; want empty", got)
		}
	})

	t.Run("restart and ID non-reuse", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tasks"+extension)
		repository, cleanup, err := factory(path)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		ctx := context.Background()
		first := create(t, ctx, repository, "first")
		second := create(t, ctx, repository, "second")
		if err := repository.Delete(ctx, first.ID); err != nil {
			t.Fatalf("Delete first: %v", err)
		}
		if err := cleanup(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		repository, cleanup, err = factory(path)
		if err != nil {
			t.Fatalf("reopen: %v", err)
		}
		persisted, err := repository.Get(ctx, second.ID)
		if err != nil {
			t.Fatalf("Get after restart: %v", err)
		}
		if persisted != second {
			t.Fatalf("task after restart = %#v; want %#v", persisted, second)
		}
		if err := repository.Delete(ctx, second.ID); err != nil {
			t.Fatalf("Delete second: %v", err)
		}
		if err := cleanup(); err != nil {
			t.Fatalf("second Close: %v", err)
		}

		repository, cleanup, err = factory(path)
		if err != nil {
			t.Fatalf("second reopen: %v", err)
		}
		defer cleanup()
		third := create(t, ctx, repository, "third")
		if third.ID != 3 {
			t.Fatalf("ID after delete-all and restart = %d; want 3", third.ID)
		}
		got, err := repository.List(ctx, task.ListFilter{})
		if err != nil {
			t.Fatalf("List after restart: %v", err)
		}
		if !reflect.DeepEqual(got, []task.Task{third}) {
			t.Fatalf("List after restart = %#v; want %#v", got, []task.Task{third})
		}
	})

	t.Run("concurrent creates", func(t *testing.T) {
		repository, cleanup := open(t, extension, factory)
		defer cleanup()
		const count = 24
		var wait sync.WaitGroup
		errorsChannel := make(chan error, count)
		for index := 0; index < count; index++ {
			wait.Add(1)
			go func(index int) {
				defer wait.Done()
				_, err := repository.Create(context.Background(), task.CreateInput{
					Title: fmt.Sprintf("task %02d", index),
				})
				errorsChannel <- err
			}(index)
		}
		wait.Wait()
		close(errorsChannel)
		for err := range errorsChannel {
			if err != nil {
				t.Errorf("Create: %v", err)
			}
		}

		got, err := repository.List(context.Background(), task.ListFilter{})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(got) != count {
			t.Fatalf("task count = %d; want %d", len(got), count)
		}
		if !sort.SliceIsSorted(got, func(i, j int) bool { return got[i].ID < got[j].ID }) {
			t.Fatalf("tasks are not ordered: %#v", got)
		}
		for index, value := range got {
			if value.ID != int64(index+1) {
				t.Fatalf("task %d ID = %d; want %d", index, value.ID, index+1)
			}
		}
	})

	t.Run("canceled contexts", func(t *testing.T) {
		repository, cleanup := open(t, extension, factory)
		defer cleanup()
		existing := create(t, context.Background(), repository, "existing")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		title := "changed"
		completed := true

		operations := []struct {
			name string
			call func() error
		}{
			{"create", func() error {
				_, err := repository.Create(ctx, task.CreateInput{Title: "new"})
				return err
			}},
			{"list", func() error {
				_, err := repository.List(ctx, task.ListFilter{})
				return err
			}},
			{"get", func() error {
				_, err := repository.Get(ctx, existing.ID)
				return err
			}},
			{"update", func() error {
				_, err := repository.Update(ctx, existing.ID, task.UpdateInput{Title: &title, Completed: &completed})
				return err
			}},
			{"delete", func() error {
				return repository.Delete(ctx, existing.ID)
			}},
		}
		for _, operation := range operations {
			t.Run(operation.name, func(t *testing.T) {
				err := operation.call()
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error = %v; want context.Canceled", err)
				}
			})
		}
		got, err := repository.Get(context.Background(), existing.ID)
		if err != nil {
			t.Fatalf("Get after canceled mutations: %v", err)
		}
		if got != existing {
			t.Fatalf("task after canceled mutations = %#v; want %#v", got, existing)
		}
	})
}

func open(t *testing.T, extension string, factory Factory) (task.Repository, func() error) {
	t.Helper()
	repository, cleanup, err := factory(filepath.Join(t.TempDir(), "tasks"+extension))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return repository, cleanup
}

func create(t *testing.T, ctx context.Context, repository task.Repository, title string) task.Task {
	t.Helper()
	created, err := repository.Create(ctx, task.CreateInput{Title: title})
	if err != nil {
		t.Fatalf("Create(%q): %v", title, err)
	}
	return created
}

func assertIDs(t *testing.T, tasks []task.Task, want ...int64) {
	t.Helper()
	got := make([]int64, len(tasks))
	for index, value := range tasks {
		got[index] = value.ID
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("IDs = %v; want %v", got, want)
	}
}

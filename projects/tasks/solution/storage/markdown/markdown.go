package markdown

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

const (
	formatVersion = 1
	header        = "# Tasks"
)

var (
	metadataPattern = regexp.MustCompile(`^<!-- rest-task-api:v([0-9]+) next-id=([1-9][0-9]*) -->$`)
	rowPattern      = regexp.MustCompile(`^- \[([ x])\] ([1-9][0-9]*): (.+)$`)
)

// Repository stores tasks in one Markdown checklist.
type Repository struct {
	path string
	mu   sync.Mutex
}

var _ task.Repository = (*Repository)(nil)

// Open opens path, initializing it only when it does not exist.
func Open(path string) (*Repository, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, task.WrapStorage("open markdown", err)
	}
	repository := &Repository{path: absolute}
	repository.mu.Lock()
	defer repository.mu.Unlock()

	_, err = os.Stat(absolute)
	switch {
	case err == nil:
		if _, err := repository.load(); err != nil {
			return nil, task.WrapStorage("open markdown", err)
		}
	case errors.Is(err, os.ErrNotExist):
		if err := repository.save(document{NextID: 1}); err != nil {
			return nil, task.WrapStorage("initialize markdown", err)
		}
	default:
		return nil, task.WrapStorage("open markdown", err)
	}
	return repository, nil
}

// Create appends one incomplete task while preserving monotonic IDs.
func (r *Repository) Create(ctx context.Context, input task.CreateInput) (task.Task, error) {
	if err := lockContext(ctx, &r.mu); err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	defer r.mu.Unlock()

	document, err := r.load()
	if err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	if document.NextID == math.MaxInt64 {
		return task.Task{}, task.WrapStorage("create task", errors.New("markdown store has exhausted task IDs"))
	}
	created := task.Task{ID: document.NextID, Title: input.Title}
	document.NextID++
	document.Tasks = append(document.Tasks, created)
	if err := contextError(ctx); err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	if err := r.save(document); err != nil {
		return task.Task{}, task.WrapStorage("create task", err)
	}
	return created, nil
}

// List returns tasks in ascending ID order, optionally filtered by completion.
func (r *Repository) List(ctx context.Context, filter task.ListFilter) ([]task.Task, error) {
	if err := lockContext(ctx, &r.mu); err != nil {
		return nil, task.WrapStorage("list tasks", err)
	}
	defer r.mu.Unlock()

	document, err := r.load()
	if err != nil {
		return nil, task.WrapStorage("list tasks", err)
	}
	result := make([]task.Task, 0, len(document.Tasks))
	for _, value := range document.Tasks {
		if filter.Completed == nil || value.Completed == *filter.Completed {
			result = append(result, value)
		}
	}
	if err := contextError(ctx); err != nil {
		return nil, task.WrapStorage("list tasks", err)
	}
	return result, nil
}

// Get returns one task by ID.
func (r *Repository) Get(ctx context.Context, id int64) (task.Task, error) {
	if err := lockContext(ctx, &r.mu); err != nil {
		return task.Task{}, task.WrapStorage("get task", err)
	}
	defer r.mu.Unlock()

	document, err := r.load()
	if err != nil {
		return task.Task{}, task.WrapStorage("get task", err)
	}
	index := findTask(document.Tasks, id)
	if index < 0 {
		return task.Task{}, task.NewNotFoundError(id)
	}
	if err := contextError(ctx); err != nil {
		return task.Task{}, task.WrapStorage("get task", err)
	}
	return document.Tasks[index], nil
}

// Update atomically applies the supplied fields to one task.
func (r *Repository) Update(ctx context.Context, id int64, input task.UpdateInput) (task.Task, error) {
	if err := lockContext(ctx, &r.mu); err != nil {
		return task.Task{}, task.WrapStorage("update task", err)
	}
	defer r.mu.Unlock()

	document, err := r.load()
	if err != nil {
		return task.Task{}, task.WrapStorage("update task", err)
	}
	index := findTask(document.Tasks, id)
	if index < 0 {
		return task.Task{}, task.NewNotFoundError(id)
	}
	if input.Title != nil {
		document.Tasks[index].Title = *input.Title
	}
	if input.Completed != nil {
		document.Tasks[index].Completed = *input.Completed
	}
	if err := contextError(ctx); err != nil {
		return task.Task{}, task.WrapStorage("update task", err)
	}
	if err := r.save(document); err != nil {
		return task.Task{}, task.WrapStorage("update task", err)
	}
	return document.Tasks[index], nil
}

// Delete atomically removes one task without changing the next ID.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	if err := lockContext(ctx, &r.mu); err != nil {
		return task.WrapStorage("delete task", err)
	}
	defer r.mu.Unlock()

	document, err := r.load()
	if err != nil {
		return task.WrapStorage("delete task", err)
	}
	index := findTask(document.Tasks, id)
	if index < 0 {
		return task.NewNotFoundError(id)
	}
	document.Tasks = append(document.Tasks[:index], document.Tasks[index+1:]...)
	if err := contextError(ctx); err != nil {
		return task.WrapStorage("delete task", err)
	}
	if err := r.save(document); err != nil {
		return task.WrapStorage("delete task", err)
	}
	return nil
}

type document struct {
	NextID int64
	Tasks  []task.Task
}

func (r *Repository) load() (document, error) {
	content, err := os.ReadFile(r.path)
	if err != nil {
		return document{}, err
	}
	if len(content) == 0 {
		return document{}, errors.New("markdown store is empty")
	}
	if !utf8.Valid(content) {
		return document{}, errors.New("markdown store is not valid UTF-8")
	}
	if content[len(content)-1] != '\n' {
		return document{}, errors.New("markdown store must end with one newline")
	}

	lines := strings.Split(string(content[:len(content)-1]), "\n")
	if len(lines) < 3 || lines[1] != header || lines[2] != "" {
		return document{}, errors.New("markdown store has an invalid header")
	}
	metadata := metadataPattern.FindStringSubmatch(lines[0])
	if metadata == nil {
		return document{}, errors.New("markdown store has invalid metadata")
	}
	version, err := strconv.ParseInt(metadata[1], 10, 64)
	if err != nil || version != formatVersion || metadata[1] != strconv.FormatInt(version, 10) {
		return document{}, fmt.Errorf("unsupported markdown store version %q", metadata[1])
	}
	nextID, err := strconv.ParseInt(metadata[2], 10, 64)
	if err != nil || nextID <= 0 {
		return document{}, errors.New("markdown store has invalid next-id")
	}

	result := document{NextID: nextID, Tasks: make([]task.Task, 0, len(lines)-3)}
	var previousID int64
	for lineNumber, line := range lines[3:] {
		match := rowPattern.FindStringSubmatch(line)
		if match == nil {
			return document{}, fmt.Errorf("invalid checklist row on line %d", lineNumber+4)
		}
		id, err := strconv.ParseInt(match[2], 10, 64)
		if err != nil || id <= 0 {
			return document{}, fmt.Errorf("invalid task ID on line %d", lineNumber+4)
		}
		if id <= previousID {
			return document{}, fmt.Errorf("task IDs must be positive, unique, and ordered")
		}
		title := match[3]
		if err := task.ValidateTitle(title); err != nil {
			return document{}, fmt.Errorf("invalid title on line %d: %w", lineNumber+4, err)
		}
		result.Tasks = append(result.Tasks, task.Task{
			ID:        id,
			Title:     title,
			Completed: match[1] == "x",
		})
		previousID = id
	}
	if nextID <= previousID {
		return document{}, fmt.Errorf("next-id must be greater than every task ID")
	}
	return result, nil
}

func (r *Repository) save(value document) (err error) {
	sort.Slice(value.Tasks, func(i, j int) bool {
		return value.Tasks[i].ID < value.Tasks[j].ID
	})
	content := serialize(value)
	directory := filepath.Dir(r.path)
	base := filepath.Base(r.path)
	temporary, err := os.CreateTemp(directory, "."+base+".tmp-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		if temporary != nil {
			if closeErr := temporary.Close(); err == nil && closeErr != nil {
				err = closeErr
			}
		}
		if removeErr := os.Remove(temporaryPath); err == nil && removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = removeErr
		}
	}()

	if _, err = temporary.WriteString(content); err != nil {
		return err
	}
	if err = temporary.Sync(); err != nil {
		return err
	}
	if err = temporary.Close(); err != nil {
		temporary = nil
		return err
	}
	temporary = nil
	if err = os.Rename(temporaryPath, r.path); err != nil {
		return err
	}
	return syncDirectory(directory)
}

func serialize(value document) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "<!-- rest-task-api:v1 next-id=%d -->\n%s\n\n", value.NextID, header)
	for _, item := range value.Tasks {
		marker := " "
		if item.Completed {
			marker = "x"
		}
		fmt.Fprintf(&builder, "- [%s] %d: %s\n", marker, item.ID, item.Title)
	}
	return builder.String()
}

func findTask(tasks []task.Task, id int64) int {
	index := sort.Search(len(tasks), func(index int) bool {
		return tasks[index].ID >= id
	})
	if index < len(tasks) && tasks[index].ID == id {
		return index
	}
	return -1
}

func lockContext(ctx context.Context, mutex *sync.Mutex) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	mutex.Lock()
	if err := contextError(ctx); err != nil {
		mutex.Unlock()
		return err
	}
	return nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	return ctx.Err()
}

package taskmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"unicode/utf8"
)

// currentSchemaVersion is the on-disk document version this build writes.
const currentSchemaVersion = 1

// document is the persisted file schema. Storing an explicit version and a
// monotonic next-id counter lets the format evolve and guarantees identifiers
// are never reused, even after tasks are removed.
type document struct {
	Version int    `json:"version"`
	NextID  int    `json:"next_id"`
	Tasks   []Task `json:"tasks"`
}

// FileStorage persists tasks to a single JSON file. Writes are atomic: a new
// file is written to a temporary path in the same directory and then renamed
// over the target. A mutex serializes access so the store is safe for
// concurrent use within one process.
type FileStorage struct {
	path string
	mu   sync.Mutex
}

// NewFileStorage returns a FileStorage backed by the file at path. The file is
// created lazily on the first write; a missing file reads as an empty store.
func NewFileStorage(path string) (*FileStorage, error) {
	if path == "" {
		return nil, errors.New("taskmanager: file storage path must not be empty")
	}
	return &FileStorage{path: path}, nil
}

// List returns every stored task.
func (s *FileStorage) List(ctx context.Context) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.load()
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, len(doc.Tasks))
	copy(tasks, doc.Tasks)
	return tasks, nil
}

// Get returns the task with the given identifier, or ErrTaskNotFound.
func (s *FileStorage) Get(ctx context.Context, id int) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.load()
	if err != nil {
		return Task{}, err
	}
	for _, task := range doc.Tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return Task{}, fmt.Errorf("taskmanager: task %d: %w", id, ErrTaskNotFound)
}

// Add stores a new task using the monotonic next identifier.
func (s *FileStorage) Add(ctx context.Context, title string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	normalized, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.load()
	if err != nil {
		return Task{}, err
	}

	task := Task{ID: doc.NextID, Title: normalized, Done: false}
	doc.Tasks = append(doc.Tasks, task)
	doc.NextID++

	if err := s.save(doc); err != nil {
		return Task{}, err
	}
	return task, nil
}

// Complete marks the task with the given identifier as done.
func (s *FileStorage) Complete(ctx context.Context, id int) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.load()
	if err != nil {
		return Task{}, err
	}

	for i := range doc.Tasks {
		if doc.Tasks[i].ID == id {
			doc.Tasks[i].Done = true
			if err := s.save(doc); err != nil {
				return Task{}, err
			}
			return doc.Tasks[i], nil
		}
	}
	return Task{}, fmt.Errorf("taskmanager: task %d: %w", id, ErrTaskNotFound)
}

// Remove deletes the task with the given identifier. The monotonic next
// identifier is preserved so the removed id is never reused.
func (s *FileStorage) Remove(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, err := s.load()
	if err != nil {
		return err
	}

	index := -1
	for i := range doc.Tasks {
		if doc.Tasks[i].ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("taskmanager: task %d: %w", id, ErrTaskNotFound)
	}

	doc.Tasks = append(doc.Tasks[:index], doc.Tasks[index+1:]...)
	return s.save(doc)
}

// load reads and validates the document from disk. A missing file yields an
// empty document; a legacy top-level JSON array is migrated in memory.
func (s *FileStorage) load() (*document, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &document{Version: currentSchemaVersion, NextID: 1}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("taskmanager: read %s: %w", s.path, err)
	}

	if !utf8.Valid(data) {
		return nil, fmt.Errorf("taskmanager: %s is not valid UTF-8", s.path)
	}

	doc, err := parseDocument(data)
	if err != nil {
		return nil, fmt.Errorf("taskmanager: parse %s: %w", s.path, err)
	}
	if err := validateDocument(doc); err != nil {
		return nil, fmt.Errorf("taskmanager: validate %s: %w", s.path, err)
	}
	return doc, nil
}

// parseDocument decodes either the current object schema or the legacy array
// schema, returning a normalized document.
func parseDocument(data []byte) (*document, error) {
	trimmed := skipLeadingSpace(data)
	if len(trimmed) == 0 {
		return &document{Version: currentSchemaVersion, NextID: 1}, nil
	}

	if trimmed[0] == '[' {
		var tasks []Task
		if err := unmarshalStrict(data, &tasks); err != nil {
			return nil, fmt.Errorf("legacy task list: %w", err)
		}
		return &document{Version: currentSchemaVersion, Tasks: tasks}, nil
	}

	var doc document
	if err := unmarshalStrict(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// validateDocument enforces the schema invariants and repairs the monotonic
// next-id counter so it always exceeds the largest stored identifier.
func validateDocument(doc *document) error {
	if doc.Version > currentSchemaVersion {
		return fmt.Errorf("unsupported schema version %d (max supported %d)", doc.Version, currentSchemaVersion)
	}
	if doc.Version <= 0 {
		doc.Version = currentSchemaVersion
	}

	seen := make(map[int]struct{}, len(doc.Tasks))
	maxID := 0
	for i, task := range doc.Tasks {
		if err := task.Validate(); err != nil {
			return fmt.Errorf("task at index %d: %w", i, err)
		}
		if _, duplicate := seen[task.ID]; duplicate {
			return fmt.Errorf("duplicate task id %d", task.ID)
		}
		seen[task.ID] = struct{}{}
		if task.ID > maxID {
			maxID = task.ID
		}
	}

	if doc.NextID <= maxID {
		doc.NextID = maxID + 1
	}
	if doc.NextID < 1 {
		doc.NextID = 1
	}
	return nil
}

// save writes the document atomically: it marshals to a temporary file in the
// target directory, flushes it to disk, and renames it over the destination.
func (s *FileStorage) save(doc *document) error {
	doc.Version = currentSchemaVersion
	if doc.Tasks == nil {
		doc.Tasks = []Task{}
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("taskmanager: encode document: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	temp, err := os.CreateTemp(dir, ".tasks-*.json.tmp")
	if err != nil {
		return fmt.Errorf("taskmanager: create temp file: %w", err)
	}
	tempName := temp.Name()

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempName)
		}
	}()

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("taskmanager: write temp file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("taskmanager: flush temp file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("taskmanager: close temp file: %w", err)
	}

	if err := os.Rename(tempName, s.path); err != nil {
		return fmt.Errorf("taskmanager: replace %s: %w", s.path, err)
	}
	cleanup = false

	syncDir(dir)
	return nil
}

// syncDir best-effort flushes a directory entry so the rename is durable. It
// ignores errors because not all platforms support directory syncing.
func syncDir(dir string) {
	handle, err := os.Open(dir)
	if err != nil {
		return
	}
	defer handle.Close()
	_ = handle.Sync()
}

// skipLeadingSpace returns data without leading JSON whitespace, used only to
// detect whether the payload is an object or a legacy array.
func skipLeadingSpace(data []byte) []byte {
	for i, b := range data {
		switch b {
		case ' ', '\t', '\r', '\n':
			continue
		default:
			return data[i:]
		}
	}
	return nil
}

// unmarshalStrict decodes JSON while rejecting unknown fields and trailing
// data so malformed files fail loudly instead of silently dropping content.
func unmarshalStrict(data []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing data")
		}
		return fmt.Errorf("unexpected trailing data: %w", err)
	}
	return nil
}

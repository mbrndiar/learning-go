package contacts

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// ErrNotDirectory reports a path that exists but is not a directory.
var ErrNotDirectory = errors.New("not a directory")

// EnsureWorkspace creates the inbox, archive, and reports/daily directories
// below root.
func EnsureWorkspace(root string) error {
	for _, relative := range []string{
		"inbox",
		"archive",
		filepath.Join("reports", "daily"),
	} {
		if err := os.MkdirAll(filepath.Join(root, relative), 0o755); err != nil {
			return fmt.Errorf("create workspace directory %s: %w", relative, err)
		}
	}
	return nil
}

// ListRegularFiles recursively returns regular files below root as relative,
// slash-separated paths in deterministic order.
func ListRegularFiles(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat directory %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s: %w", root, ErrNotDirectory)
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if !entryInfo.Mode().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory %s: %w", root, err)
	}
	sort.Strings(files)
	return files, nil
}

// MoveFile moves source to destination, creating destination's parent
// directory first.
func MoveFile(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}
	if err := os.Rename(source, destination); err != nil {
		return fmt.Errorf("move %s to %s: %w", source, destination, err)
	}
	return nil
}

// RemoveEmptyDirectory removes path only when it is a directory and empty.
func RemoveEmptyDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %w", path, ErrNotDirectory)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove empty directory %s: %w", path, err)
	}
	return nil
}

// This lesson covers common directory operations with os, io/fs, and
// path/filepath while keeping every mutation inside a temporary tree.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("directory lesson:", err)
	}
}

func run() (runErr error) {
	root, err := os.MkdirTemp("", "lesson06-directories-*")
	if err != nil {
		return fmt.Errorf("create temporary root: %w", err)
	}
	// RemoveAll is appropriate here because root is the exact directory this
	// function just created and owns. Do not use it on an unchecked path.
	defer func() {
		runErr = errors.Join(runErr, os.RemoveAll(root))
	}()

	if err := createWorkspace(root); err != nil {
		return err
	}

	inbox := filepath.Join(root, "inbox")
	if err := os.WriteFile(filepath.Join(inbox, "alpha.txt"), []byte("alpha\n"), 0o600); err != nil {
		return fmt.Errorf("write alpha: %w", err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "beta.txt"), []byte("beta\n"), 0o600); err != nil {
		return fmt.Errorf("write beta: %w", err)
	}

	fmt.Println("--- ReadDir lists one level ---")
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("read root: %w", err)
	}
	// os.ReadDir returns entries sorted by filename.
	for _, entry := range entries {
		fmt.Printf("%s directory=%t\n", entry.Name(), entry.IsDir())
	}

	fmt.Println("--- Stat describes one path ---")
	info, err := os.Stat(filepath.Join(inbox, "alpha.txt"))
	if err != nil {
		return fmt.Errorf("stat alpha: %w", err)
	}
	fmt.Printf("%s size=%d regular=%t\n", info.Name(), info.Size(), info.Mode().IsRegular())

	fmt.Println("--- WalkDir traverses a tree ---")
	files, err := listRegularFiles(root)
	if err != nil {
		return err
	}
	for _, path := range files {
		fmt.Println(path)
	}

	fmt.Println("--- rename, then remove only an empty directory ---")
	if err := os.Remove(inbox); err != nil {
		fmt.Println("remove non-empty inbox failed:", true)
	}
	archived := filepath.Join(root, "archive", "2026", "alpha.txt")
	if err := moveFile(filepath.Join(inbox, "alpha.txt"), archived); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(inbox, "beta.txt")); err != nil {
		return fmt.Errorf("remove beta: %w", err)
	}
	if err := os.Remove(inbox); err != nil {
		return fmt.Errorf("remove empty inbox: %w", err)
	}
	relative, err := filepath.Rel(root, archived)
	if err != nil {
		return fmt.Errorf("make archive path relative: %w", err)
	}
	fmt.Println("archived:", filepath.ToSlash(relative))
	fmt.Println("inbox removed:", true)

	// filepath.Clean only rewrites path syntax. It does not prove that
	// untrusted input remains inside root, and WalkDir does not follow
	// symbolic-link directories automatically. Security boundaries need an
	// explicit containment policy, not string cleanup alone.
	return nil
}

func createWorkspace(root string) error {
	for _, relative := range []string{
		"inbox",
		filepath.Join("archive", "2026"),
		"reports",
	} {
		if err := os.MkdirAll(filepath.Join(root, relative), 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", relative, err)
		}
	}
	return nil
}

func listRegularFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
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
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	sort.Strings(files)
	return files, nil
}

func moveFile(source, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}
	// Rename is a same-filesystem move. Cross-filesystem moves need an
	// explicit copy-and-remove workflow, and replacement behavior for an
	// existing destination varies by operating system.
	if err := os.Rename(source, destination); err != nil {
		return fmt.Errorf("rename %s to %s: %w", source, destination, err)
	}
	return nil
}

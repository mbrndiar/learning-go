// Package main covers defer, basic file I/O with io/os/bufio, and path
// manipulation with path/filepath.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	demoDeferOrder()
	demoFilepath()
	demoFileWriteAndRead()
}

// demoDeferOrder shows that deferred calls run in last-in-first-out order,
// after the surrounding function's other statements, even if it returns
// early or panics. defer is most often used to pair a resource's
// acquisition (open, lock, connect) with its release right next to it.
func demoDeferOrder() {
	fmt.Println("--- defer order ---")
	fmt.Println("start")
	defer fmt.Println("deferred: first registered, runs last")
	defer fmt.Println("deferred: second registered, runs second-to-last")
	defer fmt.Println("deferred: third registered, runs first")
	fmt.Println("end of function body")
}

// demoFilepath shows the standard, OS-independent way to build and inspect
// paths. Always prefer filepath.Join over manual string concatenation with
// "/" so the code also behaves correctly on Windows.
func demoFilepath() {
	fmt.Println("--- path/filepath ---")

	joined := filepath.Join("reports", "2024", "summary.CSV")
	fmt.Println("joined:", joined)
	fmt.Println("dir:   ", filepath.Dir(joined))
	fmt.Println("base:  ", filepath.Base(joined))
	fmt.Println("ext:   ", filepath.Ext(joined))

	messy := "reports/2024/../2024/./summary.CSV"
	fmt.Println("messy: ", messy)
	fmt.Println("clean: ", filepath.Clean(messy))
}

// demoFileWriteAndRead creates a temporary directory, writes a small text
// file with bufio.Writer, reads it back line by line with bufio.Scanner, and
// removes the directory before returning. os.MkdirTemp guarantees a fresh,
// uniquely named directory so the lesson never collides with real files or
// leaves anything behind.
func demoFileWriteAndRead() {
	fmt.Println("--- os/bufio file I/O ---")

	dir, err := os.MkdirTemp("", "lesson06-files-*")
	if err != nil {
		fmt.Println("create temp dir:", err)
		return
	}
	// Clean up the temporary directory no matter which return path is
	// taken below. Registering this immediately after a successful create
	// is the idiomatic pattern: acquire, then defer the release.
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "notes.txt")

	if err := writeLines(path, []string{"first note", "second note", "third note"}); err != nil {
		fmt.Println("write lines:", err)
		return
	}

	lines, err := readLines(path)
	if err != nil {
		fmt.Println("read lines:", err)
		return
	}

	fmt.Println("read back", len(lines), "lines:")
	for i, line := range lines {
		fmt.Printf("  %d: %s\n", i+1, line)
	}
}

// writeLines opens path for writing (creating or truncating it), writes each
// line through a buffered writer, and closes the file. bufio.Writer batches
// small writes so they are not sent to the operating system one at a time;
// Flush must run before the file is closed or buffered data would be lost.
func writeLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return writer.Flush()
}

// readLines opens path for reading and scans it line by line. bufio.Scanner
// defaults to splitting on newlines, which is exactly what a simple text
// file needs.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return lines, nil
}

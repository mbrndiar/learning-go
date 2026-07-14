package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	markdownLink = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	fencedCode   = regexp.MustCompile("(?s)```.*?```")
	inlineCode   = regexp.MustCompile("`[^`\\n]*`")
)

func main() {
	var failures []string

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := stripCode(string(content))
		for _, match := range markdownLink.FindAllStringSubmatch(text, -1) {
			target := match[1]
			if isExternal(target) {
				continue
			}

			target = strings.SplitN(target, "#", 2)[0]
			if target == "" {
				continue
			}
			decoded, err := url.PathUnescape(target)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s -> %s: invalid escape", path, target))
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(decoded)))
			if _, err := os.Stat(resolved); err != nil {
				failures = append(failures, fmt.Sprintf("%s -> %s", path, target))
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "check links:", err)
		os.Exit(1)
	}
	if len(failures) > 0 {
		fmt.Fprintln(os.Stderr, "missing Markdown links:")
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, " -", failure)
		}
		os.Exit(1)
	}

	fmt.Println("Markdown links: OK")
}

func stripCode(markdown string) string {
	markdown = fencedCode.ReplaceAllString(markdown, "")
	return inlineCode.ReplaceAllString(markdown, "")
}

func isExternal(target string) bool {
	return strings.HasPrefix(target, "http://") ||
		strings.HasPrefix(target, "https://") ||
		strings.HasPrefix(target, "mailto:") ||
		strings.HasPrefix(target, "#")
}

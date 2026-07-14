package main

import (
	"strings"
	"testing"
)

func TestStripCode(t *testing.T) {
	t.Parallel()

	markdown := strings.Join([]string{
		"[real](README.md)",
		"`Contains[T](values []T)`",
		"```go",
		"func Map[T any](values []T) {}",
		"```",
	}, "\n")

	got := stripCode(markdown)
	if !strings.Contains(got, "[real](README.md)") {
		t.Fatal("stripCode removed a real Markdown link")
	}
	if strings.Contains(got, "Contains") || strings.Contains(got, "func Map") {
		t.Fatalf("stripCode left code content: %q", got)
	}
}

func TestIsExternal(t *testing.T) {
	t.Parallel()

	for _, target := range []string{
		"https://go.dev",
		"http://example.com",
		"mailto:gopher@example.com",
		"#section",
	} {
		if !isExternal(target) {
			t.Errorf("isExternal(%q) = false", target)
		}
	}
	if isExternal("lessons/README.md") {
		t.Error("relative repository link reported as external")
	}
}

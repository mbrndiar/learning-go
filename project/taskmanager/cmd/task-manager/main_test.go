package main

import (
	"bytes"
	"context"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mbrndiar/learning-go/project/taskapi"
)

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), args, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

func TestManagerCLIFileBackend(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tasks.json")

	out, _, err := runCLI(t, "-backend", "file", "-file", file, "add", "local task")
	if err != nil {
		t.Fatalf("add error = %v", err)
	}
	if !strings.Contains(out, "added task 1") {
		t.Fatalf("add output = %q", out)
	}

	out, _, err = runCLI(t, "-backend", "file", "-file", file, "list")
	if err != nil {
		t.Fatalf("list error = %v", err)
	}
	if !strings.Contains(out, "local task") {
		t.Fatalf("list output = %q", out)
	}

	if _, _, err := runCLI(t, "-file", file, "complete", "1"); err != nil {
		t.Fatalf("complete error = %v", err)
	}

	out, _, err = runCLI(t, "-file", file, "list")
	if err != nil {
		t.Fatalf("list after complete error = %v", err)
	}
	if !strings.Contains(out, "x") {
		t.Fatalf("list after complete output = %q, want done marker", out)
	}

	if _, _, err := runCLI(t, "-file", file, "remove", "1"); err != nil {
		t.Fatalf("remove error = %v", err)
	}

	out, _, err = runCLI(t, "-file", file, "list")
	if err != nil {
		t.Fatalf("final list error = %v", err)
	}
	if !strings.Contains(out, "no tasks") {
		t.Fatalf("final list output = %q, want no tasks", out)
	}
}

func TestManagerCLIRESTBackend(t *testing.T) {
	store, err := taskapi.OpenSQLiteStore(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	api, err := taskapi.NewAPI(store)
	if err != nil {
		t.Fatalf("NewAPI() error = %v", err)
	}
	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)

	out, _, err := runCLI(t, "-backend", "rest", "-url", server.URL, "add", "remote task")
	if err != nil {
		t.Fatalf("add error = %v", err)
	}
	if !strings.Contains(out, "added task 1") {
		t.Fatalf("add output = %q", out)
	}

	out, _, err = runCLI(t, "-backend", "rest", "-url", server.URL, "list")
	if err != nil {
		t.Fatalf("list error = %v", err)
	}
	if !strings.Contains(out, "remote task") {
		t.Fatalf("list output = %q", out)
	}
}

func TestManagerCLIErrors(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tasks.json")

	tests := []struct {
		name string
		args []string
	}{
		{"no command", []string{"-file", file}},
		{"unknown command", []string{"-file", file, "frobnicate"}},
		{"unknown backend", []string{"-backend", "carrier-pigeon", "list"}},
		{"add without title", []string{"-file", file, "add"}},
		{"add blank title", []string{"-file", file, "add", "   "}},
		{"complete bad id", []string{"-file", file, "complete", "abc"}},
		{"bad flag", []string{"-nope"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := runCLI(t, test.args...); err == nil {
				t.Fatalf("run(%v) error = nil, want error", test.args)
			}
		})
	}
}

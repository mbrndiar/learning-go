package main

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mbrndiar/learning-go/project/taskapi"
)

func newAPIServer(t *testing.T) *httptest.Server {
	t.Helper()
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
	return server
}

func runCLI(t *testing.T, baseURL string, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	full := append([]string{"-url", baseURL}, args...)
	err := run(context.Background(), full, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

func TestClientCLIEndToEnd(t *testing.T) {
	server := newAPIServer(t)

	out, _, err := runCLI(t, server.URL, "add", "buy milk")
	if err != nil {
		t.Fatalf("add error = %v", err)
	}
	if !strings.Contains(out, "added task 1") {
		t.Fatalf("add output = %q, want added task 1", out)
	}

	out, _, err = runCLI(t, server.URL, "list")
	if err != nil {
		t.Fatalf("list error = %v", err)
	}
	if !strings.Contains(out, "buy milk") {
		t.Fatalf("list output = %q, want to contain task", out)
	}

	out, _, err = runCLI(t, server.URL, "complete", "1")
	if err != nil {
		t.Fatalf("complete error = %v", err)
	}
	if !strings.Contains(out, "completed task 1") {
		t.Fatalf("complete output = %q", out)
	}

	out, _, err = runCLI(t, server.URL, "get", "1")
	if err != nil {
		t.Fatalf("get error = %v", err)
	}
	if !strings.Contains(out, "x") {
		t.Fatalf("get output = %q, want done marker", out)
	}

	out, _, err = runCLI(t, server.URL, "remove", "1")
	if err != nil {
		t.Fatalf("remove error = %v", err)
	}
	if !strings.Contains(out, "removed task 1") {
		t.Fatalf("remove output = %q", out)
	}

	out, _, err = runCLI(t, server.URL, "list")
	if err != nil {
		t.Fatalf("final list error = %v", err)
	}
	if !strings.Contains(out, "no tasks") {
		t.Fatalf("final list output = %q, want no tasks", out)
	}
}

func TestClientCLIErrors(t *testing.T) {
	server := newAPIServer(t)

	tests := []struct {
		name string
		args []string
	}{
		{"no command", nil},
		{"unknown command", []string{"frobnicate"}},
		{"add without title", []string{"add"}},
		{"get without id", []string{"get"}},
		{"get bad id", []string{"get", "abc"}},
		{"list with extra arg", []string{"list", "extra"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := runCLI(t, server.URL, test.args...); err == nil {
				t.Fatalf("run(%v) error = nil, want error", test.args)
			}
		})
	}
}

func TestClientCLIBadFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run(context.Background(), []string{"-bogus"}, &stdout, &stderr); err == nil {
		t.Fatal("run(bad flag) error = nil, want error")
	}
}

func TestClientCLIInvalidURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run(context.Background(), []string{"-url", "not-a-url", "list"}, &stdout, &stderr); err == nil {
		t.Fatal("run(bad url) error = nil, want error")
	}
}

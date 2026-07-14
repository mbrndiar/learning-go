package main

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRunFlagError(t *testing.T) {
	var stderr bytes.Buffer
	if err := run([]string{"-nope"}, &stderr); err == nil {
		t.Fatal("run(bad flag) error = nil, want error")
	}
}

func TestRunEmptyDSN(t *testing.T) {
	var stderr bytes.Buffer
	if err := run([]string{"-db", ""}, &stderr); err == nil {
		t.Fatal("run(empty db) error = nil, want error")
	}
}

func TestRunListenError(t *testing.T) {
	// Occupy an address so the server's ListenAndServe fails immediately,
	// exercising the serve-error branch without needing a signal.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	db := filepath.Join(t.TempDir(), "tasks.db")
	runErr := run([]string{"-addr", listener.Addr().String(), "-db", db}, io.Discard)
	if runErr == nil {
		t.Fatal("run(occupied addr) error = nil, want listen error")
	}
}

func TestRunServesAndShutsDownGracefully(t *testing.T) {
	addr := freeAddr(t)
	db := filepath.Join(t.TempDir(), "tasks.db")

	done := make(chan error, 1)
	go func() {
		done <- run([]string{"-addr", addr, "-db", db}, io.Discard)
	}()

	base := "http://" + addr
	waitForServer(t, base)

	// The server is up, which means signal handling is already installed;
	// delivering SIGTERM must trigger a graceful shutdown, not kill the test.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("Kill(SIGTERM) error = %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run() error = %v, want clean shutdown", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("run() did not return after SIGTERM")
	}
}

func freeAddr(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return addr
}

func waitForServer(t *testing.T, base string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/tasks", nil)
		if err != nil {
			cancel()
			t.Fatalf("NewRequest() error = %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server at %s did not become ready", base)
}

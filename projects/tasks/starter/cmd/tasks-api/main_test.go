package main

import (
	"os"
	"testing"
)

func TestStarterServerParsesWithoutCreatingStorage(t *testing.T) {
	path := ".m3-starter-should-not-exist.db"
	_ = os.Remove(path)
	t.Cleanup(func() { _ = os.Remove(path) })
	if exit := run([]string{"--server", "nethttp", "--backend", "sqlite", "--data", path}); exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("storage was created: %v", err)
	}
}

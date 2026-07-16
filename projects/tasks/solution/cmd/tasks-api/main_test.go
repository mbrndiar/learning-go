package main

import (
	"os"
	"testing"
)

func TestUnsupportedServerDoesNotCreateStorage(t *testing.T) {
	path := ".m3-should-not-exist.db"
	_ = os.Remove(path)
	t.Cleanup(func() { _ = os.Remove(path) })
	if exit := run([]string{"--server", "chi", "--data", path}); exit != 2 {
		t.Fatalf("exit = %d", exit)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("storage was created: %v", err)
	}
}

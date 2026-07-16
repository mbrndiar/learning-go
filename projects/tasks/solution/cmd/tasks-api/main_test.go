package main

import (
	"context"
	"os"
	"testing"
)

func TestUnsupportedServerDoesNotCreateStorage(t *testing.T) {
	path := ".m3-should-not-exist.db"
	_ = os.Remove(path)
	t.Cleanup(func() { _ = os.Remove(path) })
	if exit := run([]string{"--server", "unknown", "--data", path}); exit != 2 {
		t.Fatalf("exit = %d", exit)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("storage was created: %v", err)
	}
}

func TestAllServerSelectorsReachComposition(t *testing.T) {
	for _, name := range []string{"nethttp", "chi", "gin"} {
		t.Run(name, func(t *testing.T) {
			path := ".m5-" + name + "-command.db"
			_ = os.Remove(path)
			t.Cleanup(func() { _ = os.Remove(path) })
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			exit := runContext(ctx, []string{
				"--server", name,
				"--backend", "sqlite",
				"--data", path,
				"--port", "0",
			})
			if exit != 0 {
				t.Fatalf("exit = %d", exit)
			}
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("storage was not composed: %v", err)
			}
		})
	}
}

package main

import (
	"context"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
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

func TestLifecycleFailureExitsOneNotTwo(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	path := ".m-lifecycle-exit-one.db"
	_ = os.Remove(path)
	t.Cleanup(func() { _ = os.Remove(path) })

	// A validly-shaped config (server.Validate would accept it) that fails
	// only because the port is already bound must exit 1, distinguishing
	// server.ErrLifecycle failures from server.ErrInvalidConfig ones.
	exit := run([]string{
		"--host", "127.0.0.1",
		"--port", strconv.Itoa(port),
		"--data", path,
	})
	if exit != 1 {
		t.Fatalf("exit = %d, want 1", exit)
	}
}

func TestAllServerSelectorsReachComposition(t *testing.T) {
	for _, name := range []string{"nethttp", "chi", "gin"} {
		t.Run(name, func(t *testing.T) {
			path := ".m5-" + name + "-command.db"
			_ = os.Remove(path)
			t.Cleanup(func() { _ = os.Remove(path) })

			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			port := listener.Addr().(*net.TCPAddr).Port
			if err := listener.Close(); err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			exitCh := make(chan int, 1)
			go func() {
				exitCh <- runContext(ctx, []string{
					"--server", name,
					"--backend", "sqlite",
					"--data", path,
					"--host", "127.0.0.1",
					"--port", strconv.Itoa(port),
				})
			}()

			// A storage file can exist on disk before its schema
			// initialization finishes, so wait for the server to actually
			// accept connections instead of stat-ing the file: that only
			// happens once the repository, handler, and listener have all
			// finished constructing under the still-live context. Canceling
			// before runContext starts (as this test used to) would make an
			// already-done context abort the storage open outright.
			address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
			deadline := time.Now().Add(2 * time.Second)
			for {
				conn, dialErr := net.DialTimeout("tcp", address, 50*time.Millisecond)
				if dialErr == nil {
					_ = conn.Close()
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("%s never started accepting connections: %v", address, dialErr)
				}
				time.Sleep(time.Millisecond)
			}
			cancel()

			select {
			case exit := <-exitCh:
				if exit != 0 {
					t.Fatalf("exit = %d", exit)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("runContext did not return after cancellation")
			}
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("storage was not composed: %v", err)
			}
		})
	}
}

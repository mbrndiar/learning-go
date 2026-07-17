package server_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	apinethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// freeLoopbackPort reserves and releases a loopback TCP port so a test can
// bind config.Port ahead of time and poll it for readiness afterward.
func freeLoopbackPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

// waitForAcceptingConnections blocks until host:port accepts a TCP
// connection or the deadline elapses. A successful dial proves the whole
// composition pipeline (repository open, handler construction, and
// listener bind) has finished, which is the only reliable point at which a
// test can cancel Run's context without racing a still-in-flight storage
// open.
func waitForAcceptingConnections(t *testing.T, host string, port int) {
	t.Helper()
	address := net.JoinHostPort(host, strconv.Itoa(port))
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", address, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s never started accepting connections: %v", address, err)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestRealLoopbackLifecycleRepeated(t *testing.T) {
	for iteration := 0; iteration < 10; iteration++ {
		t.Run("iteration", func(t *testing.T) {
			config := server.DefaultConfig()
			config.Port = 0
			active, err := server.New(config, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(`{"status":"ok"}`))
			}))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			t.Cleanup(func() { _ = active.Close() })
			result := make(chan error, 1)
			go func() { result <- active.Serve(ctx) }()

			httpClient := &http.Client{Timeout: time.Second}
			response, err := httpClient.Get("http://" + active.Addr().String() + "/health")
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != 200 || !bytes.Equal(body, []byte(`{"status":"ok"}`)) {
				t.Fatalf("response = %d %q", response.StatusCode, body)
			}
			cancel()
			select {
			case err := <-result:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("server did not shut down")
			}
			if _, err := httpClient.Get("http://" + active.Addr().String() + "/health"); err == nil {
				t.Fatal("server still accepts connections")
			}
		})
	}
}

func TestUnavailableAddressAndSingleServe(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	config := server.DefaultConfig()
	config.Host = "127.0.0.1"
	config.Port = listener.Addr().(*net.TCPAddr).Port
	if active, err := server.New(config, http.NotFoundHandler()); err == nil || active != nil {
		t.Fatalf("New = %#v, %v", active, err)
	}

	config.Port = 0
	active, err := server.New(config, http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := active.Serve(ctx); err != nil {
		t.Fatal(err)
	}
	if err := active.Serve(context.Background()); err == nil {
		t.Fatal("second Serve unexpectedly succeeded")
	}
}

func TestBothRepositoriesThroughNetHTTP(t *testing.T) {
	directory, err := os.MkdirTemp(".", ".m3-server-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)
	cases := []struct {
		name string
		open func() (task.Repository, func() error, error)
	}{
		{"sqlite", func() (task.Repository, func() error, error) {
			repository, err := sqlite.Open(filepath.Join(directory, "tasks.db"))
			return repository, repository.Close, err
		}},
		{"markdown", func() (task.Repository, func() error, error) {
			repository, err := markdown.Open(filepath.Join(directory, "tasks.md"))
			return repository, func() error { return nil }, err
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repository, closeRepository, err := test.open()
			if err != nil {
				t.Fatal(err)
			}
			defer closeRepository()
			handler := apinethttp.New(task.NewService(repository), slog.New(slog.NewTextHandler(io.Discard, nil)))
			config := server.DefaultConfig()
			config.Port = 0
			active, err := server.New(config, handler)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			t.Cleanup(func() { _ = active.Close() })
			result := make(chan error, 1)
			go func() { result <- active.Serve(ctx) }()
			request, _ := http.NewRequest(http.MethodPost, "http://"+active.Addr().String()+"/tasks",
				bytes.NewBufferString(`{"title":"smoke"}`))
			request.Header.Set("Content-Type", "application/json")
			response, err := (&http.Client{Timeout: time.Second}).Do(request)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			cancel()
			if serveErr := <-result; serveErr != nil {
				t.Fatal(serveErr)
			}
			if response.StatusCode != 201 || string(body) != `{"id":1,"title":"smoke","completed":false}`+"\n" {
				t.Fatalf("response = %d %q", response.StatusCode, body)
			}
		})
	}
}

func TestAllServerSelectorsValidate(t *testing.T) {
	for _, name := range []string{"nethttp", "chi", "gin"} {
		config := server.DefaultConfig()
		config.Server = name
		validated, err := config.Validate()
		if err != nil || validated.Server != name {
			t.Fatalf("server %q validation = %#v, %v", name, validated, err)
		}
	}
}

func TestRunComposesEveryServerAndBackend(t *testing.T) {
	directory, err := os.MkdirTemp(".", ".tasks-run-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, serverName := range []string{"nethttp", "chi", "gin"} {
		for _, backend := range []string{"sqlite", "markdown"} {
			t.Run(serverName+"/"+backend, func(t *testing.T) {
				config := server.DefaultConfig()
				config.Server = serverName
				config.Backend = backend
				config.Host = "127.0.0.1"
				config.Port = freeLoopbackPort(t)
				config.Data = filepath.Join(directory, serverName+"."+backend)

				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)

				result := make(chan error, 1)
				go func() { result <- server.Run(ctx, config, logger) }()

				// Run's context now reaches the storage opener, so canceling
				// it before Run starts (as this test used to) would make
				// OpenContext fail immediately instead of exercising
				// graceful shutdown. A backend file can exist on disk before
				// its schema/initialization finishes (sqlite creates the
				// file before running its init statement), so wait for the
				// server to actually accept connections instead: that only
				// happens once the repository, handler, and listener have
				// all finished constructing under the still-live context.
				waitForAcceptingConnections(t, config.Host, config.Port)
				cancel()

				select {
				case err := <-result:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(2 * time.Second):
					t.Fatal("Run did not return after cancellation")
				}
			})
		}
	}
}

func TestConfigAndLifecycleRejectInvalidBoundaries(t *testing.T) {
	valid := server.DefaultConfig()
	tests := []struct {
		name   string
		mutate func(*server.Config)
	}{
		{"server", func(config *server.Config) { config.Server = "fiber" }},
		{"backend", func(config *server.Config) { config.Backend = "memory" }},
		{"data", func(config *server.Config) { config.Data = "" }},
		{"host", func(config *server.Config) { config.Host = "example.com" }},
		{"negative port", func(config *server.Config) { config.Port = -1 }},
		{"large port", func(config *server.Config) { config.Port = 65536 }},
		{"timeout", func(config *server.Config) { config.ShutdownTimeout = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := valid
			test.mutate(&config)
			if _, err := config.Validate(); !errors.Is(err, server.ErrInvalidConfig) {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}

	config := valid
	config.Server = ""
	validated, err := config.Validate()
	if err != nil || validated.Server != "nethttp" {
		t.Fatalf("empty server validation = %#v, %v", validated, err)
	}
	if active, err := server.NewWithListener(valid, nil, http.NotFoundHandler()); active != nil ||
		!errors.Is(err, server.ErrInvalidConfig) {
		t.Fatalf("nil listener = %#v, %v", active, err)
	}
	if active, err := server.NewWithListener(valid, &stubListener{}, nil); active != nil ||
		!errors.Is(err, server.ErrInvalidConfig) {
		t.Fatalf("nil handler = %#v, %v", active, err)
	}

	var inactive *server.Server
	if inactive.Addr() != nil {
		t.Fatal("nil server has an address")
	}
	if err := inactive.Close(); err != nil {
		t.Fatalf("nil server Close() = %v", err)
	}
}

func TestCloseBeforeServeClosesOwnedListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	config := server.DefaultConfig()
	active, err := server.NewWithListener(config, listener, http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}

	if err := active.Close(); err != nil {
		t.Fatalf("Close() before Serve = %v", err)
	}

	// http.Server.Close only closes listeners it learned about through Serve,
	// so the owned listener must be closed directly. Dialing it after Close
	// must fail because the listener is gone rather than merely idle.
	if _, err := net.Dial("tcp", listener.Addr().String()); err == nil {
		t.Fatal("listener still accepts connections after Close before Serve")
	}

	// Serve must not hang or panic against an already-closed listener. Close
	// already marked the underlying http.Server as shutting down, so Serve
	// observes http.ErrServerClosed and returns promptly and successfully,
	// the same graceful outcome as calling Serve then canceling ctx.
	result := make(chan error, 1)
	go func() { result <- active.Serve(context.Background()) }()
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Serve() after Close() before Serve = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve() after Close() before Serve did not return")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	config := server.DefaultConfig()
	config.Port = 0
	active, err := server.New(config, http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}

	if err := active.Close(); err != nil {
		t.Fatalf("first Close() = %v", err)
	}
	if err := active.Close(); err != nil {
		t.Fatalf("second Close() = %v", err)
	}
	if err := active.Close(); err != nil {
		t.Fatalf("third Close() = %v", err)
	}
}

func TestCloseAfterGracefulShutdownIsIdempotent(t *testing.T) {
	config := server.DefaultConfig()
	config.Port = 0
	active, err := server.New(config, http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- active.Serve(ctx) }()
	cancel()
	if err := <-result; err != nil {
		t.Fatal(err)
	}

	if err := active.Close(); err != nil {
		t.Fatalf("Close() after Serve returned = %v", err)
	}
	if err := active.Close(); err != nil {
		t.Fatalf("second Close() after Serve returned = %v", err)
	}
}

func TestCloseReturnsSameFailureOnEveryCall(t *testing.T) {
	underlying := errors.New("boom: disk unplugged")
	listener := &countingCloseListener{closeErr: underlying}
	config := server.DefaultConfig()
	active, err := server.NewWithListener(config, listener, http.NotFoundHandler())
	if err != nil {
		t.Fatal(err)
	}

	first := active.Close()
	if first == nil {
		t.Fatal("Close() = nil, want a wrapped listener close failure")
	}
	if !errors.Is(first, server.ErrLifecycle) {
		t.Fatalf("Close() error = %v, want to match ErrLifecycle", first)
	}
	if !errors.Is(first, underlying) {
		t.Fatalf("Close() error = %v, want to match the underlying close error via %%w", first)
	}

	for i := 0; i < 3; i++ {
		if again := active.Close(); !errors.Is(again, underlying) || again != first {
			t.Fatalf("Close() call %d = %v, want the identical first-call result %v", i, again, first)
		}
	}

	if listener.closeCalls != 1 {
		t.Fatalf("listener.Close() was called %d times, want exactly 1 (sync.Once should suppress repeats)", listener.closeCalls)
	}
}

// countingCloseListener is a minimal net.Listener whose Close returns a
// caller-supplied, non-sentinel error and counts how many times Close ran,
// so tests can assert Close's sync.Once guard prevents repeated work.
type countingCloseListener struct {
	closeErr   error
	closeCalls int
}

func (*countingCloseListener) Accept() (net.Conn, error) { return nil, errors.New("unused") }
func (listener *countingCloseListener) Close() error {
	listener.closeCalls++
	return listener.closeErr
}
func (*countingCloseListener) Addr() net.Addr { return &net.TCPAddr{} }

type stubListener struct{}

func (*stubListener) Accept() (net.Conn, error) { return nil, errors.New("unused") }
func (*stubListener) Close() error              { return nil }
func (*stubListener) Addr() net.Addr            { return nil }

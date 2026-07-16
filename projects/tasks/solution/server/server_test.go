package server_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	apinethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

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

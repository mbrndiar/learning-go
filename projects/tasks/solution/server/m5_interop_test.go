package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/api"
	apichi "github.com/mbrndiar/learning-go/projects/tasks/solution/api/chi"
	apigin "github.com/mbrndiar/learning-go/projects/tasks/solution/api/gin"
	apinethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	clientnethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/client/nethttp"
	clientresty "github.com/mbrndiar/learning-go/projects/tasks/solution/client/resty"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

func TestMilestone5CompleteInteroperabilityMatrix(t *testing.T) {
	directory, err := os.MkdirTemp(".", ".m5-interoperability-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	serverFactories := []struct {
		name string
		new  func(api.Service) http.Handler
	}{
		{"nethttp", func(service api.Service) http.Handler { return apinethttp.New(service, logger) }},
		{"chi", func(service api.Service) http.Handler { return apichi.New(service, logger) }},
		{"gin", func(service api.Service) http.Handler { return apigin.New(service, logger) }},
	}
	clientFactories := []struct {
		name string
		new  func(client.Config) (client.Transport, error)
	}{
		{"nethttp", func(config client.Config) (client.Transport, error) { return clientnethttp.New(config) }},
		{"resty", func(config client.Config) (client.Transport, error) { return clientresty.New(config) }},
	}
	backendFactories := []struct {
		name      string
		extension string
		open      func(string) (task.Repository, func() error, error)
	}{
		{"sqlite", ".db", func(path string) (task.Repository, func() error, error) {
			repository, openErr := sqlite.Open(path)
			if openErr != nil {
				return nil, func() error { return nil }, openErr
			}
			return repository, repository.Close, nil
		}},
		{"markdown", ".md", func(path string) (task.Repository, func() error, error) {
			repository, openErr := markdown.Open(path)
			return repository, func() error { return nil }, openErr
		}},
	}

	for _, backendFactory := range backendFactories {
		for _, serverFactory := range serverFactories {
			for _, clientFactory := range clientFactories {
				name := backendFactory.name + "/" + serverFactory.name + "/" + clientFactory.name
				t.Run(name, func(t *testing.T) {
					path := filepath.Join(directory, nameToFile(name)+backendFactory.extension)
					runInteroperabilityPhase(t, path, backendFactory.open, serverFactory.new, clientFactory.new, false)
					runInteroperabilityPhase(t, path, backendFactory.open, serverFactory.new, clientFactory.new, true)
				})
			}
		}
	}
}

func runInteroperabilityPhase(
	t *testing.T,
	path string,
	openBackend func(string) (task.Repository, func() error, error),
	newHandler func(api.Service) http.Handler,
	newClient func(client.Config) (client.Transport, error),
	restarted bool,
) {
	t.Helper()
	repository, closeRepository, err := openBackend(path)
	if err != nil {
		t.Fatal(err)
	}
	handler := newHandler(task.NewService(repository))
	active, baseURL, stop := startLoopbackServer(t, handler)
	transport, err := newClient(client.Config{BaseURL: baseURL, Timeout: 2 * time.Second})
	if err != nil {
		stop()
		_ = closeRepository()
		t.Fatal(err)
	}
	defer func() {
		if closer, ok := transport.(interface{ Close() error }); ok {
			if closeErr := closer.Close(); closeErr != nil {
				t.Errorf("close client: %v", closeErr)
			}
		}
		stop()
		if closeErr := active.Close(); closeErr != nil {
			t.Errorf("close server: %v", closeErr)
		}
		if closeErr := closeRepository(); closeErr != nil {
			t.Errorf("close repository: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !restarted {
		assertInitialFlow(t, ctx, transport)
	} else {
		assertRestartFlow(t, ctx, transport)
	}
}

func assertInitialFlow(t *testing.T, ctx context.Context, transport client.Transport) {
	t.Helper()
	first, err := transport.Create(ctx, task.CreateInput{Title: "first"})
	if err != nil || first != (task.Task{ID: 1, Title: "first"}) {
		t.Fatalf("first Create = %#v, %v", first, err)
	}
	second, err := transport.Create(ctx, task.CreateInput{Title: "second"})
	if err != nil || second != (task.Task{ID: 2, Title: "second"}) {
		t.Fatalf("second Create = %#v, %v", second, err)
	}
	completed := true
	updated, err := transport.Update(ctx, second.ID, task.UpdateInput{Completed: &completed})
	if err != nil || updated != (task.Task{ID: 2, Title: "second", Completed: true}) {
		t.Fatalf("Update = %#v, %v", updated, err)
	}
	values, err := transport.List(ctx, task.ListFilter{Completed: &completed})
	if err != nil || len(values) != 1 || values[0] != updated {
		t.Fatalf("filtered List = %#v, %v", values, err)
	}
	if err := transport.Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	third, err := transport.Create(ctx, task.CreateInput{Title: "third"})
	if err != nil || third.ID != 3 {
		t.Fatalf("post-delete Create = %#v, %v", third, err)
	}
}

func assertRestartFlow(t *testing.T, ctx context.Context, transport client.Transport) {
	t.Helper()
	values, err := transport.List(ctx, task.ListFilter{})
	if err != nil || len(values) != 2 || values[0].ID != 2 || values[1].ID != 3 ||
		!values[0].Completed || values[1].Completed {
		t.Fatalf("restart List = %#v, %v", values, err)
	}
	value, err := transport.Get(ctx, 2)
	if err != nil || value.Title != "second" || !value.Completed {
		t.Fatalf("restart Get = %#v, %v", value, err)
	}
	fourth, err := transport.Create(ctx, task.CreateInput{Title: "fourth"})
	if err != nil || fourth.ID != 4 {
		t.Fatalf("restart Create = %#v, %v", fourth, err)
	}
}

func startLoopbackServer(t *testing.T, handler http.Handler) (*server.Server, string, func()) {
	t.Helper()
	config := server.DefaultConfig()
	config.Port = 0
	active, err := server.New(config, handler)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- active.Serve(ctx)
	}()
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		cancel()
		select {
		case serveErr := <-result:
			if serveErr != nil {
				t.Errorf("Serve: %v", serveErr)
			}
		case <-time.After(3 * time.Second):
			_ = active.Close()
			t.Errorf("server did not stop: %s", active.Addr())
		}
	}
	t.Cleanup(stop)
	return active, "http://" + active.Addr().String(), stop
}

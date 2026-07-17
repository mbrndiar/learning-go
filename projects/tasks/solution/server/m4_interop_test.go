package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	clientnethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/client/nethttp"
	clientresty "github.com/mbrndiar/learning-go/projects/tasks/solution/client/resty"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api"
	apichi "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/chi"
	apinethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

func TestMilestone4ClientServerBackendInteroperability(t *testing.T) {
	directory, err := os.MkdirTemp(".", ".m4-interoperability-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	serverFactories := map[string]func(api.Service) http.Handler{
		"nethttp": func(service api.Service) http.Handler { return apinethttp.New(service, logger) },
		"chi":     func(service api.Service) http.Handler { return apichi.New(service, logger) },
	}
	clientFactories := map[string]func(client.Config) (client.Transport, error){
		"nethttp": func(config client.Config) (client.Transport, error) { return clientnethttp.New(config) },
		"resty":   func(config client.Config) (client.Transport, error) { return clientresty.New(config) },
	}
	backendFactories := map[string]func(string) (task.Repository, func() error, error){
		"sqlite": func(path string) (task.Repository, func() error, error) {
			repository, openErr := sqlite.Open(path)
			if openErr != nil {
				return nil, func() error { return nil }, openErr
			}
			return repository, repository.Close, nil
		},
		"markdown": func(path string) (task.Repository, func() error, error) {
			repository, openErr := markdown.Open(path)
			return repository, func() error { return nil }, openErr
		},
	}

	for backendName, openBackend := range backendFactories {
		for serverName, newHandler := range serverFactories {
			for clientName, newClient := range clientFactories {
				name := backendName + "/" + serverName + "/" + clientName
				t.Run(name, func(t *testing.T) {
					extension := ".md"
					if backendName == "sqlite" {
						extension = ".db"
					}
					repository, closeRepository, openErr := openBackend(filepath.Join(directory, nameToFile(name)+extension))
					if openErr != nil {
						t.Fatal(openErr)
					}
					defer closeRepository()
					live := httptest.NewServer(newHandler(task.NewService(repository)))
					defer live.Close()
					transport, newErr := newClient(client.Config{BaseURL: live.URL, Timeout: time.Second})
					if newErr != nil {
						t.Fatal(newErr)
					}
					defer func() {
						if closer, ok := transport.(interface{ Close() error }); ok {
							_ = closer.Close()
						}
					}()

					created, operationErr := transport.Create(context.Background(), task.CreateInput{Title: "interoperable"})
					if operationErr != nil || created.ID != 1 || created.Title != "interoperable" || created.Completed {
						t.Fatalf("Create = %#v, %v", created, operationErr)
					}
					values, operationErr := transport.List(context.Background(), task.ListFilter{})
					if operationErr != nil || len(values) != 1 || values[0] != created {
						t.Fatalf("List = %#v, %v", values, operationErr)
					}
					completed := true
					updated, operationErr := transport.Update(context.Background(), created.ID,
						task.UpdateInput{Completed: &completed})
					if operationErr != nil || !updated.Completed {
						t.Fatalf("Update = %#v, %v", updated, operationErr)
					}
					if operationErr = transport.Delete(context.Background(), created.ID); operationErr != nil {
						t.Fatal(operationErr)
					}
				})
			}
		}
	}
}

func nameToFile(name string) string {
	result := make([]byte, len(name))
	for index := range name {
		if name[index] == '/' {
			result[index] = '-'
		} else {
			result[index] = name[index]
		}
	}
	return string(result)
}

package taskmanager

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/mbrndiar/learning-go/project/taskapi"
	"github.com/mbrndiar/learning-go/project/taskclient"
)

// newRESTStorage stands up an in-memory SQLite-backed API on an httptest server
// and returns a RESTStorage wired to it through the real client. This exercises
// the full Manager -> RESTStorage -> Client -> API -> SQLiteStore path.
func newRESTStorage(t *testing.T) Storage {
	t.Helper()

	store, err := taskapi.OpenSQLiteStore(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("store.Close() error = %v", err)
		}
	})

	api, err := taskapi.NewAPI(store)
	if err != nil {
		t.Fatalf("NewAPI() error = %v", err)
	}

	server := httptest.NewServer(api.Handler())
	t.Cleanup(server.Close)

	client, err := taskclient.New(server.URL, taskclient.WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("taskclient.New() error = %v", err)
	}

	storage, err := NewRESTStorage(client)
	if err != nil {
		t.Fatalf("NewRESTStorage() error = %v", err)
	}
	return storage
}

func TestRESTStorageContract(t *testing.T) {
	runStorageContract(t, newRESTStorage)
}

func TestNewRESTStorageRejectsNilClient(t *testing.T) {
	if _, err := NewRESTStorage(nil); err == nil {
		t.Fatal("NewRESTStorage(nil) error = nil, want error")
	}
}

func TestRESTStorageTranslatesNotFound(t *testing.T) {
	storage := newRESTStorage(t)
	_, err := storage.Get(context.Background(), 4242)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrTaskNotFound", err)
	}
	if !errors.Is(err, taskclient.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want underlying taskclient.ErrNotFound preserved", err)
	}
}

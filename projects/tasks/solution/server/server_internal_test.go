package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// fakeRepository satisfies task.Repository without touching a database, so
// run()'s composition can be exercised with deterministic, injected doubles.
type fakeRepository struct{}

func (fakeRepository) Create(context.Context, task.CreateInput) (task.Task, error) {
	return task.Task{}, nil
}
func (fakeRepository) List(context.Context, task.ListFilter) ([]task.Task, error) { return nil, nil }
func (fakeRepository) Get(context.Context, int64) (task.Task, error)              { return task.Task{}, nil }
func (fakeRepository) Update(context.Context, int64, task.UpdateInput) (task.Task, error) {
	return task.Task{}, nil
}
func (fakeRepository) Delete(context.Context, int64) error { return nil }

// fakeLifecycle lets tests dictate Serve/Close outcomes without binding a
// real listener.
type fakeLifecycle struct {
	serveErr error
	closeErr error
}

func (fake *fakeLifecycle) Serve(context.Context) error { return fake.serveErr }
func (fake *fakeLifecycle) Close() error                { return fake.closeErr }

func fakeDependencies(openErr, handlerErr, newServerErr, closeRepositoryErr error, lc *fakeLifecycle) runDependencies {
	return runDependencies{
		openRepository: func(context.Context, string, string) (task.Repository, func() error, error) {
			if openErr != nil {
				return nil, nil, openErr
			}
			return fakeRepository{}, func() error { return closeRepositoryErr }, nil
		},
		newHandler: func(string, *task.Service, *slog.Logger) (http.Handler, error) {
			if handlerErr != nil {
				return nil, handlerErr
			}
			return http.NotFoundHandler(), nil
		},
		newServer: func(Config, http.Handler) (lifecycle, error) {
			if newServerErr != nil {
				return nil, newServerErr
			}
			return lc, nil
		},
	}
}

func TestRunPropagatesRepositoryOpenFailureWithoutBuildingServer(t *testing.T) {
	openErr := errors.New("open failed")
	deps := fakeDependencies(openErr, nil, nil, nil, nil)
	err := run(context.Background(), DefaultConfig(), slog.Default(), deps)
	if !errors.Is(err, openErr) {
		t.Fatalf("run() error = %v, want %v", err, openErr)
	}
}

func TestRunJoinsHandlerFailureWithRepositoryCloseFailure(t *testing.T) {
	handlerErr := errors.New("handler failed")
	closeErr := errors.New("repository close failed")
	deps := fakeDependencies(nil, handlerErr, nil, closeErr, nil)
	err := run(context.Background(), DefaultConfig(), slog.Default(), deps)
	if !errors.Is(err, handlerErr) {
		t.Fatalf("run() error = %v, want to contain %v", err, handlerErr)
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("run() error = %v, want to contain %v", err, closeErr)
	}
}

func TestRunJoinsServerConstructionFailureWithRepositoryCloseFailure(t *testing.T) {
	newServerErr := errors.New("new server failed")
	closeErr := errors.New("repository close failed")
	deps := fakeDependencies(nil, nil, newServerErr, closeErr, nil)
	err := run(context.Background(), DefaultConfig(), slog.Default(), deps)
	if !errors.Is(err, newServerErr) {
		t.Fatalf("run() error = %v, want to contain %v", err, newServerErr)
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("run() error = %v, want to contain %v", err, closeErr)
	}
}

func TestRunJoinsServeCloseAndRepositoryCleanupFailures(t *testing.T) {
	serveErr := errors.New("serve failed")
	closeErr := errors.New("server close failed")
	repositoryCloseErr := errors.New("repository close failed")
	lc := &fakeLifecycle{serveErr: serveErr, closeErr: closeErr}
	deps := fakeDependencies(nil, nil, nil, repositoryCloseErr, lc)
	err := run(context.Background(), DefaultConfig(), slog.Default(), deps)
	for _, want := range []error{serveErr, closeErr, repositoryCloseErr} {
		if !errors.Is(err, want) {
			t.Fatalf("run() error = %v, want to contain %v", err, want)
		}
	}
}

func TestRunSucceedsWhenServeCloseAndRepositoryCleanupAllSucceed(t *testing.T) {
	lc := &fakeLifecycle{}
	deps := fakeDependencies(nil, nil, nil, nil, lc)
	if err := run(context.Background(), DefaultConfig(), slog.Default(), deps); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}

func TestRunNormalizesNilLoggerBeforeBuildingHandler(t *testing.T) {
	var observed *slog.Logger
	lc := &fakeLifecycle{}
	deps := runDependencies{
		openRepository: func(context.Context, string, string) (task.Repository, func() error, error) {
			return fakeRepository{}, func() error { return nil }, nil
		},
		newHandler: func(_ string, _ *task.Service, logger *slog.Logger) (http.Handler, error) {
			observed = logger
			return http.NotFoundHandler(), nil
		},
		newServer: func(Config, http.Handler) (lifecycle, error) { return lc, nil },
	}
	if err := run(context.Background(), DefaultConfig(), nil, deps); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
	if observed == nil {
		t.Fatal("nil logger was not normalized before reaching newHandler")
	}
}

// contextMarkerKey distinguishes the root context run() is given from any
// other context (e.g. context.Background()), so tests can prove the exact
// value reaches openRepository rather than merely a non-nil context.
type contextMarkerKey struct{}

func TestRunPropagatesRootContextToRepositoryOpener(t *testing.T) {
	var observedCtx context.Context
	lc := &fakeLifecycle{}
	deps := runDependencies{
		openRepository: func(ctx context.Context, _, _ string) (task.Repository, func() error, error) {
			observedCtx = ctx
			return fakeRepository{}, func() error { return nil }, nil
		},
		newHandler: func(string, *task.Service, *slog.Logger) (http.Handler, error) {
			return http.NotFoundHandler(), nil
		},
		newServer: func(Config, http.Handler) (lifecycle, error) { return lc, nil },
	}

	rootCtx := context.WithValue(context.Background(), contextMarkerKey{}, "marker")
	if err := run(rootCtx, DefaultConfig(), slog.Default(), deps); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
	if observedCtx == nil || observedCtx.Value(contextMarkerKey{}) != "marker" {
		t.Fatalf("openRepository ctx = %v, want the root context carrying the test marker value", observedCtx)
	}
}

func TestRunGuardsNilContextBeforeOpeningRepository(t *testing.T) {
	openCalled := false
	lc := &fakeLifecycle{}
	deps := runDependencies{
		openRepository: func(context.Context, string, string) (task.Repository, func() error, error) {
			openCalled = true
			return fakeRepository{}, func() error { return nil }, nil
		},
		newHandler: func(string, *task.Service, *slog.Logger) (http.Handler, error) {
			return http.NotFoundHandler(), nil
		},
		newServer: func(Config, http.Handler) (lifecycle, error) { return lc, nil },
	}

	//lint:ignore SA1012 Deliberately nil to prove run guards it instead of panicking.
	err := run(nil, DefaultConfig(), slog.Default(), deps)
	if !errors.Is(err, ErrLifecycle) {
		t.Fatalf("run(nil, ...) error = %v, want ErrLifecycle", err)
	}
	if openCalled {
		t.Fatal("openRepository was called despite a nil ctx; run() must guard before opening resources")
	}
}

func TestOpenRepositoryBackendDefaultArmRejectsUnsupportedBackend(t *testing.T) {
	if _, _, err := openRepositoryBackend(context.Background(), "memory", "path"); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("openRepositoryBackend() error = %v, want ErrInvalidConfig", err)
	}
}

func TestNewAPIHandlerDefaultArmRejectsUnsupportedServer(t *testing.T) {
	if _, err := newAPIHandler("fiber", task.NewService(fakeRepository{}), slog.Default()); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("newAPIHandler() error = %v, want ErrInvalidConfig", err)
	}
}

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	apichi "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/chi"
	apigin "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/gin"
	apinethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/server/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

// lifecycle is the subset of *Server that Run needs. It lets tests substitute
// a fake that returns deterministic Serve/Close failures without binding a
// real listener.
type lifecycle interface {
	Serve(ctx context.Context) error
	Close() error
}

// runDependencies are the composition seams Run delegates to. Tests replace
// them with in-memory doubles so repository-close and server-close failures
// can be exercised deterministically, without a live database or socket.
type runDependencies struct {
	openRepository func(ctx context.Context, backend, data string) (task.Repository, func() error, error)
	newHandler     func(serverName string, service *task.Service, logger *slog.Logger) (http.Handler, error)
	newServer      func(validated Config, handler http.Handler) (lifecycle, error)
}

func defaultRunDependencies() runDependencies {
	return runDependencies{
		openRepository: openRepositoryBackend,
		newHandler:     newAPIHandler,
		newServer: func(validated Config, handler http.Handler) (lifecycle, error) {
			return newValidated(validated, handler)
		},
	}
}

// openRepositoryBackend opens the repository named by backend using ctx, and
// returns its close function alongside it. Propagating ctx into OpenContext
// lets Run's caller abort a slow open instead of blocking indefinitely. The
// default arm defends against a backend name that reaches here despite
// Config.Validate already rejecting it.
func openRepositoryBackend(ctx context.Context, backend, data string) (task.Repository, func() error, error) {
	switch backend {
	case "sqlite":
		repository, err := sqlite.OpenContext(ctx, data)
		if err != nil {
			return nil, nil, err
		}
		return repository, repository.Close, nil
	case "markdown":
		repository, err := markdown.OpenContext(ctx, data)
		if err != nil {
			return nil, nil, err
		}
		return repository, func() error { return nil }, nil
	default:
		return nil, nil, fmt.Errorf("%w: backend %q is not supported", ErrInvalidConfig, backend)
	}
}

// newAPIHandler builds the HTTP handler named by serverName. The default arm
// defends against a server name that reaches here despite Config.Validate
// already rejecting it.
func newAPIHandler(serverName string, service *task.Service, logger *slog.Logger) (http.Handler, error) {
	switch serverName {
	case "nethttp":
		return apinethttp.New(service, logger), nil
	case "chi":
		return apichi.New(service, logger), nil
	case "gin":
		return apigin.New(service, logger), nil
	default:
		return nil, fmt.Errorf("%w: server %q is not supported", ErrInvalidConfig, serverName)
	}
}

// Run selects adapters, owns their resources, and serves until ctx is canceled.
func Run(ctx context.Context, config Config, logger *slog.Logger) error {
	return run(ctx, config, logger, defaultRunDependencies())
}

// run implements Run against injectable deps so cleanup-error propagation can
// be tested without a live database or listening socket.
func run(ctx context.Context, config Config, logger *slog.Logger, deps runDependencies) error {
	if logger == nil {
		logger = slog.Default()
	}
	validated, err := config.Validate()
	if err != nil {
		return err
	}
	if ctx == nil {
		// context.Context methods panic on a nil interface value, and
		// OpenContext relies on ctx.Done()/ctx.Err(); guard here, before any
		// resource is opened, so a nil ctx fails gracefully instead of
		// panicking deep inside a storage backend.
		return fmt.Errorf("%w: context is required", ErrLifecycle)
	}

	repository, closeRepository, err := deps.openRepository(ctx, validated.Backend, validated.Data)
	if err != nil {
		return err
	}

	service := task.NewService(repository)
	handler, err := deps.newHandler(validated.Server, service, logger)
	if err != nil {
		return errors.Join(err, closeRepository())
	}

	active, err := deps.newServer(validated, handler)
	if err != nil {
		return errors.Join(err, closeRepository())
	}

	serveErr := active.Serve(ctx)
	closeErr := active.Close()
	return errors.Join(serveErr, closeErr, closeRepository())
}

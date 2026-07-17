// Package server is the composition root for Task API processes. It selects
// concrete storage and HTTP adapters, owns their cleanup, and coordinates the
// listening server's lifecycle.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	apichi "github.com/mbrndiar/learning-go/projects/tasks/solution/api/chi"
	apigin "github.com/mbrndiar/learning-go/projects/tasks/solution/api/gin"
	apinethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/api/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/markdown"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/storage/sqlite"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

var (
	// ErrInvalidConfig identifies unsupported or unsafe server configuration.
	ErrInvalidConfig = errors.New("task server: invalid configuration")
	// ErrLifecycle identifies listener, serving, shutdown, or close failures.
	ErrLifecycle = errors.New("task server: lifecycle failure")
)

// Config selects adapters and defines the HTTP server's lifecycle limits.
type Config struct {
	Server            string
	Backend           string
	Data              string
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// DefaultConfig returns the local-learning defaults used by tasks-api.
func DefaultConfig() Config {
	return Config{
		Server: "nethttp", Backend: "sqlite", Data: "tasks.db", Host: "127.0.0.1", Port: 8000,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ShutdownTimeout:   5 * time.Second,
	}
}

// Validate applies defaults and rejects unsupported or unsafe server settings.
func (config Config) Validate() (Config, error) {
	if config.Server == "" {
		config.Server = "nethttp"
	}
	if config.Server != "nethttp" && config.Server != "chi" && config.Server != "gin" {
		return Config{}, fmt.Errorf("%w: server %q is not implemented", ErrInvalidConfig, config.Server)
	}
	if config.Backend != "sqlite" && config.Backend != "markdown" {
		return Config{}, fmt.Errorf("%w: backend must be sqlite or markdown", ErrInvalidConfig)
	}
	if config.Data == "" {
		return Config{}, fmt.Errorf("%w: data path is required", ErrInvalidConfig)
	}
	if net.ParseIP(config.Host) == nil && config.Host != "localhost" {
		return Config{}, fmt.Errorf("%w: host must be an IP address or localhost", ErrInvalidConfig)
	}
	if config.Port < 0 || config.Port > 65535 {
		return Config{}, fmt.Errorf("%w: port must be between 0 and 65535", ErrInvalidConfig)
	}
	if config.ReadHeaderTimeout <= 0 || config.ReadTimeout <= 0 || config.WriteTimeout <= 0 ||
		config.IdleTimeout <= 0 || config.ShutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("%w: all server timeouts must be positive", ErrInvalidConfig)
	}
	return config, nil
}

// Server owns one listener and the HTTP server that serves it.
type Server struct {
	listener  net.Listener
	http      *http.Server
	shutdown  time.Duration
	mu        sync.Mutex
	served    bool
	closeOnce sync.Once
	closeErr  error
}

// New opens a listener and constructs a server that owns it.
func New(config Config, handler http.Handler) (*Server, error) {
	validated, err := config.Validate()
	if err != nil {
		return nil, err
	}
	return newValidated(validated, handler)
}

// newValidated builds a Server bound to validated.Host and validated.Port.
// Callers that already hold a validated Config (Run, tests) use this instead
// of New to avoid re-running Validate.
func newValidated(validated Config, handler http.Handler) (*Server, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(validated.Host, strconv.Itoa(validated.Port)))
	if err != nil {
		return nil, fmt.Errorf("%w: listen: %v", ErrLifecycle, err)
	}
	server, err := newWithValidatedListener(validated, listener, handler)
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	return server, nil
}

// NewWithListener constructs a server around a caller-provided listener.
// Once construction succeeds, Server owns the listener lifecycle.
func NewWithListener(config Config, listener net.Listener, handler http.Handler) (*Server, error) {
	validated, err := config.Validate()
	if err != nil {
		return nil, err
	}
	return newWithValidatedListener(validated, listener, handler)
}

// newWithValidatedListener builds a Server for an already-validated config and
// listener. New and NewWithListener validate their caller-supplied Config
// before delegating here so validation never runs twice for one construction.
func newWithValidatedListener(validated Config, listener net.Listener, handler http.Handler) (*Server, error) {
	if listener == nil || handler == nil {
		return nil, fmt.Errorf("%w: listener and handler are required", ErrInvalidConfig)
	}
	return &Server{
		listener: listener,
		http: &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: validated.ReadHeaderTimeout,
			ReadTimeout:       validated.ReadTimeout,
			WriteTimeout:      validated.WriteTimeout,
			IdleTimeout:       validated.IdleTimeout,
		},
		shutdown: validated.ShutdownTimeout,
	}, nil
}

// Addr returns the bound listener address.
func (server *Server) Addr() net.Addr {
	if server == nil || server.listener == nil {
		return nil
	}
	return server.listener.Addr()
}

// Serve runs the server once and performs graceful shutdown when ctx is canceled.
// The serve goroutine reports through a buffered channel so it can finish even
// if cancellation wins the select. Shutdown uses a fresh context because the
// triggering context is already canceled, then Serve drains the goroutine before
// returning so no listener work remains in the background.
func (server *Server) Serve(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("%w: context is required", ErrLifecycle)
	}
	server.mu.Lock()
	if server.served {
		server.mu.Unlock()
		return fmt.Errorf("%w: server may only be served once", ErrLifecycle)
	}
	server.served = true
	server.mu.Unlock()

	result := make(chan error, 1)
	go func() {
		result <- server.http.Serve(server.listener)
	}()

	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("%w: serve: %v", ErrLifecycle, err)
	case <-ctx.Done():
		// Graceful shutdown needs its own finite lifetime after the parent
		// context has signaled that serving should stop.
		shutdownContext, cancel := context.WithTimeout(context.Background(), server.shutdown)
		shutdownErr := server.http.Shutdown(shutdownContext)
		cancel()
		serveErr := <-result
		if shutdownErr != nil {
			_ = server.http.Close()
			return fmt.Errorf("%w: shutdown: %v", ErrLifecycle, shutdownErr)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return fmt.Errorf("%w: serve: %v", ErrLifecycle, serveErr)
		}
		return nil
	}
}

// Close immediately closes the HTTP server and its listener. Close is
// idempotent: only the first call performs work (guarded by sync.Once), and
// every later call returns that exact same result rather than recomputing
// or masking a first-call failure, including before Serve has ever run
// (http.Server.Close only closes listeners it has learned about through
// Serve, so Close closes server.listener directly to guarantee it is
// released either way). http.ErrServerClosed and net.ErrClosed indicate the
// server or listener was already shut down, not a Close failure, so they are
// ignored; any other failure is wrapped in, and remains matchable as, both
// ErrLifecycle and the underlying close error.
func (server *Server) Close() error {
	if server == nil || server.http == nil {
		return nil
	}
	server.closeOnce.Do(func() {
		httpErr := ignoreExpectedCloseErr(server.http.Close())
		var listenerErr error
		if server.listener != nil {
			listenerErr = ignoreExpectedCloseErr(server.listener.Close())
		}
		if err := errors.Join(httpErr, listenerErr); err != nil {
			server.closeErr = fmt.Errorf("%w: close: %w", ErrLifecycle, err)
		}
	})
	return server.closeErr
}

// ignoreExpectedCloseErr reports nil for the sentinel errors that
// http.Server.Close and net.Listener.Close return when the server or
// listener was already closed, since that outcome is exactly what Close
// wants and is not itself a failure.
func ignoreExpectedCloseErr(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}

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

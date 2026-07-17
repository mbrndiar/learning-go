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
	listener net.Listener
	http     *http.Server
	shutdown time.Duration
	mu       sync.Mutex
	served   bool
}

// New opens a listener and constructs a server that owns it.
func New(config Config, handler http.Handler) (*Server, error) {
	validated, err := config.Validate()
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(validated.Host, strconv.Itoa(validated.Port)))
	if err != nil {
		return nil, fmt.Errorf("%w: listen: %v", ErrLifecycle, err)
	}
	server, err := NewWithListener(validated, listener, handler)
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

// Close immediately closes the HTTP server and its listener.
func (server *Server) Close() error {
	if server == nil || server.http == nil {
		return nil
	}
	if err := server.http.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%w: close: %v", ErrLifecycle, err)
	}
	return nil
}

// Run selects adapters, owns their resources, and serves until ctx is canceled.
func Run(ctx context.Context, config Config, logger *slog.Logger) error {
	validated, err := config.Validate()
	if err != nil {
		return err
	}
	var repository task.Repository
	var closeRepository func() error
	switch validated.Backend {
	case "sqlite":
		value, openErr := sqlite.Open(validated.Data)
		if openErr != nil {
			return openErr
		}
		repository = value
		closeRepository = value.Close
	case "markdown":
		value, openErr := markdown.Open(validated.Data)
		if openErr != nil {
			return openErr
		}
		repository = value
		closeRepository = func() error { return nil }
	}
	defer closeRepository()

	service := task.NewService(repository)
	var handler http.Handler
	switch validated.Server {
	case "nethttp":
		handler = apinethttp.New(service, logger)
	case "chi":
		handler = apichi.New(service, logger)
	case "gin":
		handler = apigin.New(service, logger)
	}
	active, err := New(validated, handler)
	if err != nil {
		return err
	}
	defer active.Close()
	return active.Serve(ctx)
}

// Package server composes Task services, storage, HTTP adapters, and lifecycle.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

var (
	// ErrInvalidConfig identifies unsupported or unsafe server configuration.
	ErrInvalidConfig = errors.New("task server: invalid configuration")
	// ErrLifecycle identifies listener, serving, shutdown, or close failures.
	ErrLifecycle = errors.New("task server: lifecycle failure")
)

// Config selects the HTTP server implementation, storage backend, and
// listener timeouts for one process.
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

// DefaultConfig returns the configuration used when no flags are given.
func DefaultConfig() Config {
	return Config{
		Server: "nethttp", Backend: "sqlite", Data: "tasks.db", Host: "127.0.0.1", Port: 8000,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

// Validate reports whether config names a supported server and backend and
// carries usable timeouts, returning an ErrInvalidConfig failure otherwise.
func (config Config) Validate() (Config, error) {
	return Config{}, task.ErrNotImplemented
}

// Server owns one listening HTTP server and its shutdown lifecycle.
type Server struct {
	listener net.Listener
	http     *http.Server
	shutdown time.Duration
	mu       sync.Mutex
	served   bool
}

// New validates config and builds a Server bound to config.Host and
// config.Port, dispatching requests to handler.
func New(config Config, handler http.Handler) (*Server, error) {
	return nil, task.ErrNotImplemented
}

// NewWithListener builds a Server like New but serves on an already bound
// listener instead of config.Host and config.Port. On success, Server owns the
// listener lifecycle.
func NewWithListener(config Config, listener net.Listener, handler http.Handler) (*Server, error) {
	return nil, task.ErrNotImplemented
}

// Addr returns the address Server is listening on.
func (server *Server) Addr() net.Addr {
	return nil
}

// Serve runs the server once until serving ends or ctx is canceled.
// Cancellation must trigger graceful shutdown within ShutdownTimeout.
func (server *Server) Serve(ctx context.Context) error {
	return task.ErrNotImplemented
}

// Close immediately closes the HTTP server and its listener.
func (server *Server) Close() error {
	return task.ErrNotImplemented
}

// Run validates config, builds the selected HTTP server and storage backend,
// owns their cleanup, and serves until ctx is done. logger receives HTTP
// boundary diagnostics from the selected adapter.
func Run(ctx context.Context, config Config, logger *slog.Logger) error {
	return task.ErrNotImplemented
}

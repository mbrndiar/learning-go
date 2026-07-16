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
	ErrInvalidConfig = errors.New("task server: invalid configuration")
	ErrLifecycle     = errors.New("task server: lifecycle failure")
)

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

func DefaultConfig() Config {
	return Config{
		Server: "nethttp", Backend: "sqlite", Data: "tasks.db", Host: "127.0.0.1", Port: 8000,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

func (config Config) Validate() (Config, error) {
	return Config{}, task.ErrNotImplemented
}

type Server struct {
	listener net.Listener
	http     *http.Server
	shutdown time.Duration
	mu       sync.Mutex
	served   bool
}

func New(config Config, handler http.Handler) (*Server, error) {
	return nil, task.ErrNotImplemented
}

func NewWithListener(config Config, listener net.Listener, handler http.Handler) (*Server, error) {
	return nil, task.ErrNotImplemented
}

func (server *Server) Addr() net.Addr {
	return nil
}

func (server *Server) Serve(ctx context.Context) error {
	return task.ErrNotImplemented
}

func (server *Server) Close() error {
	return task.ErrNotImplemented
}

func Run(ctx context.Context, config Config, logger *slog.Logger) error {
	return task.ErrNotImplemented
}

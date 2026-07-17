package server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

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

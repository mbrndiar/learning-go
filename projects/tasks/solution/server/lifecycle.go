package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

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

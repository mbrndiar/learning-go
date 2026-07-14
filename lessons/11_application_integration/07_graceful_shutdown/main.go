// Command 07_graceful_shutdown shows how to stop an http.Server without
// dropping in-flight requests: listen for a shutdown signal, then call
// Server.Shutdown with its own bounded context instead of killing the
// process outright.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// newServer builds the http.Server this lesson shuts down gracefully. A
// real handler would do real work; here it just proves it ran to
// completion even during shutdown, by writing to inFlight.
func newServer(addr string, onRequestStart, onRequestDone func()) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /work", func(w http.ResponseWriter, r *http.Request) {
		onRequestStart()
		defer onRequestDone()
		time.Sleep(30 * time.Millisecond) // stands in for real work in progress
		fmt.Fprintln(w, "done")
	})

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // guards against slow-header clients
	}
}

// listenAndServe binds server.Addr and starts serving in a background
// goroutine, returning the listener (so callers can discover the actual
// address, useful when Addr is ":0") and a channel that receives Serve's
// final error exactly once.
func listenAndServe(server *http.Server) (net.Listener, chan error, error) {
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, nil, fmt.Errorf("listen: %w", err)
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	return listener, serveErr, nil
}

// shutdown stops server gracefully: new connections are refused
// immediately, and Shutdown blocks until active requests finish or
// shutdownTimeout elapses, whichever comes first. It then drains serveErr
// so the goroutine started by listenAndServe is never leaked.
func shutdown(server *http.Server, shutdownTimeout time.Duration, serveErr chan error) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		<-serveErr
		return fmt.Errorf("shutdown: %w", err)
	}

	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// run starts server, serves until ctx is canceled, and then shuts down
// gracefully within shutdownTimeout. It returns the address it served on.
// Accepting ctx and returning the address keeps this function directly
// testable, without relying on OS signals or a fixed port.
func run(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) (addr string, err error) {
	listener, serveErr, err := listenAndServe(server)
	if err != nil {
		return "", err
	}
	addr = listener.Addr().String()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return addr, nil
		}
		return addr, err
	case <-ctx.Done():
		return addr, shutdown(server, shutdownTimeout, serveErr)
	}
}

func main() {
	// signal.NotifyContext ties context cancellation to an OS signal, so
	// the rest of the program only ever has to think in terms of context
	// cancellation, not signals directly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var inFlight int
	server := newServer(":0",
		func() { inFlight++ },
		func() { inFlight-- },
	)

	// In a real program main would block here until an OS signal arrives.
	// For this runnable demo we cancel immediately after starting so the
	// program exits on its own; main_test.go exercises the in-flight
	// drain behavior explicitly with a real request.
	demoCtx, cancelDemo := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancelDemo()

	addr, err := run(demoCtx, server, 2*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("server stopped cleanly, last address:", addr)
}

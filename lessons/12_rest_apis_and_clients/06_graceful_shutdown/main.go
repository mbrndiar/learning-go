// Command 06_graceful_shutdown drains in-flight requests after cancellation.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func serveUntilCanceled(ctx context.Context, listener net.Listener, handler http.Handler, timeout time.Duration) error {
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		err := <-serveErr
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = w.Write([]byte("finished"))
	})
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() { serverDone <- serveUntilCanceled(ctx, listener, handler, time.Second) }()
	responseDone := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + listener.Addr().String())
		if err != nil {
			responseDone <- err.Error()
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		responseDone <- string(body)
	}()
	<-started
	cancel()
	close(release)
	fmt.Println("response:", <-responseDone)
	if err := <-serverDone; err != nil {
		panic(err)
	}
}

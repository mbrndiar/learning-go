// Command task-api serves the task HTTP/JSON API backed by SQLite.
//
// It wires a SQLiteStore to the HTTP handler, starts a server with finite
// timeouts, and shuts down gracefully when interrupted.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mbrndiar/learning-go/project/taskapi"
)

// shutdownTimeout bounds how long in-flight requests may take to drain.
const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "task-api:", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("task-api", flag.ContinueOnError)
	flags.SetOutput(stderr)
	addr := flags.String("addr", ":8080", "address to listen on")
	dsn := flags.String("db", "tasks.db", "SQLite data source (file path or :memory:)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := taskapi.OpenSQLiteStore(ctx, *dsn)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			logger.Error("close store", slog.String("error", closeErr.Error()))
		}
	}()

	api, err := taskapi.NewAPI(store, taskapi.WithLogger(logger))
	if err != nil {
		return err
	}

	server := taskapi.NewServer(*addr, api.Handler())

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("task-api listening", slog.String("addr", *addr), slog.String("db", *dsn))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("task-api stopped")
	return nil
}

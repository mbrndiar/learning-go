// Command structuredlogging demonstrates log/slog: leveled, structured
// logging with key-value attributes, groups, text vs. JSON output, and a
// runtime-adjustable minimum level.
//
// Try it:
//
//	go run ./lessons/09_tooling_cli_observability/02_structured_logging
//	go run ./lessons/09_tooling_cli_observability/02_structured_logging -format=json
//	go run ./lessons/09_tooling_cli_observability/02_structured_logging -level=debug
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"
)

func main() {
	format := flag.String("format", "text", "log output format: text or json")
	levelFlag := flag.String("level", "info", "minimum level: debug, info, warn, or error")
	flag.Parse()

	level, err := parseLevel(*levelFlag)
	if err != nil {
		slog.Error("invalid -level flag", "error", err)
		os.Exit(2)
	}

	// slog.LevelVar allows the minimum level to change while the program
	// runs (for example from a config reload or a debug endpoint); a plain
	// slog.Level works too if you never need to adjust it later.
	var levelVar slog.LevelVar
	levelVar.Set(level)

	handlerOpts := &slog.HandlerOptions{Level: &levelVar}

	var handler slog.Handler
	switch *format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	default:
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger) // so slog.Info/slog.Error package-level calls use it too

	logger.Debug("starting up", "format", *format) // hidden unless -level=debug
	logger.Info("server starting", "addr", ":8080", "pid", os.Getpid())

	// logger.With returns a child logger that always includes the given
	// attributes, useful for attaching request-scoped or component-scoped
	// context without repeating it at every call site.
	requestLogger := logger.With("component", "http", "request_id", "abc123")
	requestLogger.Info("handling request", "method", "GET", "path", "/tasks")

	// Groups nest attributes under a common key in the output structure.
	logger.Info("request completed",
		slog.Group("http",
			slog.String("method", "GET"),
			slog.String("path", "/tasks"),
			slog.Int("status", 200),
		),
		slog.Duration("latency", 42*time.Millisecond),
	)

	// slog carries a context.Context through *Context methods, which lets
	// handlers attach trace/span IDs pulled from context in real
	// deployments; this lesson passes context.Background() to keep the
	// example self-contained.
	logger.WarnContext(context.Background(), "slow downstream call", "target", "inventory-service", "took", 1500*time.Millisecond)

	if _, err := os.Open("does-not-exist.txt"); err != nil {
		logger.Error("failed to open config", "error", err)
	}
}

func parseLevel(s string) (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(s))
	return level, err
}

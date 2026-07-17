// Package server is the composition root for Task API processes. It selects
// concrete storage and HTTP adapters, owns their cleanup, and coordinates the
// listening server's lifecycle.
package server

import (
	"errors"
	"fmt"
	"net"
	"time"
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

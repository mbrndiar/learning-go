// Package server composes Task services, storage, HTTP adapters, and lifecycle.
package server

import (
	"errors"
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

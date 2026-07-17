// Package client defines the library-independent remote Task boundary.
package client

import (
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

const (
	// DefaultTimeout is the finite default used by Task clients.
	DefaultTimeout = 5 * time.Second
)

// Config contains shared client configuration.
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// Validate normalizes and checks Config, returning a *ConfigError for the
// first invalid field.
func (c Config) Validate() (Config, error) {
	return Config{}, task.ErrNotImplemented
}

// NormalizeBaseURL validates raw as an absolute HTTP(S) URL and returns its
// canonical form, or a *ConfigError.
func NormalizeBaseURL(raw string) (string, error) {
	return "", task.ErrNotImplemented
}

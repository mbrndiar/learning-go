// Package client defines the library-independent remote Task boundary.
package client

import (
	"net/url"
	"strings"
	"time"
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

// Validate normalizes and validates the configuration.
func (c Config) Validate() (Config, error) {
	baseURL, err := NormalizeBaseURL(c.BaseURL)
	if err != nil {
		return Config{}, err
	}
	if c.Timeout <= 0 {
		return Config{}, &ConfigError{Field: "timeout", Message: "timeout must be positive and finite"}
	}
	c.BaseURL = baseURL
	return c, nil
}

// NormalizeBaseURL validates a server base URL and removes a trailing slash.
func NormalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", &ConfigError{Field: "base-url", Message: "base URL must be an absolute HTTP URL"}
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", &ConfigError{Field: "base-url", Message: "base URL must be an absolute HTTP URL"}
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return parsed.String(), nil
}

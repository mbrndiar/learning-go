// Package subtests implements a tiny key=value config loader and a mock
// resource, used to demonstrate subtests, test helper functions, t.Cleanup,
// and t.TempDir.
package subtests

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds parsed settings.
type Config struct {
	Host string
	Port string
}

// LoadConfig reads a simple "key=value" file (one setting per line, blank
// lines and "#" comments ignored) and returns a validated Config.
func LoadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			return Config{}, fmt.Errorf("load config: invalid line %q, want key=value", line)
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}

	cfg := Config{Host: values["host"], Port: values["port"]}
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate reports whether cfg has the required fields.
func Validate(cfg Config) error {
	if cfg.Host == "" {
		return fmt.Errorf("validate config: host is required")
	}
	if cfg.Port == "" {
		return fmt.Errorf("validate config: port is required")
	}
	return nil
}

// Connection simulates an expensive resource (a network connection, a file
// handle, a database pool...) that must be closed exactly once.
type Connection struct {
	Addr   string
	closed bool
}

// Dial "opens" a connection to addr. Callers must call Close when done.
func Dial(addr string) (*Connection, error) {
	if addr == "" {
		return nil, fmt.Errorf("dial: addr is required")
	}
	return &Connection{Addr: addr}, nil
}

// Close releases the connection. It is safe to call more than once.
func (c *Connection) Close() error {
	c.closed = true
	return nil
}

// Closed reports whether Close has been called.
func (c *Connection) Closed() bool {
	return c.closed
}

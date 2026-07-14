package subtests

import (
	"os"
	"path/filepath"
	"testing"
)

// writeConfigFile is a test helper: it creates path/config.txt with the
// given contents inside a fresh temporary directory and returns the file
// path. t.Helper marks this function as a helper so that failure line
// numbers reported by t.Fatalf below point at the caller, not at this line.
func writeConfigFile(t *testing.T, contents string) string {
	t.Helper()

	// t.TempDir creates a directory unique to this test (or subtest) and
	// registers its own cleanup, so it is removed automatically after the
	// test finishes - no manual os.RemoveAll needed.
	dir := t.TempDir()
	path := filepath.Join(dir, "config.txt")

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writeConfigFile: %v", err)
	}
	return path
}

// openTestConnection is a helper that opens a Connection and registers its
// Close with t.Cleanup, so it runs automatically when the test ends - even
// if the test fails or calls t.Fatal partway through.
func openTestConnection(t *testing.T, addr string) *Connection {
	t.Helper()

	conn, err := Dial(addr)
	if err != nil {
		t.Fatalf("openTestConnection: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("closing test connection: %v", err)
		}
	})
	return conn
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{
			name:     "valid config",
			contents: "host=localhost\nport=8080\n",
			wantHost: "localhost",
			wantPort: "8080",
		},
		{
			name:     "with comments and blank lines",
			contents: "# example config\nhost=example.com\n\nport=443\n",
			wantHost: "example.com",
			wantPort: "443",
		},
		{
			name:     "missing port",
			contents: "host=localhost\n",
			wantErr:  true,
		},
		{
			name:     "malformed line",
			contents: "not-a-key-value-pair\n",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeConfigFile(t, test.contents)

			cfg, err := LoadConfig(path)

			if test.wantErr {
				if err == nil {
					t.Fatalf("LoadConfig() = %+v, nil; want an error", cfg)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig() returned unexpected error: %v", err)
			}
			if cfg.Host != test.wantHost || cfg.Port != test.wantPort {
				t.Errorf("LoadConfig() = %+v, want Host=%q Port=%q", cfg, test.wantHost, test.wantPort)
			}
		})
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadConfig(filepath.Join(dir, "does-not-exist.txt"))
	if err == nil {
		t.Fatal("LoadConfig() with a missing file: want an error, got nil")
	}
}

func TestConnectionClosedByCleanup(t *testing.T) {
	conn := openTestConnection(t, "localhost:5432")
	if conn.Closed() {
		t.Fatal("connection should not be closed yet")
	}
	// No explicit Close call here: t.Cleanup (registered inside
	// openTestConnection) closes it automatically after this test returns.
}

func TestDialRejectsEmptyAddr(t *testing.T) {
	if _, err := Dial(""); err == nil {
		t.Fatal("Dial(\"\") = nil error, want an error")
	}
}

package cliapp

import (
	"bytes"
	"errors"
	"flag"
	"strconv"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     Config
		wantErr  bool
		wantHelp bool
	}{
		{
			name: "defaults",
			args: nil,
			want: Config{Name: "World", Count: 1, Verbose: false, LogFormat: "text"},
		},
		{
			name: "all flags set",
			args: []string{"-name=Ada", "-count=3", "-verbose", "-log-format=json"},
			want: Config{Name: "Ada", Count: 3, Verbose: true, LogFormat: "json"},
		},
		{
			name:    "count must be positive",
			args:    []string{"-count=0"},
			wantErr: true,
		},
		{
			name:    "negative count",
			args:    []string{"-count=-1"},
			wantErr: true,
		},
		{
			name:    "unknown log format",
			args:    []string{"-log-format=xml"},
			wantErr: true,
		},
		{
			name:    "unknown flag",
			args:    []string{"-bogus"},
			wantErr: true,
		},
		{
			name:     "help flag",
			args:     []string{"-help"},
			wantErr:  true,
			wantHelp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args)
			if tt.wantHelp {
				if !errors.Is(err, flag.ErrHelp) {
					t.Fatalf("ParseArgs(%v) error = %v, want flag.ErrHelp", tt.args, err)
				}
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseArgs(%v) = %+v, nil, want an error", tt.args, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseArgs(%v) unexpected error: %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("ParseArgs(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestNewLoggerText(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{LogFormat: "text", Verbose: false}
	logger := NewLogger(cfg, &buf)
	logger.Info("hello", "key", "value")
	out := buf.String()
	if !strings.Contains(out, "msg=hello") {
		t.Errorf("text log output = %q, want it to contain msg=hello", out)
	}
	if !strings.Contains(out, "key=value") {
		t.Errorf("text log output = %q, want it to contain key=value", out)
	}
}

func TestNewLoggerJSON(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{LogFormat: "json", Verbose: false}
	logger := NewLogger(cfg, &buf)
	logger.Info("hello", "key", "value")
	out := buf.String()
	if !strings.Contains(out, `"msg":"hello"`) {
		t.Errorf("json log output = %q, want it to contain \"msg\":\"hello\"", out)
	}
	if !strings.Contains(out, `"key":"value"`) {
		t.Errorf("json log output = %q, want it to contain \"key\":\"value\"", out)
	}
}

func TestNewLoggerVerbosity(t *testing.T) {
	tests := []struct {
		name        string
		verbose     bool
		wantDebugIn bool
	}{
		{"debug suppressed by default", false, false},
		{"debug shown when verbose", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(Config{LogFormat: "text", Verbose: tt.verbose}, &buf)
			logger.Debug("debug message")
			contains := strings.Contains(buf.String(), "debug message")
			if contains != tt.wantDebugIn {
				t.Errorf("verbose=%v: debug message present = %v, want %v", tt.verbose, contains, tt.wantDebugIn)
			}
		})
	}
}

func TestGreeting(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []string
	}{
		{"single greeting", Config{Name: "World", Count: 1}, []string{"Hello, World!"}},
		{"repeated greeting", Config{Name: "Ada", Count: 3}, []string{"Hello, Ada!", "Hello, Ada!", "Hello, Ada!"}},
		{"zero count", Config{Name: "Ada", Count: 0}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Greeting(tt.cfg)
			if len(got) != len(tt.want) {
				t.Fatalf("Greeting(%+v) = %v, want %v", tt.cfg, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Greeting(%+v)[%d] = %q, want %q", tt.cfg, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDiagnostics(t *testing.T) {
	diag := Diagnostics()
	for _, key := range []string{"go_version", "os", "arch", "num_cpu", "num_goroutine"} {
		if _, ok := diag[key]; !ok {
			t.Errorf("Diagnostics() missing key %q, got %v", key, diag)
		}
	}
	if !strings.HasPrefix(diag["go_version"], "go") {
		t.Errorf("Diagnostics()[go_version] = %q, want it to start with \"go\"", diag["go_version"])
	}
	if n, err := strconv.Atoi(diag["num_cpu"]); err != nil || n <= 0 {
		t.Errorf("Diagnostics()[num_cpu] = %q, want a positive integer", diag["num_cpu"])
	}
}

func TestRunPrintsGreeting(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run([]string{"-name=Ada", "-count=2"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	out := stdout.String()
	if strings.Count(out, "Hello, Ada!") != 2 {
		t.Errorf("stdout = %q, want exactly two greetings", out)
	}
}

func TestRunHelpIsNotAnError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"-help"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() with -help returned error %v, want nil", err)
	}
}

func TestRunInvalidFlagReturnsError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"-count=-5"}, &stdout, &stderr); err == nil {
		t.Fatal("Run() with invalid -count = nil error, want an error")
	}
}

func TestRunLogsDiagnostics(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"-log-format=json", "-verbose"}, &stdout, &stderr); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if !strings.Contains(stderr.String(), "go_version") {
		t.Errorf("stderr = %q, want it to contain diagnostics field go_version", stderr.String())
	}
}

package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/cli"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/scheduler"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/m5"
)

func TestCheckReportAndTargetSelection(t *testing.T) {
	configPath := writeConfig(t, twoTargetConfig())
	checkedAt := time.Date(2026, 7, 16, 8, 0, 0, 123_456_789, time.FixedZone("offset", 7200))
	dependencies := cli.Dependencies{
		Prober: fakeProber{checkedAt: checkedAt},
		Now:    func() time.Time { return checkedAt.Add(-time.Second) },
	}
	var stdout, stderr bytes.Buffer
	exitCode := cli.RunWithDependencies(
		context.Background(),
		[]string{"check", "--config", configPath, "--target", "b", "--target", "a"},
		&stdout,
		&stderr,
		dependencies,
	)
	if exitCode != cli.ExitOK || stderr.Len() != 0 {
		t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
	}
	var report domain.CheckReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v; stdout=%q", err, stdout.String())
	}
	if len(report.Results) != 2 ||
		report.Results[0].Target != "a" ||
		report.Results[1].Target != "b" ||
		report.Results[0].Sequence != 1 ||
		report.Results[1].Sequence != 2 {
		t.Fatalf("results = %+v", report.Results)
	}
	if report.Summary != (domain.Summary{Healthy: 1, Degraded: 1}) {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if got := domain.FormatTime(report.CheckedAt); got != "2026-07-16T06:00:00.123Z" {
		t.Fatalf("checked_at = %s", got)
	}
}

func TestCheckFailuresAndJSONErrors(t *testing.T) {
	validPath := writeConfig(t, twoTargetConfig())
	invalidPath := writeConfig(t, `{}`)
	unsupportedPath := writeConfig(t, strings.Replace(twoTargetConfig(), `"schema_version":1`, `"schema_version":2`, 1))
	duplicatePath := writeConfig(t, strings.Replace(twoTargetConfig(), `"name":"b"`, `"name":"a"`, 1))
	tests := []struct {
		name string
		args []string
		exit int
		code string
	}{
		{name: "missing subcommand", args: nil, exit: cli.ExitUsage, code: "usage"},
		{name: "unknown subcommand", args: []string{"unknown"}, exit: cli.ExitUsage, code: "usage"},
		{name: "missing config flag", args: []string{"check"}, exit: cli.ExitUsage, code: "usage"},
		{name: "unknown flag", args: []string{"check", "--unknown"}, exit: cli.ExitUsage, code: "usage"},
		{name: "missing file", args: []string{"check", "--config", filepath.Join(t.TempDir(), "missing.json")}, exit: cli.ExitConfigIO, code: "config_io"},
		{name: "invalid config", args: []string{"check", "--config", invalidPath}, exit: cli.ExitConfig, code: "invalid_config"},
		{name: "unsupported config", args: []string{"check", "--config", unsupportedPath}, exit: cli.ExitConfig, code: "unsupported_schema"},
		{name: "duplicate target", args: []string{"check", "--config", duplicatePath}, exit: cli.ExitConfig, code: "duplicate_target"},
		{name: "unknown target", args: []string{"check", "--config", validPath, "--target", "missing"}, exit: cli.ExitUsage, code: "target_not_found"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			args := append([]string{"--json-errors"}, test.args...)
			exitCode := cli.RunWithDependencies(context.Background(), args, &stdout, &stderr, cli.Dependencies{})
			if exitCode != test.exit || stdout.Len() != 0 {
				t.Fatalf("exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
			}
			m5.RequireJSONError(t, stderr.Bytes(), test.code)
		})
	}
}

func TestCheckCancellationAndOutputFailure(t *testing.T) {
	configPath := writeConfig(t, twoTargetConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	exitCode := cli.RunWithDependencies(ctx, []string{"--json-errors", "check", "--config", configPath}, &stdout, &stderr, cli.Dependencies{
		Prober: fakeProber{checkedAt: time.Now()},
	})
	if exitCode != cli.ExitCancelled {
		t.Fatalf("exit = %d, stderr=%q", exitCode, stderr.String())
	}
	m5.RequireJSONError(t, stderr.Bytes(), "cancelled")

	stderr.Reset()
	exitCode = cli.RunWithDependencies(context.Background(), []string{"--json-errors", "check", "--config", configPath}, failingWriter{}, &stderr, cli.Dependencies{
		Prober: fakeProber{checkedAt: time.Now()},
	})
	if exitCode != cli.ExitInternal {
		t.Fatalf("exit = %d", exitCode)
	}
	m5.RequireJSONError(t, stderr.Bytes(), "internal")
}

func TestServeLifecycleAndStartupErrors(t *testing.T) {
	configPath := writeConfig(t, twoTargetConfig())
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() {
		done <- cli.RunWithDependencies(ctx, []string{"serve", "--config", configPath, "--listen", "127.0.0.1:0"}, io.Discard, &logs, cli.Dependencies{
			Prober:  fakeProber{checkedAt: time.Now()},
			Trigger: scheduler.NewManualTrigger(),
			Listen: func(string, string) (net.Listener, error) {
				return listener, nil
			},
			Logger:          slog.New(slog.NewJSONHandler(&logs, nil)),
			ShutdownTimeout: time.Second,
		})
	}()
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + listener.Addr().String() + "/healthz")
	if err != nil {
		cancel()
		t.Fatalf("GET healthz: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("healthz status = %d", response.StatusCode)
	}
	cancel()
	select {
	case exitCode := <-done:
		if exitCode != cli.ExitOK {
			t.Fatalf("serve exit = %d; logs=%q", exitCode, logs.String())
		}
	case <-time.After(time.Second):
		t.Fatal("serve did not shut down")
	}
	if !strings.Contains(logs.String(), `"msg":"monitor server started"`) ||
		!strings.Contains(logs.String(), `"msg":"monitor server stopped"`) {
		t.Fatalf("logs = %q", logs.String())
	}

	var stderr bytes.Buffer
	exitCode := cli.RunWithDependencies(context.Background(), []string{"--json-errors", "serve", "--config", configPath, "--listen", "0.0.0.0:1"}, io.Discard, &stderr, cli.Dependencies{})
	if exitCode != cli.ExitUsage {
		t.Fatalf("non-loopback exit = %d", exitCode)
	}
	m5.RequireJSONError(t, stderr.Bytes(), "usage")

	stderr.Reset()
	exitCode = cli.RunWithDependencies(context.Background(), []string{"--json-errors", "serve", "--config", configPath, "--listen", "127.0.0.1:1"}, io.Discard, &stderr, cli.Dependencies{
		Listen: func(string, string) (net.Listener, error) {
			return nil, errors.New("fixture listen failure")
		},
	})
	if exitCode != cli.ExitInternal {
		t.Fatalf("listen failure exit = %d", exitCode)
	}
	m5.RequireJSONError(t, stderr.Bytes(), "server_start")
}

func TestTextErrorsAreConcise(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := cli.Run(context.Background(), nil, &stdout, &stderr)
	if exitCode != cli.ExitUsage || stdout.Len() != 0 || stderr.String() != "monitor: usage: expected check or serve subcommand\n" {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
}

type fakeProber struct {
	checkedAt time.Time
}

func (prober fakeProber) Probe(_ context.Context, target domain.Target) domain.Observation {
	status := domain.StatusHealthy
	httpStatus := http.StatusNoContent
	message := "status 204 was within 200..399"
	if target.Name == "b" {
		status = domain.StatusDegraded
		httpStatus = http.StatusServiceUnavailable
		message = "status 503 was outside 200..399"
	}
	return domain.Observation{
		Target: target.Name, CheckedAt: prober.checkedAt, DurationMS: 12,
		Status: status, HTTPStatus: &httpStatus, Message: message,
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("fixture output failure")
}

func writeConfig(t *testing.T, document string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "monitor.json")
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func twoTargetConfig() string {
	return `{
	  "schema_version":1,
	  "max_concurrency":2,
	  "history_limit":5,
	  "targets":[
	    {"name":"a","url":"http://127.0.0.1/a","interval_ms":100,"timeout_ms":50,"expected_status_min":200,"expected_status_max":399,"max_body_bytes":0},
	    {"name":"b","url":"http://127.0.0.1/b","interval_ms":100,"timeout_ms":50,"expected_status_min":200,"expected_status_max":399,"max_body_bytes":0}
	  ]
	}`
}

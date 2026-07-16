package m5_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/fixtures"
)

func TestSolutionCheckSubprocess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	configPath := filepath.Join(t.TempDir(), "monitor.json")
	if err := os.WriteFile(configPath, []byte(fixtures.ValidConfig(server.URL)), 0o600); err != nil {
		t.Fatal(err)
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test source")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(
		ctx,
		"go",
		"run",
		"./capstones/idiomatic/solution/monitor/cmd/monitor",
		"check",
		"--config",
		configPath,
	)
	command.Dir = repositoryRoot
	output, err := command.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Fatalf("go run failed: %v; stderr=%s", err, exitError.Stderr)
		}
		t.Fatal(err)
	}
	var report domain.CheckReport
	if err := json.Unmarshal(output, &report); err != nil {
		t.Fatalf("decode subprocess output: %v; output=%s", err, output)
	}
	if report.Summary.Healthy != 1 ||
		len(report.Results) != 1 ||
		report.Results[0].Target != "catalog" ||
		report.Results[0].HTTPStatus == nil ||
		*report.Results[0].HTTPStatus != http.StatusNoContent {
		t.Fatalf("report = %+v", report)
	}
}

package monitor_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/api"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/cli"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/probe"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/scheduler"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/contract"
)

func TestHarness(t *testing.T) {
	store := history.NewMemoryStore(1)
	prober := probe.NewHTTPProber(http.DefaultClient)
	scheduler := scheduler.New(prober, store, nil, 1, nil)

	contract.RunHarness(t, contract.Harness{
		Name:        "starter",
		Implemented: domain.Implemented,
		LoadConfig: func() error {
			_, err := domain.LoadConfig(strings.NewReader(`{}`))
			return err
		},
		Probe: func(ctx context.Context) contract.ProbeResult {
			observation := prober.Probe(ctx, domain.Target{Name: "placeholder"})
			return contract.ProbeResult{
				Status:  string(observation.Status),
				Message: observation.Message,
			}
		},
		Record: func() error {
			return store.Record(domain.Observation{})
		},
		Start: scheduler.Start,
		Wait:  scheduler.Wait,
		Serve: func() contract.HTTPResult {
			recorder := httptest.NewRecorder()
			api.NewHandler(store, nil).ServeHTTP(
				recorder,
				httptest.NewRequest(http.MethodGet, "/healthz", nil),
			)
			return contract.HTTPResult{
				StatusCode:  recorder.Code,
				ContentType: recorder.Header().Get("Content-Type"),
				Body:        recorder.Body.String(),
			}
		},
		RunCLI: func(ctx context.Context, stdout, stderr *bytes.Buffer) int {
			return cli.Run(ctx, []string{"check", "--config", "unused.json"}, stdout, stderr)
		},
		IsIncomplete: func(err error) bool {
			return errors.Is(err, domain.ErrNotImplemented)
		},
		ProbeMessage:      "TODO: implement HTTP probing",
		APIResponse:       `{"error":{"code":"not_implemented","message":"TODO: implement monitor HTTP API"}}` + "\n",
		CommandDiagnostic: "monitor: not implemented\n",
		PlaceholderRC:     1,
	})
}

package domain_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/fixtures"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/m1"
)

func TestLoadConfig(t *testing.T) {
	config, err := domain.LoadConfig(strings.NewReader(fixtures.ValidConfig("http://127.0.0.1:9001/health")))
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if config.SchemaVersion != 1 || config.MaxConcurrency != 2 || config.HistoryLimit != 5 {
		t.Fatalf("config = %+v", config)
	}
	if len(config.Targets) != 1 || config.Targets[0].URL != "http://127.0.0.1:9001/health" {
		t.Fatalf("targets = %+v", config.Targets)
	}
}

func TestLoadConfigRejectsInvalidDocuments(t *testing.T) {
	valid := fixtures.ValidConfig("http://127.0.0.1:9001/health")
	tests := []struct {
		name     string
		document string
		sentinel error
	}{
		{name: "empty", document: "", sentinel: domain.ErrInvalidConfig},
		{name: "array", document: "[]", sentinel: domain.ErrInvalidConfig},
		{name: "missing top field", document: `{"schema_version":1}`, sentinel: domain.ErrInvalidConfig},
		{name: "unknown top field", document: strings.Replace(valid, `"schema_version": 1,`, `"schema_version": 1, "extra": true,`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "unknown target field", document: strings.Replace(valid, `"name": "catalog",`, `"name": "catalog", "extra": true,`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "missing target field", document: strings.Replace(valid, `"max_body_bytes": 16`, `"unused": 16`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "fractional number", document: strings.Replace(valid, `"max_concurrency": 2`, `"max_concurrency": 1.5`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "trailing value", document: valid + `{}`, sentinel: domain.ErrInvalidConfig},
		{name: "duplicate JSON field", document: strings.Replace(valid, `"schema_version": 1,`, `"schema_version": 1, "schema_version": 1,`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "unsupported schema", document: strings.Replace(valid, `"schema_version": 1`, `"schema_version": 2`, 1), sentinel: domain.ErrUnsupportedSchema},
		{name: "bad name", document: strings.Replace(valid, `"catalog"`, `"bad name"`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "bad URL scheme", document: strings.Replace(valid, `"http://127.0.0.1:9001/health"`, `"ftp://127.0.0.1/health"`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "missing URL hostname", document: strings.Replace(valid, `"http://127.0.0.1:9001/health"`, `"http://:9001/health"`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "URL user", document: strings.Replace(valid, `"http://127.0.0.1:9001/health"`, `"http://user@127.0.0.1/health"`, 1), sentinel: domain.ErrInvalidConfig},
		{name: "URL fragment", document: strings.Replace(valid, `"http://127.0.0.1:9001/health"`, `"http://127.0.0.1/health#fragment"`, 1), sentinel: domain.ErrInvalidConfig},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := domain.LoadConfig(strings.NewReader(test.document))
			m1.RequireErrorKind(t, err, test.sentinel)
		})
	}
}

func TestLoadConfigBoundaries(t *testing.T) {
	base := map[string]any{
		"schema_version":  1,
		"max_concurrency": 1,
		"history_limit":   1,
		"targets": []any{map[string]any{
			"name": "a", "url": "http://127.0.0.1/",
			"interval_ms": 100, "timeout_ms": 1,
			"expected_status_min": 100, "expected_status_max": 599,
			"max_body_bytes": 0,
		}},
	}
	assertConfig := func(t *testing.T, config map[string]any, valid bool) {
		t.Helper()
		data, err := json.Marshal(config)
		if err != nil {
			t.Fatal(err)
		}
		_, err = domain.LoadConfig(strings.NewReader(string(data)))
		if valid && err != nil {
			t.Fatalf("LoadConfig() error = %v; JSON=%s", err, data)
		}
		if !valid && !errors.Is(err, domain.ErrInvalidConfig) {
			t.Fatalf("LoadConfig() error = %v, want invalid config; JSON=%s", err, data)
		}
	}
	for _, test := range []struct {
		field  string
		values []int
		valid  []bool
	}{
		{field: "max_concurrency", values: []int{0, 1, 32, 33}, valid: []bool{false, true, true, false}},
		{field: "history_limit", values: []int{0, 1, 1000, 1001}, valid: []bool{false, true, true, false}},
	} {
		for index, value := range test.values {
			t.Run(fmt.Sprintf("%s_%d", test.field, value), func(t *testing.T) {
				config := cloneMap(base)
				config[test.field] = value
				assertConfig(t, config, test.valid[index])
			})
		}
	}

	targetTests := []struct {
		field  string
		values []any
		valid  []bool
	}{
		{field: "interval_ms", values: []any{99, 100, 86_400_000, 86_400_001}, valid: []bool{false, true, true, false}},
		{field: "timeout_ms", values: []any{0, 1, 100, 101}, valid: []bool{false, true, true, false}},
		{field: "expected_status_min", values: []any{99, 100, 599, 600}, valid: []bool{false, true, true, false}},
		{field: "expected_status_max", values: []any{99, 100, 599, 600}, valid: []bool{false, true, true, false}},
		{field: "max_body_bytes", values: []any{-1, 0, 1_048_576, 1_048_577}, valid: []bool{false, true, true, false}},
	}
	for _, test := range targetTests {
		for index, value := range test.values {
			t.Run(fmt.Sprintf("%s_%v", test.field, value), func(t *testing.T) {
				config := cloneMap(base)
				target := config["targets"].([]any)[0].(map[string]any)
				target[test.field] = value
				if test.field == "timeout_ms" && value == 101 {
					target["interval_ms"] = 100
				}
				if test.field == "expected_status_min" {
					target["expected_status_max"] = 599
				}
				if test.field == "expected_status_max" {
					target["expected_status_min"] = 100
				}
				assertConfig(t, config, test.valid[index])
			})
		}
	}
}

func TestDuplicateTargetsAndTargetCount(t *testing.T) {
	validTarget := `{
	  "name":"catalog","url":"http://127.0.0.1/","interval_ms":100,"timeout_ms":10,
	  "expected_status_min":200,"expected_status_max":399,"max_body_bytes":0
	}`
	duplicate := fmt.Sprintf(
		`{"schema_version":1,"max_concurrency":1,"history_limit":2,"targets":[%s,%s]}`,
		validTarget,
		validTarget,
	)
	_, err := domain.LoadConfig(strings.NewReader(duplicate))
	m1.RequireErrorKind(t, err, domain.ErrDuplicateTarget)

	noTargets := `{"schema_version":1,"max_concurrency":1,"history_limit":1,"targets":[]}`
	_, err = domain.LoadConfig(strings.NewReader(noTargets))
	m1.RequireErrorKind(t, err, domain.ErrInvalidConfig)
}

func TestLoadConfigWrapsReadErrors(t *testing.T) {
	cause := errors.New("fixture read failure")
	_, err := domain.LoadConfig(errorReader{err: cause})
	m1.RequireErrorKind(t, err, domain.ErrConfigIO)
	if !errors.Is(err, cause) {
		t.Fatalf("error = %v, want wrapped cause", err)
	}
}

func TestLoadConfigRejectsInvalidUTF8(t *testing.T) {
	document := append([]byte(fixtures.ValidConfig("http://127.0.0.1/")), 0xff)
	_, err := domain.LoadConfig(strings.NewReader(string(document)))
	m1.RequireErrorKind(t, err, domain.ErrInvalidConfig)
}

func TestObservationAndReportJSON(t *testing.T) {
	checkedAt := time.Date(2026, 7, 16, 8, 0, 0, 123_987_000, time.FixedZone("offset", 2*60*60))
	status := 204
	observation := domain.Observation{
		Sequence: 1, Target: "catalog", CheckedAt: checkedAt, DurationMS: 12,
		Status: domain.StatusHealthy, PreviousStatus: domain.StatusUnknown, Transition: true,
		HTTPStatus: &status, Message: "ok",
	}
	data, err := json.Marshal(observation)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"checked_at":"2026-07-16T06:00:00.123Z"`) {
		t.Fatalf("observation JSON = %s", data)
	}
	var decoded domain.Observation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.CheckedAt.Equal(checkedAt.Truncate(time.Millisecond)) {
		t.Fatalf("decoded checked_at = %v", decoded.CheckedAt)
	}
	reportData, err := json.Marshal(domain.CheckReport{CheckedAt: checkedAt})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(reportData), `"results":[]`) {
		t.Fatalf("report JSON = %s", reportData)
	}
}

func TestStatusSummaryAndLookup(t *testing.T) {
	if domain.StatusUnknown.Valid() || !domain.StatusHealthy.Valid() {
		t.Fatal("unexpected status validity")
	}
	summary := domain.Summarize([]domain.Observation{
		{Status: domain.StatusHealthy},
		{Status: domain.StatusDegraded},
		{Status: domain.StatusUnhealthy},
		{Status: domain.StatusHealthy},
	})
	if summary != (domain.Summary{Healthy: 2, Degraded: 1, Unhealthy: 1}) {
		t.Fatalf("summary = %+v", summary)
	}
	target, ok := domain.TargetByName([]domain.Target{{Name: "catalog"}}, "catalog")
	if !ok || target.Name != "catalog" {
		t.Fatalf("TargetByName() = %+v, %v", target, ok)
	}
	if _, ok := domain.TargetByName(nil, "missing"); ok {
		t.Fatal("missing target was found")
	}
}

func FuzzLoadConfig(f *testing.F) {
	f.Add(fixtures.ValidConfig("http://127.0.0.1:9001/health"))
	f.Add(`{}`)
	f.Add(`{"schema_version":1,"schema_version":1}`)
	f.Fuzz(func(t *testing.T, document string) {
		config, err := domain.LoadConfig(strings.NewReader(document))
		if err == nil {
			data, marshalErr := json.Marshal(config)
			if marshalErr != nil {
				t.Fatalf("marshal valid config: %v", marshalErr)
			}
			if _, reloadErr := domain.LoadConfig(strings.NewReader(string(data))); reloadErr != nil {
				t.Fatalf("valid config did not round-trip: %v", reloadErr)
			}
		}
	})
}

func ExampleSummarize() {
	summary := domain.Summarize([]domain.Observation{
		{Status: domain.StatusHealthy},
		{Status: domain.StatusDegraded},
		{Status: domain.StatusHealthy},
	})
	fmt.Printf("healthy=%d degraded=%d unhealthy=%d\n", summary.Healthy, summary.Degraded, summary.Unhealthy)
	// Output: healthy=2 degraded=1 unhealthy=0
}

type errorReader struct {
	err error
}

func (reader errorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func cloneMap(source map[string]any) map[string]any {
	data, _ := json.Marshal(source)
	var clone map[string]any
	_ = json.Unmarshal(data, &clone)
	return clone
}

var _ io.Reader = errorReader{}

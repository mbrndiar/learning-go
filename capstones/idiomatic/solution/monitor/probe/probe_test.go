package probe_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/probe"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/fixtures"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/tests/m2"
)

func TestHTTPProberClassifications(t *testing.T) {
	server := httptest.NewServer(fixtures.NewScriptedHandler(
		fixtures.Step{Status: http.StatusNoContent},
		fixtures.Step{Status: http.StatusServiceUnavailable, Body: "down"},
		fixtures.Step{Status: http.StatusOK, Body: "too large"},
	))
	t.Cleanup(server.Close)
	healthProber := probe.NewHTTPProber(server.Client())
	target := validTarget(server.URL)

	healthy := healthProber.Probe(context.Background(), target)
	requireResult(t, healthy, "healthy", "")
	if healthy.HTTPStatus == nil || *healthy.HTTPStatus != http.StatusNoContent ||
		healthy.Message != "status 204 was within 200..399" {
		t.Fatalf("healthy = %+v", healthy)
	}

	degraded := healthProber.Probe(context.Background(), target)
	requireResult(t, degraded, "degraded", "")
	if degraded.HTTPStatus == nil || *degraded.HTTPStatus != http.StatusServiceUnavailable {
		t.Fatalf("degraded = %+v", degraded)
	}

	target.MaxBodyBytes = 3
	oversized := healthProber.Probe(context.Background(), target)
	requireResult(t, oversized, "unhealthy", "body_too_large")
	if oversized.HTTPStatus == nil || *oversized.HTTPStatus != http.StatusOK {
		t.Fatalf("oversized = %+v", oversized)
	}
}

func TestHTTPProberTimeoutCancellationAndTransport(t *testing.T) {
	target := validTarget("http://127.0.0.1/")
	target.TimeoutMS = 5
	timeoutResult := probe.NewHTTPProber(&http.Client{Transport: blockingTransport{}}).
		Probe(context.Background(), target)
	requireResult(t, timeoutResult, "unhealthy", "timeout")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelled := probe.NewHTTPProber(&http.Client{Transport: blockingTransport{}}).Probe(ctx, target)
	requireResult(t, cancelled, "unhealthy", "cancelled")

	transportResult := probe.NewHTTPProber(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("offline transport failure")
		}),
	}).Probe(context.Background(), target)
	requireResult(t, transportResult, "unhealthy", "transport_error")

	clientTimeout := probe.NewHTTPProber(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}),
	}).Probe(context.Background(), target)
	requireResult(t, clientTimeout, "unhealthy", "timeout")

	var attempts atomic.Int32
	noRetry := probe.NewHTTPProber(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			attempts.Add(1)
			return nil, errors.New("fixture permanent failure")
		}),
	}).Probe(context.Background(), target)
	requireResult(t, noRetry, "unhealthy", "transport_error")
	if attempts.Load() != 1 {
		t.Fatalf("transport attempts = %d, want 1", attempts.Load())
	}
}

func TestHTTPProberBodyReadErrorAndClose(t *testing.T) {
	body := &failingBody{}
	target := validTarget("http://127.0.0.1/")
	result := probe.NewHTTPProber(&http.Client{
		Transport: responseTransport(http.StatusAccepted, body),
	}).Probe(context.Background(), target)
	requireResult(t, result, "unhealthy", "body_read_error")
	if result.HTTPStatus == nil || *result.HTTPStatus != http.StatusAccepted || !body.closed.Load() {
		t.Fatalf("result = %+v, closed=%v", result, body.closed.Load())
	}

	body = &failingBody{}
	target.MaxBodyBytes = 0
	result = probe.NewHTTPProber(&http.Client{
		Transport: responseTransport(http.StatusAccepted, body),
	}).Probe(context.Background(), target)
	requireResult(t, result, "healthy", "")
	if !body.closed.Load() {
		t.Fatal("body was not closed when max_body_bytes was zero")
	}
}

func TestHTTPProberDoesNotFollowRedirects(t *testing.T) {
	var destinationRequests atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		destinationRequests.Add(1)
	}))
	t.Cleanup(destination.Close)
	redirect := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, destination.URL, http.StatusFound)
	}))
	t.Cleanup(redirect.Close)

	target := validTarget(redirect.URL)
	target.MaxBodyBytes = 1024
	result := probe.NewHTTPProber(redirect.Client()).Probe(context.Background(), target)
	requireResult(t, result, "healthy", "")
	if result.HTTPStatus == nil || *result.HTTPStatus != http.StatusFound {
		t.Fatalf("result = %+v", result)
	}
	if destinationRequests.Load() != 0 {
		t.Fatalf("redirect destination requests = %d", destinationRequests.Load())
	}
}

func TestHTTPProberClockAndInvalidURL(t *testing.T) {
	start := time.Date(2026, 7, 16, 8, 0, 0, 999_999_999, time.FixedZone("offset", 7200))
	clock := fixedClock{now: start, elapsed: 12*time.Millisecond + 999*time.Microsecond}
	target := validTarget("://bad")
	result := probe.NewHTTPProberWithClock(http.DefaultClient, clock).Probe(context.Background(), target)
	requireResult(t, result, "unhealthy", "transport_error")
	if got := domain.FormatTime(result.CheckedAt); got != "2026-07-16T06:00:00.999Z" {
		t.Fatalf("checked_at = %s", got)
	}
	if result.DurationMS != 12 {
		t.Fatalf("duration_ms = %d", result.DurationMS)
	}

	negativeClock := fixedClock{now: start, elapsed: -time.Second}
	result = probe.NewHTTPProberWithClock(http.DefaultClient, negativeClock).Probe(context.Background(), target)
	if result.DurationMS != 0 {
		t.Fatalf("negative duration_ms = %d", result.DurationMS)
	}
}

func TestClassifyStatus(t *testing.T) {
	for _, test := range []struct {
		status int
		want   domain.Status
	}{
		{status: 199, want: domain.StatusDegraded},
		{status: 200, want: domain.StatusHealthy},
		{status: 399, want: domain.StatusHealthy},
		{status: 400, want: domain.StatusDegraded},
	} {
		status, _ := probe.ClassifyStatus(test.status, 200, 399)
		if status != test.want {
			t.Fatalf("ClassifyStatus(%d) = %q", test.status, status)
		}
	}
}

func validTarget(targetURL string) domain.Target {
	return domain.Target{
		Name: "catalog", URL: targetURL, IntervalMS: 100, TimeoutMS: 50,
		ExpectedStatusMin: 200, ExpectedStatusMax: 399, MaxBodyBytes: 16,
	}
}

func requireResult(t *testing.T, result domain.Observation, status, code string) {
	t.Helper()
	actualCode := ""
	if result.ErrorCode != nil {
		actualCode = string(*result.ErrorCode)
	}
	m2.RequireResult(t, string(result.Status), status, actualCode, code)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

type blockingTransport struct{}

func (blockingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	<-request.Context().Done()
	return nil, request.Context().Err()
}

func responseTransport(status int, body io.ReadCloser) http.RoundTripper {
	return roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       body,
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})
}

type failingBody struct {
	closed atomic.Bool
}

func (*failingBody) Read([]byte) (int, error) {
	return 0, errors.New("fixture body failure")
}

func (body *failingBody) Close() error {
	body.closed.Store(true)
	return nil
}

type fixedClock struct {
	now     time.Time
	elapsed time.Duration
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

func (clock fixedClock) Since(time.Time) time.Duration {
	return clock.elapsed
}

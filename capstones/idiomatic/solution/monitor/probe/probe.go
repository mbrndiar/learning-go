// Package probe defines the target probing capability.
package probe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
)

// Prober performs one observation of a target.
type Prober interface {
	Probe(context.Context, domain.Target) domain.Observation
}

// Clock supplies deterministic wall and elapsed time.
type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type realClock struct{}

func (realClock) Now() time.Time                      { return time.Now() }
func (realClock) Since(start time.Time) time.Duration { return time.Since(start) }

// HTTPProber performs bounded standard-library HTTP probes.
type HTTPProber struct {
	client *http.Client
	clock  Clock
}

// NewHTTPProber constructs an HTTP prober. Redirects are never followed.
func NewHTTPProber(client *http.Client) *HTTPProber {
	return NewHTTPProberWithClock(client, realClock{})
}

// NewHTTPProberWithClock constructs an HTTP prober with deterministic clock seams.
func NewHTTPProberWithClock(client *http.Client, clock Clock) *HTTPProber {
	if client == nil {
		client = http.DefaultClient
	}
	if clock == nil {
		clock = realClock{}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &HTTPProber{client: &clientCopy, clock: clock}
}

// ClassifyStatus classifies a complete bounded HTTP response.
func ClassifyStatus(status, minimum, maximum int) (domain.Status, string) {
	if status >= minimum && status <= maximum {
		return domain.StatusHealthy, fmt.Sprintf("status %d was within %d..%d", status, minimum, maximum)
	}
	return domain.StatusDegraded, fmt.Sprintf("status %d was outside %d..%d", status, minimum, maximum)
}

// Probe observes one target using its configured timeout and body bound.
func (prober *HTTPProber) Probe(ctx context.Context, target domain.Target) (observation domain.Observation) {
	start := prober.clock.Now()
	observation = domain.Observation{
		Target:         target.Name,
		CheckedAt:      start.UTC().Truncate(time.Millisecond),
		Status:         domain.StatusUnhealthy,
		PreviousStatus: domain.StatusUnknown,
	}
	defer func() {
		elapsed := prober.clock.Since(start)
		if elapsed < 0 {
			elapsed = 0
		}
		observation.DurationMS = elapsed.Milliseconds()
	}()

	probeContext, cancel := context.WithTimeout(ctx, time.Duration(target.TimeoutMS)*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(probeContext, http.MethodGet, target.URL, nil)
	if err != nil {
		setFailure(&observation, domain.ErrorTransport, "request failed")
		return observation
	}
	response, err := prober.client.Do(request)
	if err != nil {
		classifyContextFailure(ctx, probeContext, err, &observation, domain.ErrorTransport, "request failed")
		return observation
	}
	defer response.Body.Close()
	observation.HTTPStatus = intPointer(response.StatusCode)

	if target.MaxBodyBytes > 0 {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, target.MaxBodyBytes+1))
		if readErr != nil {
			classifyContextFailure(
				ctx,
				probeContext,
				readErr,
				&observation,
				domain.ErrorBodyRead,
				"response body read failed",
			)
			return observation
		}
		if int64(len(body)) > target.MaxBodyBytes {
			setFailure(
				&observation,
				domain.ErrorBodyTooLarge,
				fmt.Sprintf("response body exceeded %d bytes", target.MaxBodyBytes),
			)
			return observation
		}
	}

	observation.Status, observation.Message = ClassifyStatus(
		response.StatusCode,
		target.ExpectedStatusMin,
		target.ExpectedStatusMax,
	)
	return observation
}

func classifyContextFailure(
	parent context.Context,
	probeContext context.Context,
	cause error,
	observation *domain.Observation,
	fallback domain.ErrorCode,
	fallbackMessage string,
) {
	switch {
	case errors.Is(parent.Err(), context.Canceled):
		setFailure(observation, domain.ErrorCancelled, "probe was cancelled")
	case errors.Is(probeContext.Err(), context.DeadlineExceeded),
		errors.Is(parent.Err(), context.DeadlineExceeded),
		errors.Is(cause, context.DeadlineExceeded):
		setFailure(observation, domain.ErrorTimeout, "probe timed out")
	case errors.Is(probeContext.Err(), context.Canceled),
		errors.Is(cause, context.Canceled):
		setFailure(observation, domain.ErrorCancelled, "probe was cancelled")
	default:
		setFailure(observation, fallback, fallbackMessage)
	}
}

func setFailure(observation *domain.Observation, code domain.ErrorCode, message string) {
	observation.Status = domain.StatusUnhealthy
	observation.ErrorCode = &code
	observation.Message = message
}

func intPointer(value int) *int {
	return &value
}

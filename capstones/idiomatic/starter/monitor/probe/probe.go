// Package probe defines the target probing capability.
package probe

import (
	"context"
	"net/http"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"
)

const placeholderMessage = "TODO: implement HTTP probing"

// Prober performs one observation of a target.
type Prober interface {
	Probe(context.Context, domain.Target) domain.Observation
}

// Clock supplies deterministic wall and elapsed time.
type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

// HTTPProber performs bounded standard-library HTTP probes.
type HTTPProber struct {
	placeholder struct{}
}

// NewHTTPProber constructs an HTTP prober. Redirects are never followed.
func NewHTTPProber(client *http.Client) *HTTPProber {
	return NewHTTPProberWithClock(client, nil)
}

// NewHTTPProberWithClock constructs an HTTP prober with deterministic clock seams.
func NewHTTPProberWithClock(client *http.Client, clock Clock) *HTTPProber {
	return &HTTPProber{}
}

// ClassifyStatus classifies a complete bounded HTTP response.
func ClassifyStatus(status, minimum, maximum int) (domain.Status, string) {
	return domain.StatusUnknown, placeholderMessage
}

// Probe observes one target using its configured timeout and body bound.
func (prober *HTTPProber) Probe(ctx context.Context, target domain.Target) (observation domain.Observation) {
	return domain.Observation{
		Target: target.Name, Status: domain.StatusUnknown, PreviousStatus: domain.StatusUnknown,
		Message: placeholderMessage,
	}
}

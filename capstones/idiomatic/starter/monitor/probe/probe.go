// Package probe defines the target probing capability.
package probe

import (
	"context"
	"net/http"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/starter/monitor/domain"
)

const placeholderMessage = "TODO: implement HTTP probing"

type Prober interface {
	Probe(context.Context, domain.Target) domain.Observation
}

type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type HTTPProber struct {
	placeholder struct{}
}

func NewHTTPProber(client *http.Client) *HTTPProber {
	return NewHTTPProberWithClock(client, nil)
}

func NewHTTPProberWithClock(client *http.Client, clock Clock) *HTTPProber {
	return &HTTPProber{}
}

func ClassifyStatus(status, minimum, maximum int) (domain.Status, string) {
	return domain.StatusUnknown, placeholderMessage
}

func (prober *HTTPProber) Probe(ctx context.Context, target domain.Target) (observation domain.Observation) {
	return domain.Observation{
		Target: target.Name, Status: domain.StatusUnknown, PreviousStatus: domain.StatusUnknown,
		Message: placeholderMessage,
	}
}

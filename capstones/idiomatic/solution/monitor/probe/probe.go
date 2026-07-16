// Package probe defines the target probing capability.
package probe

import (
	"context"
	"net/http"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
)

const placeholderMessage = "TODO: implement HTTP probing"

// Prober performs one observation of a target.
type Prober interface {
	Probe(context.Context, domain.Target) domain.Observation
}

// HTTPProber will implement the standard-library HTTP probe.
type HTTPProber struct{}

// NewHTTPProber constructs the HTTP prober boundary.
func NewHTTPProber(_ *http.Client) *HTTPProber {
	return &HTTPProber{}
}

// Probe returns an explicit non-observation until probing is implemented.
func (*HTTPProber) Probe(_ context.Context, target domain.Target) domain.Observation {
	return domain.Observation{
		Target:         target.Name,
		Status:         domain.StatusUnknown,
		PreviousStatus: domain.StatusUnknown,
		Message:        placeholderMessage,
	}
}

// Package fixtures provides offline monitor contract fixtures.
package fixtures

import (
	"fmt"
	"net/http"
	"sync"
)

// ValidConfig returns a minimal valid configuration for the supplied loopback URL.
func ValidConfig(targetURL string) string {
	return fmt.Sprintf(`{
  "schema_version": 1,
  "max_concurrency": 2,
  "history_limit": 5,
  "targets": [{
    "name": "catalog",
    "url": %q,
    "interval_ms": 100,
    "timeout_ms": 50,
    "expected_status_min": 200,
    "expected_status_max": 399,
    "max_body_bytes": 16
  }]
}`, targetURL)
}

// Step describes one deterministic scripted HTTP response.
type Step struct {
	Status int
	Body   string
	Block  <-chan struct{}
}

// ScriptedHandler serves configured steps in order and repeats the final step.
type ScriptedHandler struct {
	mu    sync.Mutex
	steps []Step
	next  int
}

// NewScriptedHandler constructs a handler with at least one response step.
func NewScriptedHandler(steps ...Step) *ScriptedHandler {
	return &ScriptedHandler{steps: append([]Step(nil), steps...)}
}

// ServeHTTP serves the next scripted response.
func (handler *ScriptedHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler.mu.Lock()
	index := handler.next
	if index < len(handler.steps)-1 {
		handler.next++
	}
	var step Step
	if len(handler.steps) > 0 {
		step = handler.steps[min(index, len(handler.steps)-1)]
	}
	handler.mu.Unlock()
	if step.Block != nil {
		select {
		case <-step.Block:
		case <-request.Context().Done():
			return
		}
	}
	if step.Status == 0 {
		step.Status = http.StatusOK
	}
	writer.WriteHeader(step.Status)
	_, _ = writer.Write([]byte(step.Body))
}

// Package api constructs the monitor's HTTP handler.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
)

const placeholderMessage = "TODO: implement monitor HTTP API"

type handler struct{}

// NewHandler returns an ordinary http.Handler for the monitor API.
func NewHandler(_ history.Store, _ []domain.Target) http.Handler {
	return &handler{}
}

func (*handler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(writer).Encode(domain.ErrorResponse{
		Error: domain.APIError{
			Code:    "not_implemented",
			Message: placeholderMessage,
		},
	})
}

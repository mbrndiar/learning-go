package nethttp_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/server/api/nethttp"
)

func TestStarterReturnsExplicit501(t *testing.T) {
	handler := nethttp.New(nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/tasks", nil))
	if response.Code != http.StatusNotImplemented ||
		response.Body.String() != `{"error":{"code":"not_implemented","message":"this endpoint is not implemented"}}`+"\n" {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}

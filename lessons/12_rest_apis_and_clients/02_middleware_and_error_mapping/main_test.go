package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareAndErrorMapping(t *testing.T) {
	handler := withRequestID(adapt(item))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest || rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("response = %d headers=%v", rec.Code, rec.Header())
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?id=42", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithLoggingLogsMethodAndPath(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := withLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	logged := buf.String()
	if !strings.Contains(logged, "GET") || !strings.Contains(logged, "/widgets") {
		t.Fatalf("log output = %q, want it to mention GET /widgets", logged)
	}
}

func TestWithRecoverConvertsPanicToInternalServerError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := withRecover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusInternalServerError; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
	if !strings.Contains(buf.String(), "boom") {
		t.Fatalf("log output = %q, want it to mention the recovered panic value", buf.String())
	}
}

func TestWithRequestIDIsAvailableToHandler(t *testing.T) {
	t.Parallel()

	nextID := func() string { return "fixed-id" }

	var seen string
	handler := withRequestID(nextID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := r.Context().Value(requestIDKey{}).(string)
		seen = id
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seen != "fixed-id" {
		t.Fatalf("request id seen by handler = %q, want %q", seen, "fixed-id")
	}
}

func TestChainRunsOutermostFirstAndInnermostLast(t *testing.T) {
	t.Parallel()

	var order []string
	record := func(name string) middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+":before")
				next.ServeHTTP(w, r)
				order = append(order, name+":after")
			})
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	handler := chain(record("outer"), record("inner"))(final)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	want := []string{"outer:before", "inner:before", "handler", "inner:after", "outer:after"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

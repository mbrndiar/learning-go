package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSucceeds(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := newClientWithTimeout(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	body, err := fetch(ctx, client, server.URL)
	if err != nil {
		t.Fatalf("fetch() error = %v, want nil", err)
	}
	if body != "ok" {
		t.Fatalf("fetch() body = %q, want %q", body, "ok")
	}
}

func TestFetchReturnsErrorOnNonOKStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newClientWithTimeout(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := fetch(ctx, client, server.URL); err == nil {
		t.Fatal("fetch() error = nil, want an error for a 500 response")
	}
}

func TestFetchRespectsContextDeadline(t *testing.T) {
	t.Parallel()

	released := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-released:
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(released)

	client := newClientWithTimeout(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := fetch(ctx, client, server.URL)
	if err == nil {
		t.Fatal("fetch() error = nil, want a deadline exceeded error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("fetch() error = %v, want it to wrap context.DeadlineExceeded", err)
	}
}

func TestFetchRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	client := newClientWithTimeout(time.Second)
	ctx := context.Background()

	if _, err := fetch(ctx, client, "://not-a-valid-url"); err == nil {
		t.Fatal("fetch() error = nil, want an error building the request")
	}
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestGracefulShutdownWaitsForInFlightRequest(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	inFlight := 0

	server := newServer(":0",
		func() { mu.Lock(); inFlight++; mu.Unlock() },
		func() { mu.Lock(); inFlight--; mu.Unlock() },
	)

	listener, serveErr, err := listenAndServe(server)
	if err != nil {
		t.Fatalf("listenAndServe() error = %v, want nil", err)
	}
	addr := listener.Addr().String()

	// Start a slow request, then shut down almost immediately afterward,
	// while that request is still in flight.
	requestDone := make(chan error, 1)
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/work", addr))
		if err != nil {
			requestDone <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			requestDone <- fmt.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			return
		}
		requestDone <- nil
	}()

	// Give the request a moment to actually reach the handler before we
	// shut down, so this test exercises the "in-flight" path rather than
	// racing the request's own connection setup.
	deadline := time.Now().Add(time.Second)
	for {
		mu.Lock()
		started := inFlight > 0
		mu.Unlock()
		if started {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("request never reached the handler in time")
		}
		time.Sleep(time.Millisecond)
	}

	if err := shutdown(server, 2*time.Second, serveErr); err != nil {
		t.Fatalf("shutdown() error = %v, want nil", err)
	}

	if err := <-requestDone; err != nil {
		t.Fatalf("in-flight request error = %v, want nil (shutdown must let it finish)", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if inFlight != 0 {
		t.Fatalf("inFlight = %d after shutdown, want 0", inFlight)
	}
}

func TestRunStopsWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	server := newServer(":0", func() {}, func() {})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	addr, err := run(ctx, server, time.Second)
	if err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
	if addr == "" {
		t.Fatal("run() addr is empty, want a bound address")
	}
}

// Command 04_http_client_context_timeout shows how to call an HTTP
// endpoint with an explicit context deadline, why a shared http.Client
// should always have a sane default Timeout, and how to release a
// response body deterministically.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

// fetch performs a GET request against url, bounded by ctx. It always
// closes the response body before returning, even on error paths, so the
// underlying connection can be reused (or released) promptly. Failing to
// close a response body is one of the most common resource leaks in Go
// HTTP clients.
func fetch(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close() // always close, even if reading the body fails below

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return string(body), nil
}

// newClientWithTimeout returns an *http.Client with a default Timeout. A
// client with no Timeout set (the zero value) will wait forever for a slow
// or hanging server; always set one, even if you also pass a per-request
// context deadline.
func newClientWithTimeout(d time.Duration) *http.Client {
	return &http.Client{Timeout: d}
}

func main() {
	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello from server")
	}))
	defer fast.Close()

	client := newClientWithTimeout(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	body, err := fetch(ctx, client, fast.URL)
	fmt.Println("fast server response:", body, "error:", err)

	// released is never closed, so this handler blocks until the request's
	// context is canceled: a deterministic way to demonstrate a timeout
	// without depending on wall-clock races between two competing timers.
	released := make(chan struct{})
	blocking := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-released:
		case <-r.Context().Done(): // the server also stops waiting once the client gives up
		}
	}))
	defer blocking.Close()
	defer close(released)

	shortCtx, cancelShort := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelShort()

	_, err = fetch(shortCtx, client, blocking.URL)
	fmt.Println("blocking server error:", err)
	fmt.Println("is deadline exceeded:", errors.Is(err, context.DeadlineExceeded))
}

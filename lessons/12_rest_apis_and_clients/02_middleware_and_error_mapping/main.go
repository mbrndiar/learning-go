// Command 02_middleware_and_error_mapping centralizes domain-error responses
// and composes standard-library middleware.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

var (
	ErrBadInput = errors.New("bad input")
	ErrMissing  = errors.New("missing")
)

type appHandler func(http.ResponseWriter, *http.Request) error

type requestIDKey struct{}

func adapt(next appHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			status := http.StatusInternalServerError
			message := "internal error"
			switch {
			case errors.Is(err, ErrBadInput):
				status, message = http.StatusBadRequest, err.Error()
			case errors.Is(err, ErrMissing):
				status, message = http.StatusNotFound, err.Error()
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
		}
	})
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = "generated-example-id"
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func item(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Query().Get("id") == "" {
		return fmt.Errorf("id is required: %w", ErrBadInput)
	}
	return fmt.Errorf("item not found: %w", ErrMissing)
}

func main() {
	handler := withRequestID(adapt(item))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/?id=42", nil)
	handler.ServeHTTP(recorder, request)
	fmt.Printf("status=%d request_id=%s body=%s", recorder.Code, recorder.Header().Get("X-Request-ID"), recorder.Body.String())
}

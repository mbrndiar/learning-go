// Command 03_middleware_functions shows middleware written as plain
// functions of type func(http.Handler) http.Handler, composed by wrapping
// one handler with another, with no framework involved.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// middleware wraps a handler with additional behavior, and returns a new
// handler. Because it takes and returns the same type, any number of
// middleware can be chained by nesting calls.
type middleware func(http.Handler) http.Handler

// withLogging records the method, path, and duration of every request. It
// calls next.ServeHTTP to run the wrapped handler, then logs afterward, so
// it can measure how long that handler took.
func withLogging(logger *log.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
		})
	}
}

// withRecover converts a panic inside next into a 500 response instead of
// crashing the whole server process. Without this, one bad request that
// triggers a panic (for example a nil map write) would take down every
// other in-flight request too.
func withRecover(logger *log.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Printf("recovered from panic: %v", rec)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// requestIDKey is an unexported type used as a context key, so this
// package's key can never collide with a context key from another package.
type requestIDKey struct{}

// withRequestID stores a generated request ID in the request's context,
// making it available to next and to anything next calls, via
// r.Context().Value(requestIDKey{}).
func withRequestID(nextID func() string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), requestIDKey{}, nextID())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// chain composes middleware in the order given: chain(a, b)(h) applies as
// a(b(h)), so a is the outermost layer and runs first on the way in.
func chain(mw ...middleware) middleware {
	return func(final http.Handler) http.Handler {
		handler := final
		for i := len(mw) - 1; i >= 0; i-- {
			handler = mw[i](handler)
		}
		return handler
	}
}

func main() {
	logger := log.New(io.Discard, "", 0) // discard output in this demo; see the test for assertions on it

	counter := 0
	nextID := func() string {
		counter++
		return fmt.Sprintf("req-%d", counter)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := r.Context().Value(requestIDKey{}).(string)
		fmt.Fprintf(w, "request id: %s", id)
	})

	wrapped := chain(withRecover(logger), withLogging(logger), withRequestID(nextID))(handler)

	_ = wrapped // wired up here; main_test.go exercises it against real requests
	fmt.Println("middleware chain built: recover -> logging -> request id -> handler")
}

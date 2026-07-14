package taskapi

import (
	"log/slog"
	"net/http"
	"time"
)

// WithTimeout returns a handler that derives a context with the given
// timeout from the incoming request's context, replaces the request's
// context with it before calling next, and cancels it once next returns (so
// resources tied to the timeout are released promptly either way). It does
// not itself cut off an in-flight response -- it only shortens the
// deadline seen by next and anything next calls (e.g. a TaskStore method),
// so those must themselves respect ctx.Done()/ctx.Err() to actually stop
// early.
//
// TODO(task 9): implement WithTimeout using context.WithTimeout and
// r.WithContext, remembering to call the returned cancel function (a
// deferred call is fine) so the timer is released once next.ServeHTTP
// returns.
func WithTimeout(next http.Handler, d time.Duration) http.Handler {
	panic("not implemented")
}

// statusRecorder wraps an http.ResponseWriter to capture the status code
// passed to WriteHeader, defaulting to http.StatusOK if the handler never
// calls WriteHeader explicitly (mirroring how net/http itself behaves).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// WithLogging returns a handler that logs one line per request at info
// level once next has finished, including at least the method, path,
// resulting status code, and duration.
//
// TODO(task 10): implement WithLogging. Wrap w in a *statusRecorder
// (initialized with status: http.StatusOK) before calling next.ServeHTTP so
// you can read the final status afterward, then call logger.Info with the
// request's method, URL path, the recorded status, and the elapsed
// time.Since(start).
func WithLogging(next http.Handler, logger *slog.Logger) http.Handler {
	panic("not implemented")
}

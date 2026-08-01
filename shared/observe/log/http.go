package log

import (
	"fmt"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach Flusher/Hijacker on the
// underlying writer (needed for Datastar SSE and similar streaming).
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// HTTPAccess logs one JSON access line per HTTP request (method, path, status, duration_ms).
// Path defaults to r.URL.Path; prefer HTTPAccessWithPath when a low-cardinality
// route pattern is available (e.g. chi placeholders).
func HTTPAccess(next http.Handler) http.Handler {
	return HTTPAccessWithPath(func(r *http.Request) string { return r.URL.Path })(next)
}

// HTTPAccessWithPath is like HTTPAccess but logs pathFn(r) after the handler
// returns (so routers can supply a matched pattern instead of the raw URL).
func HTTPAccessWithPath(pathFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			path := pathFn(r)
			if path == "" {
				path = r.URL.Path
			}
			InfoContext(
				r.Context(),
				fmt.Sprintf("%s %s %d", r.Method, path, rec.status),
				"method", r.Method,
				"path", path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

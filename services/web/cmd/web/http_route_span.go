package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
)

func httpRoutePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if p := rc.RoutePattern(); p != "" {
			return p
		}
	}
	return r.Pattern
}

// httpSpanName is used by otelhttp (including the post-handler rename when
// r.Pattern is set) so Explore shows METHOD + chi route pattern instead of
// the middleware operation string ("web").
func httpSpanName(_ string, r *http.Request) string {
	if p := httpRoutePattern(r); p != "" {
		return r.Method + " " + p
	}
	return r.Method
}

func otelHTTPMiddleware() func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware("web", otelhttp.WithSpanNameFormatter(httpSpanName))
}

// withHTTPRouteSpanName sets http.route after chi matches and re-applies the
// span name in case otelhttp left the operation string.
func withHTTPRouteSpanName(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		pattern := httpRoutePattern(r)
		if pattern == "" {
			return
		}
		span := trace.SpanFromContext(r.Context())
		if !span.IsRecording() {
			return
		}
		span.SetName(r.Method + " " + pattern)
		span.SetAttributes(semconv.HTTPRoute(pattern))
	})
}

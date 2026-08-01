package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

// withHTTPRouteSpanName updates the active otelhttp server span after chi
// matches a route, so Explore shows METHOD + pattern (not the middleware
// operation string) and http.route uses placeholders instead of raw IDs.
func withHTTPRouteSpanName(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		span := trace.SpanFromContext(r.Context())
		if !span.IsRecording() {
			return
		}
		rc := chi.RouteContext(r.Context())
		if rc == nil {
			return
		}
		pattern := rc.RoutePattern()
		if pattern == "" {
			return
		}
		span.SetName(r.Method + " " + pattern)
		span.SetAttributes(semconv.HTTPRoute(pattern))
	})
}

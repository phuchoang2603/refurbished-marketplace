package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
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

// httpSpanName is otelhttp's sole span-naming hook. Chi (Go 1.23+) sets
// r.Pattern during the handler, and otelhttp re-runs this formatter afterward.
func httpSpanName(_ string, r *http.Request) string {
	if p := httpRoutePattern(r); p != "" {
		return r.Method + " " + p
	}
	return r.Method
}

func otelHTTPMiddleware() func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware("web", otelhttp.WithSpanNameFormatter(httpSpanName))
}

// withHTTPRouteAttr sets http.route after chi matches. Span names come only
// from httpSpanName; this does not call SetName.
func withHTTPRouteAttr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		pattern := httpRoutePattern(r)
		if pattern == "" {
			return
		}
		span := trace.SpanFromContext(r.Context())
		if span.IsRecording() {
			span.SetAttributes(semconv.HTTPRoute(pattern))
		}
	})
}

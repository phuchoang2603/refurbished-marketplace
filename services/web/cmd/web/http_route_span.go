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

func httpSpanName(_ string, r *http.Request) string {
	if p := httpRoutePattern(r); p != "" {
		return r.Method + " " + p
	}
	return r.Method
}

func otelHTTPMiddleware() func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware("web", otelhttp.WithSpanNameFormatter(httpSpanName))
}

func withHTTPRouteAttr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		pattern := httpRoutePattern(r)
		if pattern == "" {
			return
		}
		route := semconv.HTTPRoute(pattern)
		if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
			labeler.Add(route)
		}
		span := trace.SpanFromContext(r.Context())
		if span.IsRecording() {
			span.SetAttributes(route)
		}
	})
}

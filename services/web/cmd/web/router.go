package main

import (
	"net/http"
	"time"

	"refurbished-marketplace/services/web/internal/handlers"
	sharedlog "refurbished-marketplace/shared/observe/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func newRouter(h *handlers.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
		otelhttp.NewMiddleware("web"),
		sharedlog.HTTPAccess, // after otelhttp so request context carries the span
	)
	h.Register(router)
	return router
}

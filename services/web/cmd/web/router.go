package main

import (
	"net/http"
	"time"

	"refurbished-marketplace/services/web/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func newRouter(h *handlers.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
	)
	router.Use(otelhttp.NewMiddleware("web"))
	h.Register(router)
	return router
}
